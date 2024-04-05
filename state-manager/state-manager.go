package statemanager

import (
	"fmt"
)

// /nodes
// /namespaces
// /namespaces/{namespace}/containers

type StateManager struct {
	etcdClient *EtcdClient
	listeners  []Listener
}

func Start() (*StateManager, error) {
	cli, err := NewEtcdClient()
	if err != nil {
		fmt.Printf("Failed etcd setup: %v", err)
		return nil, err
	}

	state := &StateManager{etcdClient: cli, listeners: []Listener{}}

	return state, nil
}

func (_statemanager *StateManager) Close() error {
	return _statemanager.etcdClient.Close()
}
