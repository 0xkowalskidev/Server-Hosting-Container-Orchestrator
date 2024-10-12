package workernode

import (
	"context"
	"fmt"
	"log"
	"syscall"
	"time"

	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/models"

	"github.com/containerd/typeurl/v2"

	"github.com/containerd/cgroups/v2/stats"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/containerd/errdefs"
	"github.com/opencontainers/runtime-spec/specs-go"
)

type CPUUsageSample struct {
	cpuUsage  uint64
	timestamp time.Time
}

type ContainerProbeState struct {
	// Value is the time these probes last passed.
	readinessProbe time.Time
	livenessProbe  time.Time
}

type ContainerdRuntime struct {
	client                    *containerd.Client
	config                    Config
	previousCPUUsageSampleMap map[string]CPUUsageSample // ContainerID to last cpu sample
	containerProbeStates      map[string]ContainerProbeState
}

func NewContainerdRuntime(config Config) (*ContainerdRuntime, error) {
	client, err := containerd.New(config.ContainerdPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create containerd client: %w", err)
	}
	return &ContainerdRuntime{client: client, config: config, previousCPUUsageSampleMap: make(map[string]CPUUsageSample), containerProbeStates: make(map[string]ContainerProbeState)}, nil
}

func (c *ContainerdRuntime) CreateContainer(ctx context.Context, id string, namespace string, image string, memoryLimit int, cpuLimit int, env []string) (containerd.Container, error) {
	ctx = namespaces.WithNamespace(ctx, namespace)

	imageRef, err := c.client.Pull(ctx, image, containerd.WithPullUnpack)
	if err != nil {
		return nil, fmt.Errorf("failed to pull image %s for container with id %s in namespace %s: %w", image, id, namespace, err)
	}

	volumePath := fmt.Sprintf("%s/%s", c.config.MountsPath, id)
	netnsPath := fmt.Sprintf("/var/run/netns/%s", id)

	preparedEnv := append(env, fmt.Sprintf("MEMORY=%d", memoryLimit))

	specOpts := []oci.SpecOpts{
		oci.WithLinuxNamespace(specs.LinuxNamespace{
			Type: specs.NetworkNamespace,
			Path: netnsPath,
		}),
		oci.WithImageConfig(imageRef),
		oci.WithMounts([]specs.Mount{
			{
				Destination: "/data/server",
				Type:        "linux",
				Source:      volumePath,
				Options:     []string{"rbind", "rw"},
			},
		}),
		oci.WithEnv(preparedEnv),
		oci.WithCPUCFS(100000*int64(cpuLimit), 100000),
		oci.WithMemoryLimit(uint64(memoryLimit) * 1024 * 1024 * 1024), // GB to bytes
	}

	container, err := c.client.NewContainer(
		ctx,
		id,
		containerd.WithNewSnapshot(id+"-snapshot", imageRef),
		containerd.WithNewSpec(specOpts...),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create container with id %s in namespace %s: %w", id, namespace, err)
	}

	return container, nil
}

func (c *ContainerdRuntime) RemoveContainer(ctx context.Context, id string, namespace string) error {
	ctx = namespaces.WithNamespace(ctx, namespace)

	container, err := c.GetContainer(ctx, id, namespace)
	if err != nil {
		return err
	}

	task, err := container.Task(ctx, nil)
	if err == nil {
		if err := task.Kill(ctx, syscall.SIGKILL); err != nil && !errdefs.IsNotFound(err) {
			return fmt.Errorf("failed to kill task of container %s in namespace %s: %w", id, namespace, err)
		}

		statusC, err := task.Wait(ctx)
		if err != nil {
			return fmt.Errorf("failed to wait for task of container %s in namespace %s: %w", id, namespace, err)
		}
		<-statusC

		if _, err := task.Delete(ctx); err != nil && !errdefs.IsNotFound(err) {
			return fmt.Errorf("failed to delete task of container %s in namespace %s: %w", id, namespace, err)
		}
	} else if !errdefs.IsNotFound(err) { // If task does not exist we can just delete the container
		return fmt.Errorf("failed to get task for container %s in namespace %s: %w", id, namespace, err)
	}

	if err := container.Delete(ctx, containerd.WithSnapshotCleanup); err != nil {
		return fmt.Errorf("failed to remove container with id %s in namespace %s: %w", id, namespace, err)
	}

	return nil
}

