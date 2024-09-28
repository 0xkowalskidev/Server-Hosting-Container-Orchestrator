package controlnode

import (
	"fmt"

	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/models"
	"github.com/gofiber/fiber/v3"
)

type NodeHandler struct {
	nodeService *NodeService
}

func NewNodeHandler(nodeService *NodeService) *NodeHandler {
	return &NodeHandler{nodeService: nodeService}
}

func (nh *NodeHandler) GetNode(c fiber.Ctx) error {
	nodeID := c.Params("id")
	node, err := nh.nodeService.GetNode(nodeID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error", "details": err.Error()})
	}

	if node.ID == "" {
		return c.Status(404).JSON(fiber.Map{"error": "Resource Not Found", "details": fmt.Sprintf("Node with ID=%s not found.", nodeID)})
	}

	return c.JSON(node)
}

func (nh *NodeHandler) GetNodes(c fiber.Ctx) error {
	nodes, err := nh.nodeService.GetNodes()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error", "details": err.Error()})
	}

	return c.JSON(nodes)
}

func (nh *NodeHandler) CreateNode(c fiber.Ctx) error {
	var node models.Node
	if err := c.Bind().Body(&node); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad Request", "details": err.Error()})
	}

	err := nh.nodeService.CreateNode(node)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error", "details": err.Error()})
	}

	return c.Status(201).JSON(node)
}
