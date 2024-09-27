package main

import (
	"log"
	"time"

	controlnode "github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/control_node"
	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/utils"
	"github.com/gofiber/fiber/v3"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func main() {
	// Config
	var config controlnode.Config
	utils.ParseConfigFromEnv(&config)

	// Etcd
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"}, // TODO: Get these from config
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to connect to Etcd: %v", err)
	}

	// Api
	app := fiber.New()

	/// Services
	containerService := controlnode.NewContainerService(config, client)
	nodeService := controlnode.NewNodeService(config, client)

	/// Handlers
	containerHandler := controlnode.NewContainerHandler(containerService)
	nodeHandler := controlnode.NewNodeHandler(nodeService)

	/// Routes
	//// /containers
	app.Get("/api/v1/containers", containerHandler.GetContainers)
	app.Post("/api/v1/containers", containerHandler.CreateContainer)
	//// /nodes
	app.Get("/api/v1/nodes", nodeHandler.GetNodes)
	app.Post("/api/v1/nodes", nodeHandler.CreateNode)

	log.Fatal(app.Listen(":3000")) // TODO: Get this from config
}
