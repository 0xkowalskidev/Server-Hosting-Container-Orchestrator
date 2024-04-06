package runtime

import (
	"fmt"

	statemanager "0xKowalski1/container-orchestrator/state-manager"
)

// NewRuntime creates and returns a Runtime implementation based on the provided config.
func NewRuntime(backend string) (Runtime, error) {
	switch backend {
	case "containerd":
		runtime, err := NewContainerdRuntime("/run/containerd/containerd.sock")
		if err != nil {
			return nil, err
		}
		runtime.SubscribeToEvents("example")

		return runtime, nil
	default:
		return nil, fmt.Errorf("unsupported runtime type: %s", backend)
	}
}

// Runtime defines the interface for a container runtime.
type Runtime interface {
	// CreateContainer instantiates a new container but does not start it.
	CreateContainer(namespace string, config statemanager.Container) (statemanager.Container, error)

	// StartContainer starts an existing container.
	StartContainer(namespace string, containerID string) error

	// StopContainer stops a running container.
	StopContainer(namespace string, containerID string, timeout int) error

	// RemoveContainer removes a container from the system. This may require the container to be stopped first.
	RemoveContainer(namespace string, containerID string) error

	// ListContainers returns a list of all containers managed by the runtime.
	ListContainers(namespace string) ([]statemanager.Container, error)

	// InspectContainer returns detailed information about a specific container.
	InspectContainer(namespace string, containerID string) (statemanager.Container, error)
}
