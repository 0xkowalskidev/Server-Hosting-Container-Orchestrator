package statemanager

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// EtcdClient wraps the etcd client with additional functionality.
type EtcdClient struct {
	*clientv3.Client
}

func NewEtcdClient() (*EtcdClient, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		fmt.Printf("Failed to connect to etcd: %v", err)
		return nil, err
	}

	return &EtcdClient{Client: cli}, nil
}

// Storable represents an entity that can be saved to etcd.
type Storable interface {
	Key() string            // Generates a unique etcd key for the entity.
	Value() (string, error) // Serializes the entity to a string for storage.
}

// SaveEntity saves any Storable entity to etcd using the client.
func (ec *EtcdClient) SaveEntity(entity Storable) error {
	key := entity.Key()

	valueStr, err := entity.Value()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = ec.Put(ctx, key, valueStr)
	return err
}

func (ec *EtcdClient) Close() error {
	return ec.Client.Close()
}
