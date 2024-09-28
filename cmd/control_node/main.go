package main

import (
	"log"
	"time"

	controlnode "github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/control_node"
	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/gofiber/template/html/v2"
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
	engine := html.New("./control_node/templates", ".html") // Template engine
	app := fiber.New(fiber.Config{
		Views: engine,
	})
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
	app.Use("/static*", static.New("./control_node/static", static.Config{
		CacheDuration: 60 * time.Second, // Cache file handlers for 1 minute
		MaxAge:        86400,            // Cache files on the client for 1 day
		Compress:      true,             // Compress and cache static files
	}))

	//// Routes
	app.Get("/", func(c fiber.Ctx) error {
		containers, err := containerService.GetContainers()
		if err != nil {
			// TODO: Do something else here
			return c.Status(500).JSON(fiber.Map{"error": "Error getting containers", "details": err.Error()})
		}

		if c.Get("X-Partial-Content") == "true" {
			return c.Render("containers_page", fiber.Map{"Containers": containers})
		} else {
			return c.Render("containers_page", fiber.Map{"Containers": containers}, "layout")
		}
	})

	log.Fatal(app.Listen(":3000")) // TODO: Get this from config
}
