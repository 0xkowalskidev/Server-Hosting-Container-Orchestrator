package controlnode

import (
	"context"
	"fmt"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdClient struct {
	*clientv3.Client
	listeners     []Listener
	subscriptions map[string]map[chan string]struct{} // ContainerID -> Set of Subscriber Channels
	mu            sync.Mutex
}

func NewEtcdClient() (*EtcdClient, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		return nil, fmt.Errorf("Failed to connect to etcd: %v", err)
	}

	return &EtcdClient{Client: client}, nil
}

// All Storable models will implement this interface
type Storable interface {
	Key() string            // Generates a unique etcd key for the entity.
	Value() (string, error) // Serializes the entity to a string for storage.
}

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

// Events
type EventType string

const (
	NodeAdded        EventType = "NodeAdded"
	NodeRemoved      EventType = "NodeRemoved"
	NamespaceAdded   EventType = "NamespaceAdded"
	NamespaceRemoved EventType = "NamespaceRemoved"
	ContainerAdded   EventType = "ContainerAdded"
	ContainerRemoved EventType = "ContainerRemoved"
)

type Event struct {
	Type EventType
	Data interface{}
}

type Listener func(Event)

func (ec *EtcdClient) Subscribe(listener Listener) {
	ec.listeners = append(ec.listeners, listener)
}

// emit broadcasts an event to all subscribers.
func (ec *EtcdClient) emit(event Event) {
	for _, listener := range ec.listeners {
		listener(event)
	}
}
