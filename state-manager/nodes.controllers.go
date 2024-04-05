package statemanager

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// AddNode adds a new node to the cluster
func (sm *StateManager) AddNode(node Node) error {
	return sm.etcdClient.SaveEntity(node)
}

// RemoveNode removes a node from the cluster by its ID
func (sm *StateManager) RemoveNode(nodeID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := "/nodes/" + nodeID
	_, err := sm.etcdClient.Delete(ctx, key)
	return err
}

// GetNode retrieves a node by its ID
func (sm *StateManager) GetNode(nodeID string) (*Node, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := "/nodes/" + nodeID
	resp, err := sm.etcdClient.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("node not found")
	}

	var node Node
	err = json.Unmarshal(resp.Kvs[0].Value, &node)
	if err != nil {
		return nil, err
	}

	populatedContainers := make([]Container, 0, len(node.Containers))
	for _, container := range node.Containers {
		container, err := sm.GetContainer(container.NamespaceID, container.ID)
		if err != nil {
			fmt.Printf("Failed to populate container for node: %v", err)
			continue

		}
		populatedContainers = append(populatedContainers, *container)
	}
	node.Containers = populatedContainers

	return &node, nil
}

// ListNode lists all nodes in the cluster
func (sm *StateManager) ListNodes() ([]Node, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := sm.etcdClient.Get(ctx, "/nodes/", clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	nodes := make([]Node, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var node Node
		if err := json.Unmarshal(kv.Value, &node); err != nil {
			// Handle or log the error
			continue
		}

		populatedContainers := make([]Container, 0, len(node.Containers))
		for _, container := range node.Containers {
			container, err := sm.GetContainer(container.NamespaceID, container.ID)
			if err != nil {
				fmt.Printf("Failed to populate container for node: %v", err)
				continue
			}
			populatedContainers = append(populatedContainers, *container)
		}
		node.Containers = populatedContainers

		nodes = append(nodes, node)
	}

	return nodes, nil
}

func (sm *StateManager) AssignContainerToNode(namespaceID, containerID, nodeID string) error {
	// Fetch the node
	node, err := sm.GetNode(nodeID)
	if err != nil {
		return err
	}

	container, err := sm.GetContainer(namespaceID, containerID)

	if err != nil {
		return err
	}

	// Update the container's NodeID
	containerPatch := ContainerPatch{NodeID: &nodeID}
	if err := sm.PatchContainer(namespaceID, containerID, containerPatch); err != nil {
		return err
	}

	// Add the container to the node's list of containers
	node.Containers = append(node.Containers, Container{ID: container.ID, NamespaceID: container.NamespaceID}) // Other data is fetched in getNodes/listNodes
	// Save the updated node back to etcd
	return sm.etcdClient.SaveEntity(node)
}
