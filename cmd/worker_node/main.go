package main

import (
	"context"
	"log"

	"github.com/0xKowalski1/Server-Hosting-Container-Orchestrator/config"
	workernode "github.com/0xKowalski1/Server-Hosting-Container-Orchestrator/worker_node"
)

func main() {
	var cfg config.Config
	config.ParseConfigFromEnv(&cfg)

	runtime, err := workernode.NewRuntime(cfg)

	if err != nil {
		log.Fatalf("Failed to initialize runtime: %v", err)
	}

	ctx := context.Background()
	_, err = runtime.CreateContainer(ctx, "test", cfg.NamespaceMain, "ghcr.io/0xkowalski1/minecraft-server:latest")

	if err != nil {
		log.Fatalf("Failed to create container: %v", err)
	}
}
