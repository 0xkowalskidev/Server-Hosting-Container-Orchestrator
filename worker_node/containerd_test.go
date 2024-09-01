package workernode_test

import (
	"context"
	"testing"

	"github.com/0xKowalski1/Server-Hosting-Container-Orchestrator/config"
	workernode "github.com/0xKowalski1/Server-Hosting-Container-Orchestrator/worker_node"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/stretchr/testify/require"
)

func TestCreateContainer(t *testing.T) {
	var cfg config.Config
	config.ParseConfigFromEnv(&cfg)

	runtime, err := workernode.NewRuntime(cfg)
	require.NoError(t, err)
	require.NotNil(t, runtime)

	ctx := namespaces.WithNamespace(context.Background(), cfg.NamespaceMain)

	containerID := "integration-test-container"
	image := "docker.io/library/alpine:latest"

	// Create the container
	createdContainer, err := runtime.CreateContainer(ctx, containerID, cfg.NamespaceMain, image)
	require.NoError(t, err)
	require.NotNil(t, createdContainer)

	// Retrieve the task associated with the container
	task, err := createdContainer.Task(ctx, nil)
	require.NoError(t, err)

	// Verify the task status
	status, err := task.Status(ctx)
	require.NoError(t, err)
	require.Equal(t, containerd.Running, status.Status)
}
