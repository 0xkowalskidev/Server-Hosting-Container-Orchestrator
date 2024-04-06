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
	// IMPORTANT Should probably check if namespaces match here

	key := "/namespaces/" + namespaceID + "/containers/" + containerID
	_, err := sm.etcdClient.Client.Delete(ctx, key)

	if err == nil {
		sm.emit(Event{Type: ContainerRemoved, Data: containerID})
	}

	return err
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
