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

func (ns *NodeService) GetNode(nodeID string) (models.Node, error) {
	var node models.Node

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := ns.etcdClient.Get(ctx, fmt.Sprintf("/%s/nodes/%s", ns.config.EtcdNamespace, nodeID))
	if err != nil {
		return node, fmt.Errorf("Failed to get node with id %s from etcd: %v", nodeID, err)
	}

	if len(resp.Kvs) == 0 {
		return node, nil // Empty node
	}

	if err := json.Unmarshal(resp.Kvs[0].Value, &node); err != nil {
		return node, fmt.Errorf("Failed to decode node data from etcd: %v", err)
	}

	return node, nil
}

func (ns *NodeService) GetNodes() ([]models.Node, error) {
	nodes := make([]models.Node, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := ns.etcdClient.Get(ctx, fmt.Sprintf("/%s/nodes", ns.config.EtcdNamespace), clientv3.WithPrefix())
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

func (ns *NodeService) PutNode(node models.Node) error {
	nodeData, err := json.Marshal(node)
	if err != nil {
		return fmt.Errorf("Failed to serialize node: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = ns.etcdClient.Put(ctx, fmt.Sprintf("/%s/nodes/%s", ns.config.EtcdNamespace, node.ID), string(nodeData))
	if err != nil {
		return fmt.Errorf("Failed to store node data in etcd: %v", err)
	}

	return nil
}
