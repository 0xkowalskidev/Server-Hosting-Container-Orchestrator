package workernode

import (
	"context"
	"fmt"
	"log"
	"syscall"
	"time"

	"0xKowalski1/container-orchestrator/api-wrapper"
	"0xKowalski1/container-orchestrator/config"
	"0xKowalski1/container-orchestrator/models"

	"github.com/containerd/containerd"
	eventstypes "github.com/containerd/containerd/api/events"
	"github.com/containerd/containerd/events"
	"github.com/containerd/containerd/namespaces"

	"github.com/containerd/typeurl/v2"

	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
)

// ContainerdRuntime implements the Runtime interface for containerd.
type ContainerdRuntime struct {
	client *containerd.Client
	cfg    *config.Config
}

// NewContainerdRuntime creates a new instance of ContainerdRuntime with the given containerd client.
func NewContainerdRuntime(cfg *config.Config) (*ContainerdRuntime, error) {
	client, err := containerd.New(cfg.ContainerdSocketPath)

	if err != nil {
		return nil, err
	}

	runtime := &ContainerdRuntime{
		client: client,
		cfg:    cfg,
	}

	runtime.SubscribeToEvents()

	return runtime, nil
}

func (r *ContainerdRuntime) SyncContainers(node *models.Node) error {
	desiredContainers := node.Containers

	// List actual containers
	actualContainers, err := r.ListContainers()
	if err != nil {
		log.Printf("Error listing containers: %v", err)
		return err
	}

	// Map actual container IDs for easier lookup
	actualMap := make(map[string]models.Container)
	for _, c := range actualContainers {
		ic, err := r.InspectContainer(c.ID)
		if err != nil {
			log.Printf("Error inspecting container: %v", err)
		}
		actualMap[c.ID] = ic
	}

	for _, desiredContainer := range desiredContainers {
		// Create missing containers
		if _, exists := actualMap[desiredContainer.ID]; !exists {
			// Create container if it does not exist in actual state

			_, err := r.CreateContainer(desiredContainer)
			if err != nil {
				log.Printf("Failed to create container: %v", err)
				continue
			}
		}

		r.reconcileContainerState(desiredContainer, actualMap[desiredContainer.ID])
	}

	// Stop extra containers
	for _, c := range actualContainers {
		found := false
		for _, d := range desiredContainers {
			if d.ID == c.ID {
				found = true
				break
			}
		}
		if !found {
			// Check errors
			r.StopContainer(c.ID, c.StopTimeout)
			r.RemoveContainer(c.ID)
		}
	}

	return nil
}

func (r *ContainerdRuntime) reconcileContainerState(desiredContainer models.Container, actualContainer models.Container) {
	switch desiredContainer.DesiredStatus {
	case "running":
		if actualContainer.Status != "running" {

			err := r.StartContainer(desiredContainer.ID)
			if err != nil {
				log.Fatalf("Failed to start container: %v", err)
			}

		}
	case "stopped":
		if actualContainer.Status != "stopped" {
			err := r.StopContainer(desiredContainer.ID, desiredContainer.StopTimeout)
			if err != nil {
				log.Fatalf("Failed to stop container: %v", err)
			}
		}
	}
}

// This should return a pointer to a container
// CreateContainer instantiates a new container but does not start it.
func (_runtime *ContainerdRuntime) CreateContainer(containerSpec models.Container) (models.Container, error) {
	ctx := namespaces.WithNamespace(context.Background(), _runtime.cfg.Namespace) // put this in the _runtime struct

	image, err := _runtime.client.Pull(ctx, containerSpec.Image, containerd.WithPullUnpack)
	if err != nil {
		log.Printf("Error pulling image: %v", err)
		return models.Container{}, err
	}

	volumePath := _runtime.cfg.StoragePath + containerSpec.ID

	mounts := []oci.SpecOpts{
		oci.WithMounts([]specs.Mount{
			{
				Destination: "/data/server",
				Type:        "linux",
				Source:      volumePath,
				Options:     []string{"rbind", "rw"},
			},
		}),
	}

	specOpts := []oci.SpecOpts{
		oci.WithLinuxNamespace(specs.LinuxNamespace{
			Type: "network",
			Path: _runtime.cfg.NetworkNamespacePath + containerSpec.ID,
		}),
		oci.WithImageConfig(image),
		oci.WithEnv(containerSpec.Env), // Should apply memory to env
		oci.WithMemoryLimit(uint64(containerSpec.MemoryLimit * 1024 * 1024 * 1024)), // in bytes
		oci.WithCPUs(fmt.Sprint(containerSpec.CpuLimit)),
	}

	specOpts = append(specOpts, mounts...)

	cont, err := _runtime.client.NewContainer(ctx, containerSpec.ID, containerd.WithImage(image), containerd.WithNewSnapshot(containerSpec.ID+"-snapshot", image), containerd.WithNewSpec(specOpts...))

	if err != nil {
		log.Printf("Error creating container: %v", err)
		return models.Container{}, err
	}

	if err != nil {
		fmt.Println("Failed to setup container network:", err)
		return models.Container{}, err
	}

	fmt.Printf("Container network setup successful")

	return models.Container{ID: cont.ID()}, nil
}

