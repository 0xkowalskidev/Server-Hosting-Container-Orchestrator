package controlnode

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"0xKowalski1/container-orchestrator/config"
	"0xKowalski1/container-orchestrator/models"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// ContainerService handles operations related to containers
type ContainerService struct {
	cfg        *config.Config
	etcdClient *EtcdClient
}

// NewContainerService creates a new ContainerService
func NewContainerService(cfg *config.Config, etcdClient *EtcdClient) *ContainerService {
	return &ContainerService{
		cfg:        cfg,
		etcdClient: etcdClient,
	}
}

// AddContainer adds a new container to a namespace
func (cs *ContainerService) CreateContainer(containerRequest models.CreateContainerRequest) (*models.Container, error) {
	container := models.Container{
		ID:            containerRequest.ID,
		Image:         containerRequest.Image,
		Env:           containerRequest.Env,
		StopTimeout:   containerRequest.StopTimeout,
		MemoryLimit:   containerRequest.MemoryLimit,
		CpuLimit:      containerRequest.CpuLimit,
		StorageLimit:  containerRequest.StorageLimit,
		NamespaceID:   cs.cfg.Namespace, // Ensure the container knows its namespaceID
		DesiredStatus: "running",
		Ports:         containerRequest.Ports,
	}

	err := cs.etcdClient.SaveEntity(container)
	if err != nil {
		return nil, err
	}

	cs.etcdClient.emit(Event{Type: ContainerAdded, Data: container})

	return &container, nil
}

// RemoveContainer removes a container from a namespace by its ID
func (cs *ContainerService) DeleteContainer(containerID string, nodeService *NodeService) error {
	namespaceID := cs.cfg.Namespace
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := nodeService.RemoveContainerFromNode(containerID)

	if err != nil {
		fmt.Printf("Failed to remove container from node: %v", err)

		return err
	}

	key := "/namespaces/" + namespaceID + "/containers/" + containerID
	_, err = cs.etcdClient.Client.Delete(ctx, key)
	if err != nil {
		fmt.Printf("Failed to delete container: %v", err)
		return err
	}

	cs.etcdClient.emit(Event{Type: ContainerRemoved, Data: containerID})

	return nil
}

// GetContainer retrieves a container by its ID and namespaceID
func (cs *ContainerService) GetContainer(containerID string) (*models.Container, error) {
	namespaceID := cs.cfg.Namespace
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := "/namespaces/" + namespaceID + "/containers/" + containerID
	resp, err := cs.etcdClient.Client.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("container not found")
	}

	var container models.Container
	err = json.Unmarshal(resp.Kvs[0].Value, &container)
	if err != nil {
		return nil, err
	}

	return &container, nil
}

// ListContainers lists all containers in a namespace
func (cs *ContainerService) GetContainers() ([]models.Container, error) {
	namespaceID := cs.cfg.Namespace
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	prefix := "/namespaces/" + namespaceID + "/containers/"
	resp, err := cs.etcdClient.Client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	containers := make([]models.Container, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var container models.Container
		if err := json.Unmarshal(kv.Value, &container); err != nil {
			continue
		}
		containers = append(containers, container)
	}

	return containers, nil
}

// ListUnscheduledContainers lists all containers that do not have a NodeID set.
func (cs *ContainerService) GetUnscheduledContainers() ([]models.Container, error) {
	containers, err := cs.GetContainers()
	if err != nil {
		return nil, err
	}

	unscheduledContainers := make([]models.Container, 0, len(containers))

	for _, container := range containers {
		if container.NodeID == "" { // Filter for unscheduled (no NodeID set)
			unscheduledContainers = append(unscheduledContainers, container)
		}
	}

	return unscheduledContainers, nil
}

// PatchContainer updates specific fields of a container in a namespace.
func (cs *ContainerService) UpdateContainer(containerID string, patch models.UpdateContainerRequest) error {
	container, err := cs.GetContainer(containerID)
	if err != nil {
		return err
	}

	// Apply the patch
	if patch.DesiredStatus != nil {
		container.DesiredStatus = *patch.DesiredStatus
	}
	if patch.NodeID != nil {
		container.NodeID = *patch.NodeID
	}
	if patch.Status != nil {
		container.Status = *patch.Status
	}

	return cs.etcdClient.SaveEntity(*container)
}

// SubscribeToStatus subscribes to status updates for a container
func (cs *ContainerService) SubscribeToStatus(containerID string) (chan string, error) {
	cs.etcdClient.mu.Lock()
	defer cs.etcdClient.mu.Unlock()

	// Create a new channel for this subscription
	statusChan := make(chan string)

	if _, ok := cs.etcdClient.subscriptions[containerID]; !ok {
		cs.etcdClient.subscriptions[containerID] = make(map[chan string]struct{})
		// Start watching etcd for changes to this container's status
		go cs.watchStatus(containerID)
	}

	cs.etcdClient.subscriptions[containerID][statusChan] = struct{}{}

	return statusChan, nil
}

// UnsubscribeFromStatus unsubscribes from status updates for a container
func (cs *ContainerService) UnsubscribeFromStatus(containerID string, statusChan chan string) {
	cs.etcdClient.mu.Lock()
	defer cs.etcdClient.mu.Unlock()

	subscribers, ok := cs.etcdClient.subscriptions[containerID]
	if !ok {
		return
	}

	// Remove the channel from the subscribers
	delete(subscribers, statusChan)
	close(statusChan)

	// If there are no more subscribers, stop watching this container's status
	if len(subscribers) == 0 {
		delete(cs.etcdClient.subscriptions, containerID)
	}
}

// watchStatus watches the status of a container for changes
func (cs *ContainerService) watchStatus(containerID string) {
	ctx := context.Background()
	watchChan := cs.etcdClient.Watch(ctx, "/namespaces/"+cs.cfg.Namespace+"/containers/"+containerID)

	for watchResp := range watchChan {
		for _, event := range watchResp.Events {
			cs.etcdClient.mu.Lock()
			containerData := string(event.Kv.Value)
			for subscriberChan := range cs.etcdClient.subscriptions[containerID] {
				subscriberChan <- containerData
			}
			cs.etcdClient.mu.Unlock()
		}
	}
}
