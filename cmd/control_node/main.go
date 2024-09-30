package main

import (
	"context"
	"fmt"
	"log"
	"strings"
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
	etcdClient, err := clientv3.New(clientv3.Config{
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
	app.Use(compress.New())

	/// Services
	containerService := controlnode.NewContainerService(config, etcdClient)
	nodeService := controlnode.NewNodeService(config, etcdClient)

	/// Handlers
	containerHandler := controlnode.NewContainerHandler(containerService)
	nodeHandler := controlnode.NewNodeHandler(nodeService)

	/// Api Routes
	//// /containers
	app.Get("/api/containers", containerHandler.GetContainers)
	app.Post("/api/containers", containerHandler.CreateContainer)
	app.Patch("/api/containers/:id", containerHandler.UpdateContainer)
	//// /nodes
	app.Get("/api/nodes", nodeHandler.GetNodes)
	app.Get("/api/nodes/:id", nodeHandler.GetNode)
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

	// Schedular
	schedular := controlnode.NewSchedular(containerService, nodeService)

	schedular.ScheduleContainers() // Initial run incase events where missed while offline

	go func() {
		watchChannel := etcdClient.Watch(context.Background(), fmt.Sprintf("/%s", config.EtcdNamespace), clientv3.WithPrefix())
		for watchResp := range watchChannel {
			for _, ev := range watchResp.Events {
				if strings.HasPrefix(string(ev.Kv.Key), fmt.Sprintf("/%s/containers", config.EtcdNamespace)) {
					switch ev.Type {
					// Schedule containers when a new container is added, as it will be unscheduled
					// And when a container is deleted, incase containers are waiting on space to be free.
					case clientv3.EventTypePut:
						schedular.ScheduleContainers()
					case clientv3.EventTypeDelete:
						schedular.ScheduleContainers()
						// TODO: Remove container from node
					}
				} else if strings.HasPrefix(string(ev.Kv.Key), fmt.Sprintf("/%s/nodes", config.EtcdNamespace)) {
					// Schedule containers if a new node is added, incase unscheduled containers are waiting if nodes are all full
					switch ev.Type {
					case clientv3.EventTypePut:
						schedular.ScheduleContainers()
					}
				}
			}
		}
	}()

	log.Fatal(app.Listen(":3000")) // TODO: Get this from config
}
