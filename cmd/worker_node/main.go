package main

import (
	"fmt"
	"log"
	"time"

	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/models"
	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/utils"
	workernode "github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/worker_node"
	"github.com/go-resty/resty/v2"
)

func main() {
	var config workernode.Config
	utils.ParseConfigFromEnv(&config)

	_, err := workernode.NewRuntime(config)
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
		}
		switch resp.StatusCode() {
		case 200:
			// Subscribe To cluster and do agent stuff
			log.Println(node.ID)
		case 404:
			newNode := models.Node{
				ID: config.NodeID,
			}
			_, err := client.R().SetBody(newNode).SetResult(&node).Post(nodesApiEndpoint)
			if err != nil {
				log.Printf("Failed to join cluster: %v", err)
			}
		default:
			log.Printf("Failed to get node from cluster: %v", err)
		}
	}
}
