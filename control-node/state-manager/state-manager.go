package statemanager

import (
	"0xKowalski1/container-orchestrator/config"
	"0xKowalski1/container-orchestrator/models"
	"fmt"
	"sync"
)

type StateManager struct {
	etcdClient    *EtcdClient
	listeners     []Listener
	cfg           *config.Config
	subscriptions map[string]map[chan string]struct{} // ContainerID -> Set of Subscriber Channels
	mu            sync.Mutex
}

func NewStateManager(cfg *config.Config) (*StateManager, error) {
	cli, err := NewEtcdClient()
	if err != nil {
		fmt.Printf("Failed etcd setup: %v", err)
		return nil, err
	}

	stateManager := &StateManager{
		etcdClient:    cli,
		listeners:     []Listener{},
		cfg:           cfg,
		subscriptions: make(map[string]map[chan string]struct{}),
	}

	// Check configured namespace is the only namespace that exists, create it if no namespaces exist.
	namespaces, err := stateManager.ListNamespaces()
	if err != nil {
		return nil, err
	}

	switch len(namespaces) {
	case 0:
		if err := stateManager.AddNamespace(models.Namespace{ID: cfg.Namespace}); err != nil {
			fmt.Println("Error creating namespace:", err)
			return nil, err
		}
	case 1:
		if namespaces[0].ID != cfg.Namespace {
			return nil, fmt.Errorf("Existing namespace does not match configured: %s != %s", namespaces[0].ID, cfg.Namespace)
		}
	default:
		return nil, fmt.Errorf("etcd has multiple namespaces, there should only be one, panic!")
	}

	return stateManager, nil

}

func (sm *StateManager) Close() error {
	return sm.etcdClient.Close()
}
