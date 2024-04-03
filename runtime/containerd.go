package runtime

import (
	"context"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"log"
)

// ContainerdRuntime implements the Runtime interface for containerd.
type ContainerdRuntime struct {
	client *containerd.Client
}

// NewContainerdRuntime creates a new instance of ContainerdRuntime with the given containerd client.
func NewContainerdRuntime(socketPath string) (*ContainerdRuntime, error) {
	client, err := containerd.New(socketPath)

	if err != nil {
		return nil, err
	}

	return &ContainerdRuntime{
		client: client,
	}, nil
}

// CreateContainer instantiates a new container but does not start it.
//func CreateContainer(ctx context.Context, config ContainerConfig) (Container, error) {}

// StartContainer starts an existing container.
//func StartContainer(ctx context.Context, containerID string) error {}

// StopContainer stops a running container.
//func StopContainer(ctx context.Context, containerID string, timeout int) error {}

// RemoveContainer removes a container from the system. This may require the container to be stopped first.
//func RemoveContainer(ctx context.Context, containerID string) error {}

// ListContainers returns a list of all containers managed by the runtime.
func (_runtime *ContainerdRuntime) ListContainers(namespace string) ([]Container, error) {
	var containers []Container
	ctx := namespaces.WithNamespace(context.Background(), namespace)

	// List containers from containerd
	conts, err := _runtime.client.Containers(ctx)
	if err != nil {
		log.Printf("Error listing containers: %v", err)
		return nil, err
	}

	// Map containerd containers to generic Container struct
	for _, cont := range conts {
		/*info*/ _, err := cont.Info(ctx)
		if err != nil {
			log.Printf("Error getting container info: %v", err)
			continue
		}

		containers = append(containers, Container{
			ID:     cont.ID(),
			Status: string("running"), //temp
		})
	}

	return containers, nil
}

// GetContainerLogs returns the logs for a specific container.
//func GetContainerLogs(ctx context.Context, containerID string) (io.ReadCloser, error) {}

// InspectContainer returns detailed information about a specific container.
//func InspectContainer(ctx context.Context, containerID string) (Container, error) {}
