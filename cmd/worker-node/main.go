package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"0xKowalski1/container-orchestrator/api-wrapper"
	"0xKowalski1/container-orchestrator/config"
	"0xKowalski1/container-orchestrator/models"
	"0xKowalski1/container-orchestrator/utils"
	workernode "0xKowalski1/container-orchestrator/worker-node"
)

type ApiResponse struct {
	Node struct {
		ID         string             `json:"ID"`
		Containers []models.Container `json:"Containers"`
	} `json:"node"`
}

func main() {
	cfgPath := "config.json"
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	runtime, err := workernode.NewContainerdRuntime(cfg)

	if err != nil {
		fmt.Printf("Error initializing runtime: %v\n", err)
		os.Exit(1)
	}

	storage := workernode.NewStorageManager(cfg, &utils.FileOps{}, &utils.CmdRunner{})

	networking := workernode.NewNetworkingManager(cfg, &utils.CmdRunner{})

	metricsAndLogsApi := workernode.NewMetricsAndLogsApi(cfg)

	go metricsAndLogsApi.Start()

	apiClient := api.NewApiWrapper(cfg.ControlNodeIp)

	// Should do self discovery/cfg for this
	nodeConfig := models.CreateNodeRequest{
		ID:           "node-1",
		MemoryLimit:  16,
		CpuLimit:     4,
		StorageLimit: 10,
		NodeIp:       cfg.NodeIp,
	}

	// Check if node exists
	_, err = apiClient.GetNode(nodeConfig.ID)
	if err != nil {
		// If it does not (or not authed, which will fail) try and Join Cluster
		_, err := apiClient.JoinCluster(nodeConfig)
		if err != nil {
			log.Printf("Error joining cluster: %v", err)
			panic(err) // Probably want to retry & log
		}
	}

	// If node already exists and it isnt use, then auth should catch it

	ticker := time.NewTicker(5 * time.Second) // Switch to SSE instead of polling at some point
	defer ticker.Stop()

	for range ticker.C {
		node, err := apiClient.GetNode(nodeConfig.ID)
		if err != nil {
			log.Printf("Error checking for nodes desired state: %v", err)
			continue
		}

		err = storage.SyncStorage(node.Containers)
		if err != nil {
			log.Printf("Error syncing storage: %v", err)
			continue
		}

		err = networking.SyncNetworking(node.Containers)
		if err != nil {
			log.Printf("Error syncing network: %v", err)
			continue
		}

		err = runtime.SyncContainers(node.Containers)
		if err != nil {
			log.Printf("Error syncing containers: %v", err)
			continue
		}

	}
}
