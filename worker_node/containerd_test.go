package workernode_test

import (
	"context"
	"syscall"
	"testing"

	"github.com/0xKowalski1/Server-Hosting-Container-Orchestrator/config"
	workernode "github.com/0xKowalski1/Server-Hosting-Container-Orchestrator/worker_node"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/errdefs"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) (workernode.Runtime, config.Config) {
	var cfg config.Config
	config.ParseConfigFromEnv(&cfg)

	runtime, err := workernode.NewRuntime(cfg)
	require.NoError(t, err)
	require.NotNil(t, runtime)

	teardown(t, cfg)

	return runtime, cfg
}

func teardown(t *testing.T, cfg config.Config) {
	ctx := namespaces.WithNamespace(context.Background(), cfg.NamespaceMain)

	client, err := containerd.New(cfg.ContainerdPath)
	require.NoError(t, err)

	containers, err := client.Containers(ctx)
	if err != nil && !errdefs.IsNotFound(err) {
		t.Fatalf("Failed to list containers: %v", err)
	}

	// Iterate through each container and clean up
	for _, container := range containers {
		task, err := container.Task(ctx, nil)
		if err == nil {
			// Check if the task is running before attempting to kill it
			status, err := task.Status(ctx)
			if err != nil {
				t.Logf("Failed to get status for task in container %s: %v", container.ID(), err)
				continue
			}

			// Only kill the task if it is in the Running state
			if status.Status == containerd.Running {
				// Kill the task
				if err := task.Kill(ctx, syscall.SIGKILL); err != nil {
					t.Logf("Failed to kill task for container %s: %v", container.ID(), err)
					continue
				}

				// Wait for the task to exit
				statusCh, err := task.Wait(ctx)
				if err != nil {
					t.Logf("Failed to wait for task to exit for container %s: %v", container.ID(), err)
					continue
				}

				// Wait on the exit status
				exitStatus := <-statusCh
				if err := exitStatus.Error(); err != nil {
					t.Logf("Task for container %s exited with error: %v", container.ID(), err)
				}
			}

			// Delete the task
			if _, err := task.Delete(ctx); err != nil {
				t.Logf("Failed to delete task for container %s: %v", container.ID(), err)
			}
		} else if !errdefs.IsNotFound(err) {
			t.Logf("Failed to retrieve task for container %s: %v", container.ID(), err)
			continue
		}

		// Attempt to delete the container
		if err := container.Delete(ctx, containerd.WithSnapshotCleanup); err != nil {
			t.Logf("Failed to delete container %s: %v", container.ID(), err)
		}
	}

	t.Logf("Successfully cleaned up containerd namespace: %s", cfg.NamespaceMain)
}

// CreateContainer
func TestCreateContainer(t *testing.T) {
	runtime, cfg := setup(t)
	defer teardown(t, cfg)

	ctx := namespaces.WithNamespace(context.Background(), cfg.NamespaceMain)

	containerID := "test-create-container"
	image := "docker.io/library/alpine:latest"

	createdContainer, err := runtime.CreateContainer(ctx, containerID, cfg.NamespaceMain, image)
	require.NoError(t, err)
	require.NotNil(t, createdContainer)
}

// RemoveContainer
func TestRemoveContainer(t *testing.T) {
	runtime, cfg := setup(t)
	defer teardown(t, cfg)

	ctx := namespaces.WithNamespace(context.Background(), cfg.NamespaceMain)

	containerID := "test-remove-container"
	image := "docker.io/library/alpine:latest"

	createdContainer, err := runtime.CreateContainer(ctx, containerID, cfg.NamespaceMain, image)
	require.NoError(t, err)
	require.NotNil(t, createdContainer)

	err = createdContainer.Delete(ctx, containerd.WithSnapshotCleanup)
	require.NoError(t, err)

	_, err = runtime.GetContainer(ctx, containerID, cfg.NamespaceMain)
	require.Error(t, err)
	require.True(t, errdefs.IsNotFound(err))
}

