package workernode_test

import (
	"context"
	"syscall"
	"testing"

	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/utils"
	workernode "github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/worker_node"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/errdefs"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) (*workernode.ContainerdRuntime, workernode.Config) {
	var cfg workernode.Config
	utils.ParseConfigFromEnv(&cfg)

	runtime, err := workernode.NewRuntime(cfg)
	require.NoError(t, err)
	require.NotNil(t, runtime)

	teardown(t, cfg)

	return runtime, cfg
}

func teardown(t *testing.T, cfg workernode.Config) {
	ctx := namespaces.WithNamespace(context.Background(), cfg.NamespaceMain)

	client, err := containerd.New(cfg.ContainerdPath)
	require.NoError(t, err)

	containers, err := client.Containers(ctx)
	if err != nil && !errdefs.IsNotFound(err) {
		t.Fatalf("Failed to list containers: %v", err)
	}

	for _, container := range containers {
		task, err := container.Task(ctx, nil)
		if err == nil {
			// Check if the task is running before attempting to kill it
			status, err := task.Status(ctx)
			if err != nil {
				t.Logf("Failed to get status for task in container %s: %v", container.ID(), err)
				continue
			}

			if status.Status == containerd.Running {
				if err := task.Kill(ctx, syscall.SIGKILL); err != nil {
					t.Logf("Failed to kill task for container %s: %v", container.ID(), err)
					continue
				}

				statusCh, err := task.Wait(ctx)
				if err != nil {
					t.Logf("Failed to wait for task to exit for container %s: %v", container.ID(), err)
					continue
				}

				exitStatus := <-statusCh
				if err := exitStatus.Error(); err != nil {
					t.Logf("Task for container %s exited with error: %v", container.ID(), err)
				}
			}

			if _, err := task.Delete(ctx); err != nil {
				t.Logf("Failed to delete task for container %s: %v", container.ID(), err)
			}
		} else if !errdefs.IsNotFound(err) {
			t.Logf("Failed to retrieve task for container %s: %v", container.ID(), err)
			continue
		}

		if err := container.Delete(ctx, containerd.WithSnapshotCleanup); err != nil {
			t.Logf("Failed to delete container %s: %v", container.ID(), err)
		}
	}

	t.Logf("Successfully cleaned up containerd namespace: %s", cfg.NamespaceMain)
}

// CreateContainer
func TestCreateContainer_Valid(t *testing.T) {
	runtime, cfg := setup(t)
	defer teardown(t, cfg)

	ctx := namespaces.WithNamespace(context.Background(), cfg.NamespaceMain)

	containerID := "test-create-container"
	image := "docker.io/library/alpine:latest"

	createdContainer, err := runtime.CreateContainer(ctx, containerID, cfg.NamespaceMain, image)
	require.NoError(t, err)
	require.NotNil(t, createdContainer)
}

func TestCreateContainer_EmptyContainerID(t *testing.T) {
	runtime, cfg := setup(t)
	defer teardown(t, cfg)

	ctx := namespaces.WithNamespace(context.Background(), cfg.NamespaceMain)

	emptyContainerID := ""
	image := "docker.io/library/alpine:latest"

	_, err := runtime.CreateContainer(ctx, emptyContainerID, cfg.NamespaceMain, image)
	require.Error(t, err, "Creating a container with an empty ID should return an error")
}

func TestCreateContainer_EmptyImageName(t *testing.T) {
	runtime, cfg := setup(t)
	defer teardown(t, cfg)

	ctx := namespaces.WithNamespace(context.Background(), cfg.NamespaceMain)

	containerID := "test-create-container-empty-image"
	emptyImage := ""

	_, err := runtime.CreateContainer(ctx, containerID, cfg.NamespaceMain, emptyImage)
	require.Error(t, err, "Creating a container with an empty image name should return an error")
}

func TestCreateContainer_InvalidImageName(t *testing.T) {
	runtime, cfg := setup(t)
	defer teardown(t, cfg)

	ctx := namespaces.WithNamespace(context.Background(), cfg.NamespaceMain)

	containerID := "test-create-container-invalid-image"
	invalidImage := "invalid-image-name"

	_, err := runtime.CreateContainer(ctx, containerID, cfg.NamespaceMain, invalidImage)
	require.Error(t, err, "Creating a container with an invalid image name should return an error")
}

