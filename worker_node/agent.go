package workernode

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/models"
	"github.com/containerd/containerd"
	"github.com/go-resty/resty/v2"
)

type Agent struct {
	config  Config
	client  *resty.Client
	runtime *ContainerdRuntime
}

func NewAgent(config Config, client *resty.Client, runtime *ContainerdRuntime) *Agent {
	return &Agent{config: config, client: client, runtime: runtime}
}

func (a *Agent) StartAgent() {

	ticker := time.NewTicker(5 * time.Second) // TODO: Switch to SSE instead of polling at some point
	defer ticker.Stop()
	for range ticker.C {
		var node models.Node
		resp, err := a.client.R().
			SetResult(&node).
			Get(fmt.Sprintf("%s/%s", fmt.Sprintf("%s/nodes", a.config.ControlNodeURI), a.config.NodeID))
		if err != nil {
			log.Printf("Failed to connect to control node endpoint: %v", err)
			continue
		}
		switch resp.StatusCode() {
		case 200:
			err := a.SyncNode(node)
			if err != nil {
				log.Printf("Failed to sync node: %v", err)
			}
		case 404:
			err := a.JoinCluster()
			if err != nil {
				log.Printf("Failed to join cluster: %v", err)
			}
		default:
			log.Printf("Failed to get node from cluster: %v", err)
		}
	}
}

func (a *Agent) JoinCluster() error {
	newNode := models.Node{
		ID:        a.config.NodeID,
		Namespace: a.config.ContainerdNamespace,
	}
	_, err := a.client.R().SetBody(newNode).Post(fmt.Sprintf("%s/nodes", a.config.ControlNodeURI))
	if err != nil {
		return err
	}

	return nil
}

func (a *Agent) SyncNode(node models.Node) error {
	// Sync with control node state
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	containerdContainers, err := a.runtime.GetContainers(ctx, node.Namespace)
	if err != nil {
		return err
	}
	containerdContainersMap := make(map[string]containerd.Container)
	for _, containerdContainer := range containerdContainers {
		containerdContainersMap[containerdContainer.ID()] = containerdContainer
	}
	// TODO: Loop over all containers, check state, make map of uncreated containers, create them
	for _, desiredContainer := range node.Containers {
		if containerdContainersMap[desiredContainer.ID] == nil {
			// Create container
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) // TODO: not sure how long to make this
			defer cancel()
			_, err := a.runtime.CreateContainer(ctx, desiredContainer.ID, node.Namespace, desiredContainer.Image)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *Agent) MatchContainerState(desiredContainer models.Container, actualContainer containerd.Container) {

}
