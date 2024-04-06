package runtime

import (
	"fmt"

	"0xKowalski1/container-orchestrator/config"
	"0xKowalski1/container-orchestrator/models"
)

// NewRuntime creates and returns a Runtime implementation based on the provided config.
func NewRuntime(backend string, cfg *config.Config) (Runtime, error) {
	switch backend {
	case "containerd":
		runtime, err := NewContainerdRuntime("/run/containerd/containerd.sock")
		if err != nil {
			return nil, err
		}
		runtime.SubscribeToEvents(cfg.Namespace)

		return runtime, nil
	default:
		return nil, fmt.Errorf("unsupported runtime type: %s", backend)
	}
}

// Runtime defines the interface for a container runtime.
type Runtime interface {
	// CreateContainer instantiates a new container but does not start it.
	CreateContainer(namespace string, config models.Container) (models.Container, error)

	// StartContainer starts an existing container.
	StartContainer(namespace string, containerID string) error

	// StopContainer stops a running container.
	StopContainer(namespace string, containerID string, timeout int) error

	// RemoveContainer removes a container from the system. This may require the container to be stopped first.
	RemoveContainer(namespace string, containerID string) error

	// ListContainers returns a list of all containers managed by the runtime.
	ListContainers(namespace string) ([]models.Container, error)

	// InspectContainer returns detailed information about a specific container.
	InspectContainer(namespace string, containerID string) (models.Container, error)
}
