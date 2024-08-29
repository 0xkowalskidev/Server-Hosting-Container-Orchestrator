package workernode_test

import (
	"context"
	"syscall"
	"testing"
	"time"

	workernode "github.com/0xKowalski1/Server-Hosting-Container-Orchestrator/worker_node"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/stretchr/testify/require"
)

func TestCreateContainer(t *testing.T) {
	runtime, err := workernode.NewRuntime("containerd")
	require.NoError(t, err)
	require.NotNil(t, runtime)

	ctx := namespaces.WithNamespace(context.Background(), "default")

	// Define the container ID and image to use
	containerID := "integration-test-container"
	image := "docker.io/library/alpine:latest"

	// Create the container
	createdContainer, err := runtime.CreateContainer(ctx, containerID, image)
	require.NoError(t, err)
	require.NotNil(t, createdContainer)

	// Retrieve the task associated with the container
	task, err := createdContainer.Task(ctx, nil)
	require.NoError(t, err)

	// Verify the task status
	status, err := task.Status(ctx)
	require.NoError(t, err)
	require.Equal(t, containerd.Running, status.Status)

	// Allow some time for the container to run (e.g., to ensure it's not immediately failing)
	time.Sleep(2 * time.Second)

	// Clean up: stop the task, delete the container and snapshot
	err = task.Kill(ctx, syscall.SIGTERM) // Use syscall.SIGTERM instead of containerd.SIGTERM
	require.NoError(t, err)

	_, err = task.Delete(ctx, containerd.WithProcessKill)
	require.NoError(t, err)

	err = createdContainer.Delete(ctx, containerd.WithSnapshotCleanup)
	require.NoError(t, err)
}
