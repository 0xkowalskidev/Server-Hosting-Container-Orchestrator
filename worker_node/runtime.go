package workernode

import (
	"context"
	"errors"

	"github.com/containerd/containerd"
)

type Runtime interface {
	CreateContainer(ctx context.Context, id string, image string) (containerd.Container, error)
}

func NewRuntime(runtimeType string) (Runtime, error) {
	switch runtimeType {
	case "containerd":
		return newContainerdRuntime("/run/containerd/containerd.sock")
	default:
		return nil, errors.New("unsupported runtime type")
	}
}
