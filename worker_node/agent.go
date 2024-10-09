package workernode

import (
	"context"
	"fmt"
	"log"
	"syscall"
	"time"

	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/models"
	"github.com/containerd/containerd"
	"github.com/go-resty/resty/v2"
)

type Agent struct {
	config         Config
	client         *resty.Client
	runtime        *ContainerdRuntime
	storageManager *StorageManager
	networkManager *NetworkManager
}

func NewAgent(config Config, client *resty.Client, runtime *ContainerdRuntime, storageManager *StorageManager, networkManager *NetworkManager) *Agent {
	return &Agent{config: config, client: client, runtime: runtime, storageManager: storageManager, networkManager: networkManager}
}

func (a *Agent) StartAgent() {
	var node models.Node

	connectTicker := time.NewTicker(2 * time.Second)
	defer connectTicker.Stop()

ConnectLoop:
	for {
		select {
		case <-connectTicker.C:
			resp, err := a.client.R().
				SetResult(&node).
				Get(fmt.Sprintf("%s/nodes/%s", a.config.ControlNodeURI, a.config.NodeID))
			if err != nil {
				log.Printf("Failed to connect to control node endpoint: %v", err)
				continue
			}
			switch resp.StatusCode() {
			case 200:
				connectTicker.Stop()
				break ConnectLoop
			case 404:
				newNode := models.Node{
					ID:        a.config.NodeID,
					Namespace: a.config.ContainerdNamespace,
				}
				_, err := a.client.R().SetBody(newNode).SetResult(&node).Post(fmt.Sprintf("%s/nodes", a.config.ControlNodeURI))
				if err != nil {
					log.Printf("Failed to join cluster: %v", err)
				}

				connectTicker.Stop()
				break ConnectLoop
			default:
				log.Printf("Failed to get node from cluster: %v", err)
			}
		}
	}

	syncTicker := time.NewTicker(2 * time.Second) // TODO: Switch to SSE instead of polling at some point
	defer syncTicker.Stop()
	for range syncTicker.C {
		err := a.SyncNode(node)
		if err != nil {
			log.Printf("Failed to sync node: %v", err)
		}
	}
}

func (a *Agent) SyncNode(node models.Node) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var desiredContainers []models.Container

	_, err := a.client.R().
		SetResult(&desiredContainers).
		SetQueryParam("nodeID", node.ID).
		Get(fmt.Sprintf("%s/containers", a.config.ControlNodeURI))
	if err != nil {
		return fmt.Errorf("Failed to get nodes containers: %v", err)
	}

	a.storageManager.SyncStorage(desiredContainers)
	a.networkManager.SyncNetwork(desiredContainers)

	actualContainers, err := a.runtime.GetContainers(ctx, node.Namespace)
	if err != nil {
		return fmt.Errorf("Failed to get containers from runtime: %v", err)
	}

	// Build maps for quick lookup
	actualContainerMap := make(map[string]containerd.Container)
	for _, container := range actualContainers {
		actualContainerMap[container.ID()] = container
	}

	desiredContainerMap := make(map[string]models.Container)
	for _, container := range desiredContainers {
		desiredContainerMap[container.ID] = container
	}

	// Delete containers not in the desired state or match state
	for id, container := range actualContainerMap {
		if _, exists := desiredContainerMap[id]; !exists {
			ctx := context.Background()
			if err := a.runtime.RemoveContainer(ctx, container.ID(), node.Namespace); err != nil {
				log.Printf("Failed to delete container: %v", err)
				continue
			}
		} else {
			err := a.MatchContainerState(node.Namespace, desiredContainerMap[id], container)
			if err != nil {
				log.Printf("Failed to match container state: %v", err)
			}
		}
	}

	// Create containers in desired state
	for id, desiredContainer := range desiredContainerMap {
		if _, exists := actualContainerMap[id]; !exists {
			ctx := context.Background()

			container, err := a.runtime.CreateContainer(ctx, desiredContainer.ID, node.Namespace, desiredContainer.Image, desiredContainer.MemoryLimit, desiredContainer.CPULimit, desiredContainer.Env)
			if err != nil {
				log.Printf("Failed to create container: %v", err)
				continue
			}

			err = a.MatchContainerState(node.Namespace, desiredContainer, container)
			if err != nil {
				log.Printf("Failed to match container state: %v", err)
			}
		}
	}

	return nil
}

func (a *Agent) MatchContainerState(namespace string, desiredContainer models.Container, actualContainer containerd.Container) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	actualStatus, err := a.runtime.GetContainerStatus(ctx, desiredContainer.ID, namespace)
	if err != nil {
		return fmt.Errorf("Failed to get container status: %v", err)
	}

	if actualStatus != desiredContainer.DesiredStatus {
		switch desiredContainer.DesiredStatus {
		case models.StatusRunning:
			_, err := a.runtime.StartContainer(ctx, desiredContainer.ID, namespace)
			if err != nil {
				return fmt.Errorf("Failed to start container: %v", err)
			}
		case models.StatusStopped:
			err := a.runtime.StopContainer(ctx, desiredContainer.ID, namespace, syscall.SIGKILL)
			if err != nil {
				return fmt.Errorf("Failed to stop container: %v", err)
			}
		}
	}

	return nil
}
