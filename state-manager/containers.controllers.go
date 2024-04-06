package statemanager

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// AddContainer adds a new container to a namespace
func (sm *StateManager) AddContainer(namespaceID string, container Container) error {
	//Check namespace exists
	_, err := sm.GetNamespace(namespaceID)
	if err != nil {
		return err
	}

	container.NamespaceID = namespaceID // Ensure the container knows its namespaceID
	container.DesiredStatus = "running"

	err = sm.etcdClient.SaveEntity(container)
	if err != nil {
		return err
	}

	sm.emit(Event{Type: ContainerAdded, Data: container})

	return nil
}

// RemoveContainer removes a container from a namespace by its ID
func (sm *StateManager) RemoveContainer(namespaceID, containerID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := "/namespaces/" + namespaceID + "/containers/" + containerID
	_, err := sm.etcdClient.Client.Delete(ctx, key)

	if err == nil {
		sm.emit(Event{Type: ContainerRemoved, Data: containerID})
	}

	return err
}

// GetContainer retrieves a container by its ID and namespaceID
func (sm *StateManager) GetContainer(namespaceID, containerID string) (*Container, error) {
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

	var container Container
	err = json.Unmarshal(resp.Kvs[0].Value, &container)
	if err != nil {
		return nil, err
	}

	return &container, nil
}

// ListContainers lists all containers in a namespace
func (sm *StateManager) ListContainers(namespaceID string) ([]Container, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	prefix := "/namespaces/" + namespaceID + "/containers/"
	resp, err := sm.etcdClient.Client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	containers := make([]Container, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var container Container
		if err := json.Unmarshal(kv.Value, &container); err != nil {
			continue
		}
		containers = append(containers, container)
	}

	return containers, nil
}

// ListUnscheduledContainers lists all containers across all namespaces that do not have a NodeID set.
func (sm *StateManager) ListUnscheduledContainers() ([]Container, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Prefix for namespaces in etcd
	namespacePrefix := "/namespaces/"
	resp, err := sm.etcdClient.Client.Get(ctx, namespacePrefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	unscheduledContainers := []Container{}
	for _, kv := range resp.Kvs {
		var namespace Namespace
		if err := json.Unmarshal(kv.Value, &namespace); err != nil {
			continue // Skip on error
		}

		// Now, list containers within this namespace
		namespaceContainers, err := sm.ListContainers(namespace.ID)
		if err != nil {
			continue // Skip on error
		}

		for _, container := range namespaceContainers {
			if container.NodeID == "" { // Filter for unscheduled (no NodeID set)
				unscheduledContainers = append(unscheduledContainers, container)
			}
		}
	}

	return unscheduledContainers, nil
}

// PatchContainer updates specific fields of a container in a namespace.
func (sm *StateManager) PatchContainer(namespaceID, containerID string, patch UpdateContainerRequest) error {
	container, err := sm.GetContainer(namespaceID, containerID)
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