func (c *ContainerdRuntime) StartContainer(ctx context.Context, id string, namespace string) (containerd.Task, error) {
	ctx = namespaces.WithNamespace(ctx, namespace)

	container, err := c.GetContainer(ctx, id, namespace)
	if err != nil {
		return nil, err
	}

	task, err := container.NewTask(ctx, cio.LogFile(fmt.Sprintf("%s/%s.log", c.config.LogsPath, id)))
	if err != nil {
		return nil, fmt.Errorf("failed to create task for container with id %s in namespace %s: %w", id, namespace, err)
	}

	if err := task.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start task for container with id %s in namespace %s: %w", id, namespace, err)
	}

	return task, nil
}

func (c *ContainerdRuntime) StopContainer(ctx context.Context, id string, namespace string, signal syscall.Signal) error {
	ctx = namespaces.WithNamespace(ctx, namespace)

	container, err := c.GetContainer(ctx, id, namespace)
	if err != nil {
		return err
	}

	task, err := container.Task(ctx, nil)
	if err != nil {
		if errdefs.IsNotFound(err) {
			// Task doesn't exist, nothing to stop
			return nil
		}
		return fmt.Errorf("failed to get task for container with id %s in namespace %s: %w", id, namespace, err)
	}

	if err := task.Kill(ctx, signal); err != nil && !errdefs.IsNotFound(err) {
		return fmt.Errorf("failed to send signal to task for container with id %s in namespace %s: %w", id, namespace, err)
	}

	statusCh, err := task.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for task of container %s in namespace %s: %w", id, namespace, err)
	}

	// Wait for the exit status to be received
	select {
	case <-statusCh:
	case <-ctx.Done():
		return fmt.Errorf("context cancelled while waiting for task of container %s in namespace %s to exit", id, namespace)
	}

	if _, err := task.Delete(ctx); err != nil && !errdefs.IsNotFound(err) {
		return fmt.Errorf("failed to delete task of container %s in namespace %s: %w", id, namespace, err)
	}

	return nil
}

func (c *ContainerdRuntime) GetContainer(ctx context.Context, id string, namespace string) (containerd.Container, error) {
	ctx = namespaces.WithNamespace(ctx, namespace)

	container, err := c.client.LoadContainer(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load container with id %s in namespace %s: %w", id, namespace, err)
	}

	return container, nil
}

func (c *ContainerdRuntime) GetContainers(ctx context.Context, namespace string) ([]containerd.Container, error) {
	ctx = namespaces.WithNamespace(ctx, namespace)

	containers, err := c.client.Containers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load containers in namespace %s: %w", namespace, err)
	}

	return containers, nil
}

func (c *ContainerdRuntime) ExecContainer(ctx context.Context, id string, namespace string, execID string, scriptPath string, arg string) (int, error) {
	ctx = namespaces.WithNamespace(ctx, namespace)

	container, err := c.client.LoadContainer(ctx, id)
	if err != nil {
		return -1, fmt.Errorf("failed to load container with ID %s: %w", id, err)
	}

	task, err := container.Task(ctx, nil)
	if err != nil {
		return -1, fmt.Errorf("failed to retrieve task for container %s: %w", id, err)
	}

	processSpec := &specs.Process{
		Args: []string{scriptPath, arg},
		Cwd:  "/",
	}

	process, err := task.Exec(ctx, execID, processSpec, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return -1, fmt.Errorf("failed to exec process in container %s: %w", id, err)
	}

	defer func() {
		_, err := process.Delete(ctx)
		if err != nil {
			log.Printf("warning: failed to delete process %s in container %s: %v", execID, id, err)
		}
	}()

	if err := process.Start(ctx); err != nil {
		return -1, fmt.Errorf("failed to start process in container %s: %w", id, err)
	}

	exitStatusC, err := process.Wait(ctx)
	if err != nil {
		return -1, fmt.Errorf("failed to wait for process in container %s: %w", id, err)
	}

	status := <-exitStatusC
	exitCode, _, err := status.Result()
	if err != nil {
		return -1, fmt.Errorf("failed to retrieve exit code for process in container %s: %w", id, err)
	}

	return int(exitCode), nil
}

