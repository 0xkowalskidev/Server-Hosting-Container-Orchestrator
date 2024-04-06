package statemanager

import (
	"0xKowalski1/container-orchestrator/models"
	"context"
	"encoding/json"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// AddContainer adds a new container to a namespace
func (sm *StateManager) AddContainer(container models.Container) error {
	container.NamespaceID = sm.cfg.Namespace // Ensure the container knows its namespaceID
	container.DesiredStatus = "running"

	err := sm.etcdClient.SaveEntity(container)
	if err != nil {
		return err
	}

	sm.emit(Event{Type: ContainerAdded, Data: container})

	return nil
}

// RemoveContainer removes a container from a namespace by its ID
func (sm *StateManager) RemoveContainer(containerID string) error {
	namespaceID := sm.cfg.Namespace
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// IMPORTANT Should probably check if namespaces match here (and everywhere else we change state)

	err := sm.RemoveContainerFromNode(containerID)

	if err != nil {
		fmt.Printf("Failed to remove container from node: %v", err)

		return err
	}

	key := "/namespaces/" + namespaceID + "/containers/" + containerID
	_, err = sm.etcdClient.Client.Delete(ctx, key)

	if err != nil {
		fmt.Printf("Failed to delete container: %v", err)
		return err
	}

	sm.emit(Event{Type: ContainerRemoved, Data: containerID})

	return nil
}

// GetContainer retrieves a container by its ID and namespaceID
func (sm *StateManager) GetContainer(containerID string) (*models.Container, error) {
	namespaceID := sm.cfg.Namespace
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := "/namespaces/" + namespaceID + "/containers/" + containerID
	resp, err := sm.etcdClient.Client.Get(ctx, key)
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
func (sm *StateManager) ListContainers() ([]models.Container, error) {
	namespaceID := sm.cfg.Namespace
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	prefix := "/namespaces/" + namespaceID + "/containers/"
	resp, err := sm.etcdClient.Client.Get(ctx, prefix, clientv3.WithPrefix())
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
func (sm *StateManager) ListUnscheduledContainers() ([]models.Container, error) {
	containers, err := sm.ListContainers()
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
func (sm *StateManager) PatchContainer(containerID string, patch models.UpdateContainerRequest) error {
	container, err := sm.GetContainer(containerID)
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

	return sm.etcdClient.SaveEntity(*container)
}

func (sm *StateManager) SubscribeToStatus(containerID string) (chan string, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Create a new channel for this subscription
	statusChan := make(chan string)

	if _, ok := sm.subscriptions[containerID]; !ok {
		sm.subscriptions[containerID] = make(map[chan string]struct{})
		// Start watching etcd for changes to this container's status
		go sm.watchStatus(containerID)
	}

	sm.subscriptions[containerID][statusChan] = struct{}{}

	return statusChan, nil
}

func (sm *StateManager) UnsubscribeFromStatus(containerID string, statusChan chan string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	subscribers, ok := sm.subscriptions[containerID]
	if !ok {
		return
	}

	// Remove the channel from the subscribers
	delete(subscribers, statusChan)
	close(statusChan)

	// If there are no more subscribers, stop watching this container's status
	if len(subscribers) == 0 {
		delete(sm.subscriptions, containerID)
	}
}

func (sm *StateManager) watchStatus(containerID string) {
	ctx := context.Background()
	watchChan := sm.etcdClient.Watch(ctx, "/namespaces/"+sm.cfg.Namespace+"/containers/"+containerID)

	for watchResp := range watchChan {
		for _, event := range watchResp.Events {
			sm.mu.Lock()
			containerData := string(event.Kv.Value)
			for subscriberChan := range sm.subscriptions[containerID] {
				subscriberChan <- containerData
			}
			sm.mu.Unlock()
		}
	}
}
