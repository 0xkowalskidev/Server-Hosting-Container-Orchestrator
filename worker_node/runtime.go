package workernode

import (
	"context"
	"errors"
	"syscall"

	"github.com/0xKowalski1/Server-Hosting-Container-Orchestrator/config"
	"github.com/containerd/containerd"
)

type Runtime interface {
	CreateContainer(ctx context.Context, id string, namespace string, image string) (containerd.Container, error)
	RemoveContainer(ctx context.Context, id string, namespace string) error
	StartContainer(ctx context.Context, id string, namespace string) (containerd.Task, error)
	StopContainer(ctx context.Context, id string, namespace string, signal syscall.Signal) (<-chan containerd.ExitStatus, error)
	GetContainer(ctx context.Context, id string, namespace string) (containerd.Container, error)
	GetContainers(ctx context.Context, namespace string) ([]containerd.Container, error)
}

func NewRuntime(cfg config.Config) (Runtime, error) {
	switch cfg.RuntimeType {
	case "containerd":
		return newContainerdRuntime(cfg)
	default:
		return nil, errors.New("unsupported runtime type")
	}
}
