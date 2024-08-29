package workernode

import (
	"context"
	"errors"

	"github.com/0xKowalski1/Server-Hosting-Container-Orchestrator/config"
	"github.com/containerd/containerd"
)

type Runtime interface {
	CreateContainer(ctx context.Context, id string, image string) (containerd.Container, error)
}

func NewRuntime(cfg config.Config) (Runtime, error) {
	switch cfg.RuntimeType {
	case "containerd":
		return newContainerdRuntime(cfg)
	default:
		return nil, errors.New("unsupported runtime type")
	}
}
