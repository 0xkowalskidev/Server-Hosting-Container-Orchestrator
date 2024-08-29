package workernode

import (
	"context"
	"fmt"
	"log"

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

func (c *ContainerdRuntime) CreateContainer(ctx context.Context, id string, image string) (containerd.Container, error) {
	log.Println(c.cfg.NamespaceMain)
	ctx = namespaces.WithNamespace(ctx, c.cfg.NamespaceMain)

	// Pull the image if it doesn't exist locally
	imageRef, err := c.client.Pull(ctx, image, containerd.WithPullUnpack)
	if err != nil {
		return nil, fmt.Errorf("failed to pull image %s: %w", image, err)
	}

	specOpts := []oci.SpecOpts{
		oci.WithImageConfig(imageRef),
	}

	// Create the container
	container, err := c.client.NewContainer(
		ctx,
		id,
		containerd.WithNewSnapshot(id+"-snapshot", imageRef),
		containerd.WithNewSpec(specOpts...),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Create a task to run the container
	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return nil, fmt.Errorf("failed to create task for container: %w", err)
	}

	// Start the container
	if err := task.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start task for container: %w", err)
	}

	return container, nil
}
