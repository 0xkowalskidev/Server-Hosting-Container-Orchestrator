package runtime

import (
	"context"
	"fmt"
	"log"
	"syscall"
	"time"

	statemanager "0xKowalski1/container-orchestrator/state-manager"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"

	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/containers"
	"github.com/containerd/containerd/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
)

// ContainerdRuntime implements the Runtime interface for containerd.
type ContainerdRuntime struct {
	client *containerd.Client
}

// NewContainerdRuntime creates a new instance of ContainerdRuntime with the given containerd client.
func NewContainerdRuntime(socketPath string) (*ContainerdRuntime, error) {
	client, err := containerd.New(socketPath)

	if err != nil {
		return nil, err
	}

	return &ContainerdRuntime{
		client: client,
	}, nil
}

// CreateContainer instantiates a new container but does not start it.
func (_runtime *ContainerdRuntime) CreateContainer(namespace string, config statemanager.Container) (statemanager.Container, error) {
	ctx := namespaces.WithNamespace(context.Background(), namespace)

	image, err := _runtime.client.Pull(ctx, config.Image, containerd.WithPullUnpack)
	if err != nil {
		log.Printf("Error pulling image: %v", err)
		return statemanager.Container{}, err
	}

	cont, err := _runtime.client.NewContainer(ctx, config.ID, containerd.WithImage(image), containerd.WithNewSnapshot(config.ID+"-snapshot", image), containerd.WithNewSpec(
		oci.WithHostNamespace(specs.NetworkNamespace),
		oci.WithImageConfig(image),
		oci.WithEnv(config.Env),
		func(ctx context.Context, _ oci.Client, _ *containers.Container, s *oci.Spec) error {
			//temporary
			s.Mounts = append(s.Mounts, specs.Mount{
				Destination: "/etc/resolv.conf",
				Type:        "bind",
				Source:      "/home/kowalski/dev/container-orchestrator/resolv.conf",
				Options:     []string{"rbind", "ro"},
			})

			return nil
		},
	))

	if err != nil {
		log.Printf("Error creating container: %v", err)
		return statemanager.Container{}, err
	}

	return statemanager.Container{ID: cont.ID()}, nil
}

// StartContainer starts an existing container.
func (_runtime *ContainerdRuntime) StartContainer(namespace string, containerID string) error {
	ctx := namespaces.WithNamespace(context.Background(), namespace)

	container, err := _runtime.client.LoadContainer(ctx, containerID)
	if err != nil {
		log.Printf("Failed to load container %s: %v", containerID, err)
		return err
	}

	logPath := "/home/kowalski/dev/container-orchestrator/log.log"

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
func (_runtime *ContainerdRuntime) StopContainer(namespace string, containerID string, timeout int) error {
	log.Printf("Attempting to stop container %s in namespace %s with timeout %d", containerID, namespace, timeout)

	ctx := namespaces.WithNamespace(context.Background(), namespace)

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

// RemoveContainer removes a container from the system. This may require the container to be stopped first.
func (_runtime *ContainerdRuntime) RemoveContainer(namespace string, containerID string) error {
	ctx := namespaces.WithNamespace(context.Background(), namespace)

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
func (_runtime *ContainerdRuntime) ListContainers(namespace string) ([]statemanager.Container, error) {
	var containers []statemanager.Container
	ctx := namespaces.WithNamespace(context.Background(), namespace)

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

		containers = append(containers, statemanager.Container{
			ID: cont.ID(),
		})
	}

	return containers, nil
}

// InspectContainer returns detailed information about a specific container.
func (_runtime *ContainerdRuntime) InspectContainer(namespace string, containerID string) (statemanager.Container, error) {
	ctx := namespaces.WithNamespace(context.Background(), namespace)

	container, err := _runtime.client.LoadContainer(ctx, containerID)
	if err != nil {
		log.Printf("Failed to load container %s: %v", containerID, err)
		return statemanager.Container{}, err
	}

	info, err := container.Info(ctx)
	if err != nil {
		log.Printf("Failed to get info for container %s: %v", containerID, err)
		return statemanager.Container{}, err
	}

	// Initialize status as "stopped" assuming that if there's no task, the container is not running.
	containerStatus := "stopped"

	// Attempt to retrieve the task associated with the container
	task, err := container.Task(ctx, nil)
	if err != nil {
		// If an error occurs retrieving the task, log it and proceed with the assumption the container is stopped.
		fmt.Println("No running task found for container, assuming it's stopped:", err)
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
	c := statemanager.Container{
		ID:     info.ID,
		Status: containerStatus,
	}

	return c, nil

}
