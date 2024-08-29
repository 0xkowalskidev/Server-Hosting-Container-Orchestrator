package main

import (
	"context"
	"log"

	workernode "github.com/0xKowalski1/Server-Hosting-Container-Orchestrator/worker_node"
)

func main() {
	runtime, err := workernode.NewRuntime("containerd")

	if err != nil {
		log.Fatalf("Failed to initialize runtime: %v", err)
	}

	ctx := context.Background()
	_, err = runtime.CreateContainer(ctx, "test", "ghcr.io/0xkowalski1/minecraft-server:latest")

	if err != nil {
		log.Fatalf("Failed to create container: %v", err)
	}
}
