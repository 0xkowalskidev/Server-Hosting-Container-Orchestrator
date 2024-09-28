package main

import (
	"log"
	"time"

	controlnode "github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/control_node"
	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/static"
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

	// HTTP Server
	app := fiber.New()

	app.Use(compress.New()) // Enable gzip compression

	/// Services
	containerService := controlnode.NewContainerService(config, client)
	nodeService := controlnode.NewNodeService(config, client)

	/// Handlers
	containerHandler := controlnode.NewContainerHandler(containerService)
	nodeHandler := controlnode.NewNodeHandler(nodeService)

	/// Api Routes
	//// /containers
	app.Get("/api/containers", containerHandler.GetContainers)
	app.Post("/api/containers", containerHandler.CreateContainer)
	//// /nodes
	app.Get("/api/nodes", nodeHandler.GetNodes)
	app.Post("/api/nodes", nodeHandler.CreateNode)

	/// Control Panel Routes
	//// Static Files
	app.Get("/static*", static.New("./control_node/control_panel/static"))

	//// Routes
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendFile("./control_node/index.html")
	})

	log.Fatal(app.Listen(":3000")) // TODO: Get this from config
}
