package controlnode

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/models"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type NodeService struct {
	config     Config
	etcdClient *clientv3.Client
}

func NewNodeService(config Config, etcdClient *clientv3.Client) *NodeService {
	return &NodeService{
		config:     config,
		etcdClient: etcdClient,
	}
}

func (ns *NodeService) GetNodes() ([]models.Node, error) {
	nodes := make([]models.Node, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := ns.etcdClient.Get(ctx, fmt.Sprintf("/%s/nodes", ns.config.Namespace), clientv3.WithPrefix())
	if err != nil {
		return nodes, fmt.Errorf("Failed to get nodes from etcd: %v", err)
	}

	for _, kv := range resp.Kvs {
		var node models.Node
		if err := json.Unmarshal(kv.Value, &node); err != nil {
			return nodes, fmt.Errorf("Failed to decode node data from etcd: %v", err)
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

func (ns *NodeService) CreateNode(node models.Node) error {
	nodeData, err := json.Marshal(node)
	if err != nil {
		return fmt.Errorf("Failed to serialize node: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = ns.etcdClient.Put(ctx, fmt.Sprintf("/%s/nodes/%s", ns.config.Namespace, node.ID), string(nodeData))
	if err != nil {
		return fmt.Errorf("Failed to store node data in etcd: %v", err)
	}

	return nil
}
