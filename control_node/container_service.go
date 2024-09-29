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

func (cs *ContainerService) GetContainers() ([]models.Container, error) {
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
		containers = append(containers, container)
	}

	return containers, nil
}

func (cs *ContainerService) PutContainer(container models.Container) error {
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
