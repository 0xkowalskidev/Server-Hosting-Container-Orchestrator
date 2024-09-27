package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/models"
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

	// Containers
	app.Get("/api/v1/containers", func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := client.Get(ctx, "/containers", clientv3.WithPrefix())
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to retrieve containers from etcd", "details": err.Error()})
		}

		containers := make([]models.Container, 0)
		for _, kv := range resp.Kvs {
			var container models.Container
			if err := json.Unmarshal(kv.Value, &container); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "Failed to decode container data", "details": err.Error()})
			}
			containers = append(containers, container)
		}

		return c.JSON(containers)
	})

	app.Post("/api/v1/containers", func(c fiber.Ctx) error {
		var container models.Container
		if err := c.Bind().Body(&container); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Bad request", "details": err.Error()})
		}

		containerData, err := json.Marshal(container)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Error serializing container", "details": err.Error()})
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		_, err = client.Put(ctx, "/containers/"+container.ID, string(containerData))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to store container data", "details": err.Error()})
		}

		return c.Status(201).JSON(container)
	})

	// Nodes
	app.Get("/api/v1/nodes", func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := client.Get(ctx, "/nodes", clientv3.WithPrefix())
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to retrieve nodes from etcd", "details": err.Error()})
		}

		nodes := make([]models.Node, 0)
		for _, kv := range resp.Kvs {
			var node models.Node
			if err := json.Unmarshal(kv.Value, &node); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "Failed to decode node data", "details": err.Error()})
			}
			nodes = append(nodes, node)
		}

		return c.JSON(nodes)
	})

	app.Post("/api/v1/nodes", func(c fiber.Ctx) error {
		var node models.Node
		if err := c.Bind().Body(&node); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Bad request", "details": err.Error()})
		}

		nodeData, err := json.Marshal(node)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Error serializing node", "details": err.Error()})
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		_, err = client.Put(ctx, "/nodes/"+node.ID, string(nodeData))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to store node data", "details": err.Error()})
		}

		return c.Status(201).JSON(node)
	})

	log.Fatal(app.Listen(":3000"))
}
