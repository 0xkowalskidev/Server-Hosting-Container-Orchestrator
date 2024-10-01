package main

import (
	"log"

	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/utils"
	workernode "github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/worker_node"
	"github.com/go-resty/resty/v2"
)

func main() {
	var config workernode.Config
	utils.ParseConfigFromEnv(&config)

	runtime, err := workernode.NewContainerdRuntime(config)
	if err != nil {
		log.Fatalf("Failed to initialize runtime: %v", err)
	}

	storageManager := workernode.NewStorageManager(config, &utils.FileOps{})

	networkManager, err := workernode.NewNetworkManager(config, &utils.FileOps{})
	if err != nil {
		log.Fatalf("Failed to initialize network manager: %v", err)
	}

	client := resty.New()
	agent := workernode.NewAgent(config, client, runtime, storageManager, networkManager)

	agent.StartAgent()
}