// StartContainer
func TestStartContainer(t *testing.T) {
	runtime, cfg := setup(t)
	defer teardown(t, cfg)

	ctx := namespaces.WithNamespace(context.Background(), cfg.NamespaceMain)

	containerID := "test-start-container"
	image := "docker.io/library/alpine:latest"

	createdContainer, err := runtime.CreateContainer(ctx, containerID, cfg.NamespaceMain, image)
	require.NoError(t, err)
	require.NotNil(t, createdContainer)

	task, err := runtime.StartContainer(ctx, containerID, cfg.NamespaceMain)
	require.NoError(t, err)
	require.NotNil(t, task)

	status, err := task.Status(ctx)
	require.NoError(t, err)
	require.Equal(t, containerd.Running, status.Status)
}

// StopContainer
func TestStopContainer(t *testing.T) {
	runtime, cfg := setup(t)
	defer teardown(t, cfg)

	ctx := namespaces.WithNamespace(context.Background(), cfg.NamespaceMain)

	containerID := "test-stop-container"
	image := "docker.io/library/alpine:latest"

	createdContainer, err := runtime.CreateContainer(ctx, containerID, cfg.NamespaceMain, image)
	require.NoError(t, err)
	require.NotNil(t, createdContainer)

	task, err := runtime.StartContainer(ctx, containerID, cfg.NamespaceMain)
	require.NoError(t, err)
	require.NotNil(t, task)

	status, err := task.Status(ctx)
	require.NoError(t, err)
	require.Equal(t, containerd.Running, status.Status)

	exitCh, err := runtime.StopContainer(ctx, containerID, cfg.NamespaceMain)
	require.NoError(t, err)

	// Wait for the exit status
	exitStatus := <-exitCh
	require.NoError(t, exitStatus.Error())

	status, err = task.Status(ctx)
	require.NoError(t, err)
	require.Equal(t, containerd.Stopped, status.Status)
}

// GetContainer
func TestGetContainer(t *testing.T) {
	runtime, cfg := setup(t)
	defer teardown(t, cfg)

	ctx := namespaces.WithNamespace(context.Background(), cfg.NamespaceMain)

	containerID := "test-get-container"
	image := "docker.io/library/alpine:latest"

	createdContainer, err := runtime.CreateContainer(ctx, containerID, cfg.NamespaceMain, image)
	require.NoError(t, err)
	require.NotNil(t, createdContainer)

	retrievedContainer, err := runtime.GetContainer(ctx, containerID, cfg.NamespaceMain)
	require.NoError(t, err)
	require.NotNil(t, retrievedContainer)

	require.Equal(t, createdContainer.ID(), retrievedContainer.ID(), "The container IDs should match")
}

// GetContainers
func TestGetContainers(t *testing.T) {
	runtime, cfg := setup(t)
	defer teardown(t, cfg)

	ctx := namespaces.WithNamespace(context.Background(), cfg.NamespaceMain)

	containerIDs := []string{"test-get-containers-1", "test-get-containers-2", "test-get-containers-3"}
	image := "docker.io/library/alpine:latest"

	for _, containerID := range containerIDs {
		createdContainer, err := runtime.CreateContainer(ctx, containerID, cfg.NamespaceMain, image)
		require.NoError(t, err)
		require.NotNil(t, createdContainer)
	}

	containers, err := runtime.GetContainers(ctx, cfg.NamespaceMain)
	require.NoError(t, err)
	require.NotNil(t, containers)
	require.Len(t, containers, len(containerIDs), "The number of containers retrieved should match the number created")

	retrievedIDs := make(map[string]bool)
	for _, container := range containers {
		retrievedIDs[container.ID()] = true
	}

	for _, containerID := range containerIDs {
		require.True(t, retrievedIDs[containerID], "Container %s should be present in the retrieved list", containerID)
	}
}