func TestCreateContainer_DuplicateContainerID(t *testing.T) {
	runtime, cfg := setup(t)
	defer teardown(t, cfg)

	ctx := namespaces.WithNamespace(context.Background(), cfg.NamespaceMain)

	containerID := "test-create-container-duplicate"
	image := "docker.io/library/alpine:latest"

	// Create the first container
	_, err := runtime.CreateContainer(ctx, containerID, cfg.NamespaceMain, image)
	require.NoError(t, err)

	// Attempt to create another container with the same ID
	_, err = runtime.CreateContainer(ctx, containerID, cfg.NamespaceMain, image)
	require.Error(t, err, "Creating a container with a duplicate ID should return an error")
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

	err = runtime.RemoveContainer(ctx, containerID, cfg.NamespaceMain)
	require.NoError(t, err)

	_, err = runtime.GetContainer(ctx, containerID, cfg.NamespaceMain)
	require.Error(t, err)
	require.True(t, errdefs.IsNotFound(err))
}

func TestRemoveContainer_NonExistentContainer(t *testing.T) {
	runtime, cfg := setup(t)
	defer teardown(t, cfg)

	ctx := namespaces.WithNamespace(context.Background(), cfg.NamespaceMain)

	nonExistentContainerID := "non-existent-container"

	// Attempt to remove a non-existent container
	err := runtime.RemoveContainer(ctx, nonExistentContainerID, cfg.NamespaceMain)
	require.Error(t, err, "Removing a non-existent container should return an error")
	require.True(t, errdefs.IsNotFound(err), "Error should be of type NotFound")
}

func TestRemoveContainer_AlreadyRemovedContainer(t *testing.T) {
	runtime, cfg := setup(t)
	defer teardown(t, cfg)

	ctx := namespaces.WithNamespace(context.Background(), cfg.NamespaceMain)

	containerID := "test-already-removed-container"
	image := "docker.io/library/alpine:latest"

	createdContainer, err := runtime.CreateContainer(ctx, containerID, cfg.NamespaceMain, image)
	require.NoError(t, err)
	require.NotNil(t, createdContainer)

	err = createdContainer.Delete(ctx, containerd.WithSnapshotCleanup)
	require.NoError(t, err)

	// Attempt to remove the container again
	err = runtime.RemoveContainer(ctx, containerID, cfg.NamespaceMain)
	require.Error(t, err, "Removing an already removed container should return an error")
	require.True(t, errdefs.IsNotFound(err), "Error should be of type NotFound")
}