// StartContainer starts an existing container.
func (_runtime *ContainerdRuntime) StartContainer(containerID string) error {
	ctx := namespaces.WithNamespace(context.Background(), _runtime.cfg.Namespace)

	container, err := _runtime.client.LoadContainer(ctx, containerID)
	if err != nil {
		log.Printf("Failed to load container %s: %v", containerID, err)
		return err
	}

	logPath := _runtime.cfg.LogPath + _runtime.cfg.Namespace + "-" + containerID + ".log"

	task, err := container.NewTask(ctx, cio.LogFile(logPath))
	if err != nil {
		log.Printf("Failed to create task for container %s: %v", containerID, err)
		return err
	}
	defer task.Delete(ctx)

	if err := task.Start(ctx); err != nil {
		log.Printf("Failed to start task for container %s: %v", containerID, err)

		return err
	}

	log.Printf("Successfully started container %s", containerID)

	return nil
}

// StopContainer stops a running container.
func (_runtime *ContainerdRuntime) StopContainer(containerID string, timeout int) error {
	log.Printf("Attempting to stop container %s with timeout %d", containerID, timeout)

	ctx := namespaces.WithNamespace(context.Background(), _runtime.cfg.Namespace)

	log.Printf("Loading container %s", containerID)
	container, err := _runtime.client.LoadContainer(ctx, containerID)
	if err != nil {
		log.Printf("Failed to load container %s: %v", containerID, err)
		return err
	}

	log.Printf("Loading task for container %s", containerID)
	task, err := container.Task(ctx, cio.Load)
	if err != nil {
		log.Printf("Failed to load task for container %s: %v", containerID, err)
		return err
	}

	log.Printf("Sending SIGTERM to container %s", containerID)
	if err := task.Kill(ctx, syscall.SIGTERM); err != nil {
		log.Printf("Failed to send SIGTERM to container %s: %v", containerID, err)
		return err
	}

	log.Printf("Waiting for container %s to exit", containerID)
	exitCh, err := task.Wait(ctx)
	if err != nil {
		log.Printf("Failed to wait on task for container %s: %v", containerID, err)
		return fmt.Errorf("failed to wait on task for container %s: %v", containerID, err)
	}

	select {
	case <-exitCh:
		log.Printf("Container %s stopped gracefully", containerID)
	case <-time.After(time.Duration(timeout) * time.Second):
		log.Printf("Timeout reached; sending SIGKILL to container %s", containerID)
		if err := task.Kill(ctx, syscall.SIGKILL); err != nil { // Forcefully stop the container
			log.Printf("Failed to send SIGKILL to container %s: %v", containerID, err)
			return err
		}
		log.Printf("Waiting for SIGKILL to take effect for container %s", containerID)
		<-exitCh // Wait for the SIGKILL to take effect
	}

	log.Printf("Deleting task for container %s", containerID)
	if _, err := task.Delete(ctx); err != nil {
		log.Printf("Failed to delete task for container %s: %v", containerID, err)
		return err
	}

	log.Printf("Successfully stopped and deleted container %s", containerID)

	return nil
}

// RemoveContainer removes a container from the system. This requires the container to be stopped first.
func (_runtime *ContainerdRuntime) RemoveContainer(containerID string) error {
	ctx := namespaces.WithNamespace(context.Background(), _runtime.cfg.Namespace)

	container, err := _runtime.client.LoadContainer(ctx, containerID)
	if err != nil {
		log.Printf("Failed to load container %s: %v", containerID, err)
		return err
	}

	// Attempt to delete the container. If the container has a running task, containerd will return an error.
	if err := container.Delete(ctx, containerd.WithSnapshotCleanup); err != nil {
		log.Printf("Failed to delete container %s: %v", containerID, err)
		return err
	}

	log.Printf("Successfully deleted container %s", containerID)

	return nil
}

