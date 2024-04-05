package statemanager

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// AddNamespace adds a new namespace to the cluster
func (sm *StateManager) AddNamespace(namespace Namespace) error {
	return sm.etcdClient.SaveEntity(namespace)
}

// RemoveNamespace removes a namespace from the cluster by its ID
func (sm *StateManager) RemoveNamespace(namespaceID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := "/namespaces/" + namespaceID
	_, err := sm.etcdClient.Client.Delete(ctx, key, clientv3.WithPrefix()) // Use WithPrefix to ensure all contained containers are also deleted
	return err
}

// GetNamespace retrieves a namespace by its ID
func (sm *StateManager) GetNamespace(namespaceID string) (*Namespace, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := "/namespaces/" + namespaceID
	resp, err := sm.etcdClient.Client.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("namespace not found")
	}

	var namespace Namespace
	err = json.Unmarshal(resp.Kvs[0].Value, &namespace)
	if err != nil {
		return nil, err
	}

	return &namespace, nil
}

// ListNamespaces lists all namespaces in the cluster
func (sm *StateManager) ListNamespaces() ([]Namespace, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := sm.etcdClient.Client.Get(ctx, "/namespaces/", clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	namespaces := make([]Namespace, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var namespace Namespace
		if err := json.Unmarshal(kv.Value, &namespace); err != nil {
			continue
		}
		namespaces = append(namespaces, namespace)
	}

	return namespaces, nil
}
