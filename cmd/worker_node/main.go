package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/models"
	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/utils"
	workernode "github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/worker_node"
	"github.com/containerd/containerd"
	"github.com/go-resty/resty/v2"
)

func main() {
	var config workernode.Config
	utils.ParseConfigFromEnv(&config)

	runtime, err := workernode.NewRuntime(config)
	if err != nil {
		log.Fatalf("Failed to initialize runtime: %v", err)
	}

	// Agent
	client := resty.New()

	ticker := time.NewTicker(5 * time.Second) // TODO: Switch to SSE instead of polling at some point
	defer ticker.Stop()
	for range ticker.C {
		nodesApiEndpoint := fmt.Sprintf("%s/nodes", config.ControlNodeURI)
		var node models.Node
		resp, err := client.R().
			SetResult(&node).
			Get(fmt.Sprintf("%s/%s", nodesApiEndpoint, config.NodeID))
		if err != nil {
			log.Printf("Failed to connect to control node endpoint: %v", err)
			continue
		}
		switch resp.StatusCode() {
		case 200:
			// Sync with control node state
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			containerdContainers, err := runtime.GetContainers(ctx, node.Namespace)
			if err != nil {
				log.Printf("Failed to get actualContainers in agent: %v", err)
				continue
			}
			containerdContainersMap := make(map[string]containerd.Container)
			for _, containerdContainer := range containerdContainers {
				containerdContainersMap[containerdContainer.ID()] = containerdContainer
			}

			for _, desiredContainer := range node.Containers {
				if containerdContainersMap[desiredContainer.ID] == nil {
					// Create container
					ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) // TODO: not sure how long to make this
					defer cancel()
					_, err := runtime.CreateContainer(ctx, desiredContainer.ID, node.Namespace, desiredContainer.Image)
					if err != nil {
						log.Printf("Failed to create container in containerd: %v", err)
						continue
					}
				}
			}

		case 404:
			// Join the cluster
			newNode := models.Node{
				ID:        config.NodeID,
				Namespace: config.ContainerdNamespace,
			}
			_, err := client.R().SetBody(newNode).SetResult(&node).Post(nodesApiEndpoint)
			if err != nil {
				log.Printf("Failed to join cluster: %v", err)
				continue
			}
		default:
			log.Printf("Failed to get node from cluster: %v", err)
		}
	}
}