// ListContainers returns a list of all containers managed by the runtime.
func (_runtime *ContainerdRuntime) ListContainers() ([]models.Container, error) {
	var containers []models.Container
	ctx := namespaces.WithNamespace(context.Background(), _runtime.cfg.Namespace)

	// List containers from containerd
	conts, err := _runtime.client.Containers(ctx)
	if err != nil {
		log.Printf("Error listing containers: %v", err)
		return nil, err
	}

	// Map containerd containers to generic Container struct
	for _, cont := range conts {
		/*info*/ _, err := cont.Info(ctx)
		if err != nil {
			log.Printf("Error getting container info: %v", err)
			continue
		}

		containers = append(containers, models.Container{
			ID: cont.ID(),
		})
	}

	return containers, nil
}

// InspectContainer returns detailed information about a specific container.
func (_runtime *ContainerdRuntime) InspectContainer(containerID string) (models.Container, error) {
	ctx := namespaces.WithNamespace(context.Background(), _runtime.cfg.Namespace)

	container, err := _runtime.client.LoadContainer(ctx, containerID)
	if err != nil {
		log.Printf("Failed to load container %s: %v", containerID, err)
		return models.Container{}, err
	}

	info, err := container.Info(ctx)
	if err != nil {
		log.Printf("Failed to get info for container %s: %v", containerID, err)
		return models.Container{}, err
	}

	// Initialize status as "stopped" assuming that if there's no task, the container is not running.
	containerStatus := "stopped"

	// Attempt to retrieve the task associated with the container
	task, err := container.Task(ctx, nil)
	if err != nil {
		// No task, probably fine
	} else {
		// If a task is found, retrieve its status
		status, err := task.Status(ctx)
		if err != nil {
			fmt.Println("Error retrieving task status:", err)
		} else {
			containerStatus = string(status.Status)
		}
	}

	// Construct your container representation including its status
	c := models.Container{
		ID:     info.ID,
		Status: containerStatus,
	}

	return c, nil

}

// Events
// SubscribeToEvents starts listening to containerd events and handles them.
func (_runtime *ContainerdRuntime) SubscribeToEvents() {
	ctx := namespaces.WithNamespace(context.Background(), _runtime.cfg.Namespace)

	// Process events and errors
	go func() {
		ch, errs := _runtime.client.EventService().Subscribe(ctx)

		for {
			select {
			case envelope := <-ch:
				if err := _runtime.processEvent(envelope, _runtime.cfg.Namespace); err != nil {
					log.Printf("Error processing event: %v", err)
				}

			case e := <-errs:
				log.Printf("Received an error: %v", e)
			}
		}
	}()
}

func (_runtime *ContainerdRuntime) processEvent(envelope *events.Envelope, namespace string) error {
	// MAKE SURE NAMESPACE IS CORRECT HERE

	apiClient := api.NewApiWrapper(_runtime.cfg.ControlNodeIp)
	//Should probably check namespace here
	event, err := typeurl.UnmarshalAny(envelope.Event)
	if err != nil {
		return err
	}

	switch e := event.(type) {
	case *eventstypes.TaskStart:
		log.Printf("Task started: ContainerID=%s, PID=%d", e.ContainerID, e.Pid)
		status := "running"
		containerPatch := models.UpdateContainerRequest{Status: &status}
		_, err := apiClient.UpdateContainer(e.ContainerID, containerPatch)
		if err != nil {
			log.Printf("Error updating container %s to status 'running': %v", e.ContainerID, err)
		}

	case *eventstypes.TaskDelete:
		log.Printf("Task deleted: ContainerID=%s, PID=%d, ExitStatus=%d", e.ContainerID, e.Pid, e.ExitStatus)
		status := "stopped"
		containerPatch := models.UpdateContainerRequest{Status: &status}
		_, err := apiClient.UpdateContainer(e.ContainerID, containerPatch)
		if err != nil {
			log.Printf("Error updating container %s to status 'stopped': %v", e.ContainerID, err)
		}

	/* case *eventstypes.TaskExit:
	log.Printf("Task exit: ContainerID=%s, PID=%d, ExitStatus=%d", e.ContainerID, e.Pid, e.ExitStatus) */ // Dont think we need this one

	default:
		log.Printf("Unhandled event type: %s", envelope.Topic)
	}

	return nil
}
