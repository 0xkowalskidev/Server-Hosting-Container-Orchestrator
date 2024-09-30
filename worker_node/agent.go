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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	actualContainers, err := a.runtime.GetContainers(ctx, node.Namespace)
	if err != nil {
		return err
	}

	// Build maps for quick lookup
	actualContainerMap := make(map[string]containerd.Container)
	for _, container := range actualContainers {
		actualContainerMap[container.ID()] = container
	}

	desiredContainerMap := make(map[string]models.Container)
	for _, container := range node.Containers {
		desiredContainerMap[container.ID] = container
	}

	// Delete containers not in the desired state or match state
	for id, container := range actualContainerMap {
		if _, exists := desiredContainerMap[id]; !exists {
			deleteCtx, deleteCancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer deleteCancel()

			// TODO: containers must be stopped to be deleted

			if err := a.runtime.RemoveContainer(deleteCtx, container.ID(), node.Namespace); err != nil {
				log.Printf("Failed to delete container: %v", err)
				continue
			}
		} else {
			a.MatchContainerState(desiredContainerMap[id], container)
		}
	}

	// Create containers in desired state
	for id, desiredContainer := range desiredContainerMap {
		if _, exists := actualContainerMap[id]; !exists {
			createCtx, createCancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer createCancel()

			container, err := a.runtime.CreateContainer(createCtx, desiredContainer.ID, node.Namespace, desiredContainer.Image)
			if err != nil {
				log.Printf("Failed to create container: %v", err)
				continue
			}

			a.MatchContainerState(desiredContainer, container)
		}
	}

	return nil
}

func (a *Agent) MatchContainerState(desiredContainer models.Container, actualContainer containerd.Container) {
	// Match status
}