func (c *ContainerdRuntime) GetContainerStatus(ctx context.Context, id string, namespace string) (models.ContainerStatus, error) {
	ctx = namespaces.WithNamespace(ctx, namespace)

	container, err := c.GetContainer(ctx, id, namespace)
	if err != nil {
		return models.StatusUnknown, err
	}

	task, err := container.Task(ctx, nil)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return models.StatusStopped, nil
		} else {
			return models.StatusUnknown, fmt.Errorf("Failed to get task: %v", err)
		}
	}

	status, err := task.Status(ctx)
	if err != nil {
		return models.StatusUnknown, fmt.Errorf("Failed to get task status: %v", err)
	}
	switch status.Status {
	case containerd.Running:
		containerProbeState := c.containerProbeStates[id]

		if containerProbeState.readinessProbe.IsZero() {
			exitCode, err := c.ExecContainer(ctx, id, namespace, id+"-readiness-probe", "/data/scripts/readiness_probe.sh", "")
			if err != nil {
				log.Println(err) // TODO do something else here
				return models.StatusRunning, nil
			}

			if exitCode == 0 {
				containerProbeState.readinessProbe = time.Now()
			}
		} else if containerProbeState.livenessProbe.IsZero() || time.Since(containerProbeState.livenessProbe) >= 10*time.Second { // TODO Probably want to take interval from container config
			exitCode, err := c.ExecContainer(ctx, id, namespace, id+"-liveness-probe", "/data/scripts/liveness_probe.sh", "")
			if err != nil {
				log.Println(err) // TODO do something else here
				return models.StatusRunning, nil
			}

			if exitCode == 0 {
				containerProbeState.livenessProbe = time.Now()
			}
		}

		c.containerProbeStates[id] = containerProbeState

		if !c.containerProbeStates[id].readinessProbe.IsZero() && !c.containerProbeStates[id].livenessProbe.IsZero() {
			return models.StatusReady, nil
		} else {
			return models.StatusRunning, nil
		}
	case containerd.Stopped:
		return models.StatusStopped, nil
	default:
		return models.StatusUnknown, fmt.Errorf("Unhandled task status type: %s", status.Status)
	}
}

func (c *ContainerdRuntime) GetContainerMetrics(ctx context.Context, id string, namespace string) (models.Metrics, error) {
	var taskMetrics models.Metrics
	ctx = namespaces.WithNamespace(ctx, namespace)

	container, err := c.GetContainer(ctx, id, namespace)
	if err != nil {
		return taskMetrics, err
	}

	// Needed to get container core count
	spec, err := container.Spec(ctx)
	if err != nil {
		return taskMetrics, fmt.Errorf("Failed to get container spec: %v", err)
	}

	task, err := container.Task(ctx, nil)
	if err != nil {
		return taskMetrics, fmt.Errorf("Failed to get task: %v", err)
	}

	rawMetrics, err := task.Metrics(ctx)
	if err != nil {
		return taskMetrics, fmt.Errorf("Failed to get metrics: %v", err)
	}

	metrics, err := typeurl.UnmarshalAny(rawMetrics.Data)
	if err != nil {
		return taskMetrics, fmt.Errorf("Failed to unmarshal metrics: %v", err)
	}

	switch metric := metrics.(type) {
	case *stats.Metrics:
		// Memory
		taskMetrics.MemoryUsage = float64(metric.Memory.Usage) / (1024 * 1024 * 1024)                // Convert to GB
		taskMetrics.MemoryLimit = float64(*spec.Linux.Resources.Memory.Limit) / (1024 * 1024 * 1024) // Convert to GB, this will cause a crash if memory limit is not set. Which is fine, as imo memory limit not being set is crash worthy. TODO make this an fatal assert?

		//CPU
		newCPU := metric.CPU.UsageUsec
		newTimestamp := time.Now()

		previousSample, exists := c.previousCPUUsageSampleMap[id]
		if exists {
			intervalSeconds := newTimestamp.Sub(previousSample.timestamp).Seconds()
			usageDelta := float64(newCPU - previousSample.cpuUsage)

			quota := spec.Linux.Resources.CPU.Quota
			period := spec.Linux.Resources.CPU.Period
			coreCount := 1.0 // Default to 1 core if quota/period are not set TODO fatal assert?

			if quota != nil && period != nil && *period > 0 {
				coreCount = float64(*quota) / float64(*period)
			}

			taskMetrics.CPUUsage = (usageDelta / intervalSeconds) / 1e6 / coreCount * 100 // Convert to percentage
		} else {
			taskMetrics.CPUUsage = 0 // Initial value; no previous sample to compare
		}

		c.previousCPUUsageSampleMap[id] = CPUUsageSample{cpuUsage: newCPU, timestamp: newTimestamp}
	default:
		return taskMetrics, fmt.Errorf("Unknown metrics type %T", metrics)
	}

	return taskMetrics, nil
}
