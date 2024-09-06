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
	"github.com/opencontainers/runtime-spec/specs-go"
)

type ContainerdRuntime struct {
	client *containerd.Client
	cfg    config.Config
}

func NewRuntime(cfg config.Config) (*ContainerdRuntime, error) {
	client, err := containerd.New(cfg.ContainerdPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create containerd client: %w", err)
	}
	return &ContainerdRuntime{client: client, cfg: cfg}, nil
}

func (c *ContainerdRuntime) CreateContainer(ctx context.Context, id string, namespace string, image string) (containerd.Container, error) {
	ctx = namespaces.WithNamespace(ctx, namespace)

	imageRef, err := c.client.Pull(ctx, image, containerd.WithPullUnpack)
	if err != nil {
		return nil, fmt.Errorf("failed to pull image %s for container with id %s in namespace %s: %w", image, namespace, id, err)
	}

	specOpts := []oci.SpecOpts{
		oci.WithImageConfig(imageRef),
		oci.WithHostNamespace(specs.NetworkNamespace),
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

	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return nil, fmt.Errorf("failed to create task for container with id %s in namespace %s: %w", id, namespace, err)
	}

	if err := task.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start task for container with id %s in namespace %s: %w", id, namespace, err)
	}

	return task, nil
}

func (c *ContainerdRuntime) StopContainer(ctx context.Context, id string, namespace string, signal syscall.Signal) (<-chan containerd.ExitStatus, error) {
	ctx = namespaces.WithNamespace(ctx, namespace)

	container, err := c.GetContainer(ctx, id, namespace)
	if err != nil {
		return nil, err
	}

	task, err := container.Task(ctx, cio.Load)
	if err != nil {
		return nil, fmt.Errorf("failed to load task for container with id %s in namespace %s: %w", id, namespace, err)
	}

	if err := task.Kill(ctx, signal); err != nil {
		return nil, fmt.Errorf("failed to send SIGKILL to kill task for container with id %s in namespace %s: %w", id, namespace, err)
	}

	statusCh, err := task.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed waiting for task to exit for container with id %s in namespace %s: %w", id, namespace, err)
	}

	return statusCh, nil
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