// StartContainer
func TestStartContainer_Valid(t *testing.T) {
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

func TestStartContainer_NonExistentContainer(t *testing.T) {
	runtime, cfg := setup(t)
	defer teardown(t, cfg)

	ctx := namespaces.WithNamespace(context.Background(), cfg.NamespaceMain)

	nonExistentContainerID := "non-existent-container"

	// Attempt to start a non-existent container
	task, err := runtime.StartContainer(ctx, nonExistentContainerID, cfg.NamespaceMain)
	require.Error(t, err, "Starting a non-existent container should return an error")
	require.Nil(t, task, "Task should be nil for a non-existent container")
}

func TestStartContainer_AlreadyRunningContainer(t *testing.T) {
	runtime, cfg := setup(t)
	defer teardown(t, cfg)

	ctx := namespaces.WithNamespace(context.Background(), cfg.NamespaceMain)

	containerID := "test-already-running-container"
	image := "docker.io/library/alpine:latest"

	createdContainer, err := runtime.CreateContainer(ctx, containerID, cfg.NamespaceMain, image)
	require.NoError(t, err)
	require.NotNil(t, createdContainer)

	task, err := runtime.StartContainer(ctx, containerID, cfg.NamespaceMain)
	require.NoError(t, err)
	require.NotNil(t, task)

	// Attempt to start the container again while it's already running
	taskAgain, err := runtime.StartContainer(ctx, containerID, cfg.NamespaceMain)
	require.Error(t, err, "Starting an already running container should return an error")
	require.Nil(t, taskAgain, "Task should be nil when the container is already running")
}

// StopContainer
func TestStopContainer_Valid(t *testing.T) {
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

	exitCh, err := runtime.StopContainer(ctx, containerID, cfg.NamespaceMain, syscall.SIGKILL)
	require.NoError(t, err)

	// Wait for the exit status
	exitStatus := <-exitCh
	require.NoError(t, exitStatus.Error())

	status, err = task.Status(ctx)
	require.NoError(t, err)
	require.Equal(t, containerd.Stopped, status.Status)
}

func TestStopContainer_NonExistentContainer(t *testing.T) {
	runtime, cfg := setup(t)
	defer teardown(t, cfg)

	ctx := namespaces.WithNamespace(context.Background(), cfg.NamespaceMain)

	nonExistentContainerID := "non-existent-container"

	// Attempt to stop a non-existent container
	exitCh, err := runtime.StopContainer(ctx, nonExistentContainerID, cfg.NamespaceMain, syscall.SIGKILL)
	require.Error(t, err, "Stopping a non-existent container should return an error")
	require.Nil(t, exitCh, "Exit channel should be nil for a non-existent container")
}

func TestStopContainer_AlreadyStoppedContainer(t *testing.T) {
	runtime, cfg := setup(t)
	defer teardown(t, cfg)

	ctx := namespaces.WithNamespace(context.Background(), cfg.NamespaceMain)

	containerID := "test-already-stopped-container"
	image := "docker.io/library/alpine:latest"

	createdContainer, err := runtime.CreateContainer(ctx, containerID, cfg.NamespaceMain, image)
	require.NoError(t, err)
	require.NotNil(t, createdContainer)

	task, err := runtime.StartContainer(ctx, containerID, cfg.NamespaceMain)
	require.NoError(t, err)
	require.NotNil(t, task)

	// Stop the container
	exitCh, err := runtime.StopContainer(ctx, containerID, cfg.NamespaceMain, syscall.SIGKILL)
	require.NoError(t, err)

	// Wait for the exit status
	exitStatus := <-exitCh
	require.NoError(t, exitStatus.Error())

	status, err := task.Status(ctx)
	require.NoError(t, err)
	require.Equal(t, containerd.Stopped, status.Status)

	// Attempt to stop the container again
	exitCh, err = runtime.StopContainer(ctx, containerID, cfg.NamespaceMain, syscall.SIGKILL)
	require.Error(t, err, "Stopping an already stopped container should return an error")
	require.Nil(t, exitCh, "Exit channel should be nil when the container is already stopped")
}

// GetContainer
func TestGetContainer_Valid(t *testing.T) {
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

func TestGetContainer_NonExistentContainer(t *testing.T) {
	runtime, cfg := setup(t)
	defer teardown(t, cfg)

	ctx := namespaces.WithNamespace(context.Background(), cfg.NamespaceMain)

	nonExistentContainerID := "non-existent-container"

	// Attempt to retrieve a non-existent container
	retrievedContainer, err := runtime.GetContainer(ctx, nonExistentContainerID, cfg.NamespaceMain)
	require.Error(t, err, "Retrieving a non-existent container should return an error")
	require.Nil(t, retrievedContainer, "Retrieved container should be nil for a non-existent container")
}

func TestGetContainer_AfterRemoval(t *testing.T) {
	runtime, cfg := setup(t)
	defer teardown(t, cfg)

	ctx := namespaces.WithNamespace(context.Background(), cfg.NamespaceMain)

	containerID := "test-get-container-after-removal"
	image := "docker.io/library/alpine:latest"

	createdContainer, err := runtime.CreateContainer(ctx, containerID, cfg.NamespaceMain, image)
	require.NoError(t, err)
	require.NotNil(t, createdContainer)

	err = createdContainer.Delete(ctx, containerd.WithSnapshotCleanup)
	require.NoError(t, err)

	// Attempt to retrieve the container after it has been removed
	retrievedContainer, err := runtime.GetContainer(ctx, containerID, cfg.NamespaceMain)
	require.Error(t, err, "Retrieving a container after it has been removed should return an error")
	require.Nil(t, retrievedContainer, "Retrieved container should be nil for a removed container")
}

// GetContainers
func TestGetContainers_Valid(t *testing.T) {
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

func TestGetContainers_NoContainers(t *testing.T) {
	runtime, cfg := setup(t)
	defer teardown(t, cfg)

	ctx := namespaces.WithNamespace(context.Background(), cfg.NamespaceMain)

	// Attempt to retrieve containers when none exist
	containers, err := runtime.GetContainers(ctx, cfg.NamespaceMain)
	require.NoError(t, err, "Retrieving containers in an empty namespace should not return an error")
	require.Empty(t, containers, "The containers list should be empty when no containers exist")
}

func TestGetContainers_AfterSomeAreRemoved(t *testing.T) {
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

	// Remove one of the containers
	err := runtime.RemoveContainer(ctx, containerIDs[1], cfg.NamespaceMain)
	require.NoError(t, err, "Removing a container should not return an error")

	containers, err := runtime.GetContainers(ctx, cfg.NamespaceMain)
	require.NoError(t, err)
	require.NotNil(t, containers)
	require.Len(t, containers, len(containerIDs)-1, "The number of containers retrieved should match the number remaining")

	retrievedIDs := make(map[string]bool)
	for _, container := range containers {
		retrievedIDs[container.ID()] = true
	}

	require.False(t, retrievedIDs[containerIDs[1]], "Removed container %s should not be present in the retrieved list", containerIDs[1])
	require.True(t, retrievedIDs[containerIDs[0]], "Container %s should be present in the retrieved list", containerIDs[0])
	require.True(t, retrievedIDs[containerIDs[2]], "Container %s should be present in the retrieved list", containerIDs[2])
}
