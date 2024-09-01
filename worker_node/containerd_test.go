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

	// Deletes all containers/tasks in the containerd namespace
	for _, container := range containers {
		task, err := container.Task(ctx, nil)
		if err == nil {
			if err := task.Kill(ctx, syscall.SIGINT); err != nil {
				t.Logf("Failed to kill task for container %s: %v", container.ID(), err)
			}

			_, err = task.Delete(ctx)
			if err != nil {
				t.Logf("Failed to delete task for container %s: %v", container.ID(), err)
			}
		} else if !errdefs.IsNotFound(err) {
			t.Logf("Failed to retrieve task for container %s: %v", container.ID(), err)
		}

		if err := container.Delete(ctx, containerd.WithSnapshotCleanup); err != nil {
			t.Logf("Failed to delete container %s: %v", container.ID(), err)
		}
	}

	t.Logf("Successfully cleaned up containerd namespace: %s", cfg.NamespaceMain)
}

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
