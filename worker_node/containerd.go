package workernode

import (
	"context"
	"fmt"
	"syscall"

	"github.com/0xKowalski1/Server-Hosting-Container-Orchestrator/config"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
)

type ContainerdRuntime struct {
	client *containerd.Client
	cfg    config.Config
}

func newContainerdRuntime(cfg config.Config) (Runtime, error) {
	client, err := containerd.New(cfg.ContainerdPath)

	if err != nil {
		return nil, err
	}

	return &ContainerdRuntime{client: client, cfg: cfg}, nil
}

func (c *ContainerdRuntime) CreateContainer(ctx context.Context, id string, namespace string, image string) (containerd.Container, error) {
	ctx = namespaces.WithNamespace(ctx, namespace)

	imageRef, err := c.client.Pull(ctx, image, containerd.WithPullUnpack)
	if err != nil {
		return nil, fmt.Errorf("Failed to pull image %s for container with id %s: %w", image, id, err)
	}

	specOpts := []oci.SpecOpts{
		oci.WithImageConfig(imageRef),
	}

	container, err := c.client.NewContainer(
		ctx,
		id,
		containerd.WithNewSnapshot(id+"-snapshot", imageRef),
		containerd.WithNewSpec(specOpts...),
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to create container with id %s: %w", id, err)
	}

	return container, nil
}

func (c *ContainerdRuntime) RemoveContainer(ctx context.Context, id string, namespace string) error {
	ctx = namespaces.WithNamespace(ctx, namespace)

	container, err := c.client.LoadContainer(ctx, id)
	if err != nil {
		return fmt.Errorf("Failed to load container with id %s: %w", id, err)
	}

	if err := container.Delete(ctx, containerd.WithSnapshotCleanup); err != nil {
		return fmt.Errorf("Failed to remove container with id %s: %w", id, err)

	}

	return nil
}

func (c *ContainerdRuntime) StartContainer(ctx context.Context, id string, namespace string) (containerd.Task, error) {
	ctx = namespaces.WithNamespace(ctx, namespace)

	container, err := c.client.LoadContainer(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Failed to load container with id %s: %w", id, err)
	}

	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return nil, fmt.Errorf("Failed to create task for container with id %s: %w", id, err)
	}

	if err := task.Start(ctx); err != nil {
		return nil, fmt.Errorf("Failed to start task for container with id %s: %w", id, err)
	}

	return task, nil
}

func (c *ContainerdRuntime) StopContainer(ctx context.Context, id string, namespace string) (<-chan containerd.ExitStatus, error) {
	ctx = namespaces.WithNamespace(ctx, namespace)

	container, err := c.client.LoadContainer(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Failed to load container with id %s: %w", id, err)
	}

	task, err := container.Task(ctx, cio.Load)
	if err != nil {
		return nil, fmt.Errorf("Failed to load task for container with id %s: %w", id, err)
	}

	if err := task.Kill(ctx, syscall.SIGKILL); err != nil {
		return nil, fmt.Errorf("Failed to send SIGKILL to kill task for container with id %s: %v", id, err)
	}

	statusCh, err := task.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to wait for task to exit for container with id %s: %v", id, err)
	}

	return statusCh, nil
}

func (c *ContainerdRuntime) GetContainer(ctx context.Context, id string, namespace string) (containerd.Container, error) {
	ctx = namespaces.WithNamespace(ctx, namespace)

	container, err := c.client.LoadContainer(ctx, id)
	if err != nil {
		return container, fmt.Errorf("Failed to load container with id %s in namespace %s: %w", id, namespace, err)

	}

	return container, nil
}

func (c *ContainerdRuntime) GetContainers(ctx context.Context, namespace string) ([]containerd.Container, error) {
	ctx = namespaces.WithNamespace(ctx, namespace)

	containers, err := c.client.Containers(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to load containers in namespace %s: %w", namespace, err)
	}

	return containers, nil
}
