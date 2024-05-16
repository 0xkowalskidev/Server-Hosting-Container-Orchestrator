package statemanager

import (
	"0xKowalski1/container-orchestrator/models"
	"context"
	"encoding/json"
	"log"
	"strings"

	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// AddNamespace adds a new namespace to the cluster
func (sm *StateManager) AddNamespace(namespace models.Namespace) error {
	return sm.etcdClient.SaveEntity(namespace)
}

// ListNamespaces lists all namespaces in the cluster
func (sm *StateManager) ListNamespaces() ([]models.Namespace, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := sm.etcdClient.Client.Get(ctx, "/namespaces/", clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	namespaces := make([]models.Namespace, 0)
	for _, kv := range resp.Kvs {
		key := string(kv.Key)

		segments := strings.Split(key, "/")
		// Direct children of "/namespaces/" have 3 segments
		if len(segments) == 3 {
			var ns models.Namespace
			if err := json.Unmarshal(kv.Value, &ns); err != nil {
				log.Printf("Error unmarshalling namespace value for key %s: %v", key, err)
				return nil, err
			}
			ns.ID = segments[2]
			namespaces = append(namespaces, ns)
		}
	}

	return namespaces, nil
}
