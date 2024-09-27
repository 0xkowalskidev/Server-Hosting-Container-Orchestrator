package main

import (
	"log"
	"time"

	controlnode "github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/control_node"
	"github.com/gofiber/fiber/v3"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func main() {
	app := fiber.New()

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		panic(err)
	}

	// Services
	containerService := controlnode.NewContainerService(client)
	nodeService := controlnode.NewNodeService(client)

	// Handlers
	containerHandler := controlnode.NewContainerHandler(containerService)
	nodeHandler := controlnode.NewNodeHandler(nodeService)
	// Routes
	//// /containers
	app.Get("/api/v1/containers", containerHandler.GetContainers)
	app.Post("/api/v1/containers", containerHandler.CreateContainer)

	//// /nodes
	app.Get("/api/v1/nodes", nodeHandler.GetNodes)
	app.Post("/api/v1/nodes", nodeHandler.CreateNode)

	log.Fatal(app.Listen(":3000"))
}
