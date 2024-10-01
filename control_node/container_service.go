package controlnode

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/models"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type ContainerService struct {
	config     Config
	etcdClient *clientv3.Client
}

func NewContainerService(config Config, etcdClient *clientv3.Client) *ContainerService {
	return &ContainerService{
		config:     config,
		etcdClient: etcdClient,
	}
}

func (cs *ContainerService) GetContainer(containerID string) (models.Container, error) {
	var container models.Container

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := cs.etcdClient.Get(ctx, fmt.Sprintf("/%s/containers/%s", cs.config.EtcdNamespace, containerID))
	if err != nil {
		return container, fmt.Errorf("Failed to get container with id %s from etcd: %v", containerID, err)
	}

	if len(resp.Kvs) == 0 {
		return container, nil // Empty Container
	}

	if err := json.Unmarshal(resp.Kvs[0].Value, &container); err != nil {
		return container, fmt.Errorf("Failed to decode container data from etcd: %v", err)
	}

	return container, nil
}

func (cs *ContainerService) GetContainers(nodeID string) ([]models.Container, error) {
	containers := make([]models.Container, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := cs.etcdClient.Get(ctx, fmt.Sprintf("/%s/containers", cs.config.EtcdNamespace), clientv3.WithPrefix())
	if err != nil {
		return containers, fmt.Errorf("Failed to get containers from etcd: %v", err)
	}

	for _, kv := range resp.Kvs {
		var container models.Container
		if err := json.Unmarshal(kv.Value, &container); err != nil {
			return containers, fmt.Errorf("Failed to decode container data from etcd: %v", err)
		}

		// Filter containers by nodeID if provided
		if nodeID == "" || container.NodeID == nodeID {
			containers = append(containers, container)
		}
	}

	return containers, nil
}

func (cs *ContainerService) PutContainer(container models.Container) error {
	container.SetDefaults()
	// TODO: Validate here

	containerData, err := json.Marshal(container)
	if err != nil {
		return fmt.Errorf("Failed to serialize container: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = cs.etcdClient.Put(ctx, fmt.Sprintf("/%s/containers/%s", cs.config.EtcdNamespace, container.ID), string(containerData))
	if err != nil {
		return fmt.Errorf("Failed to store container data in etcd: %v", err)
	}

	return nil
}

func (cs *ContainerService) DeleteContainer(containerID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := cs.etcdClient.Delete(ctx, fmt.Sprintf("/%s/containers/%s", cs.config.EtcdNamespace, containerID))
	if err != nil {
		return fmt.Errorf("Failed to delete container with id %s from etcd: %v", containerID, err)
	}

	return nil
}
