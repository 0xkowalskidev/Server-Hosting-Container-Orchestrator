package main

import (
	"context"
	"log"

	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/utils"
	workernode "github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/worker_node"
)

func main() {
	var config workernode.Config
	utils.ParseConfigFromEnv(&config)

	runtime, err := workernode.NewRuntime(config)

	if err != nil {
		log.Fatalf("Failed to initialize runtime: %v", err)
	}

	ctx := context.Background()
	_, err = runtime.CreateContainer(ctx, "test", "gameservers", "ghcr.io/0xKowalskiDev/minecraft-server:latest")

	if err != nil {
		log.Fatalf("Failed to create container: %v", err)
	}

	_, err = runtime.StartContainer(ctx, "test", "gameservers")

	if err != nil {
		log.Fatalf("Failed to start container: %v", err)
	}

}
