package controlnode

import (
	"fmt"

	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/models"
	"github.com/gofiber/fiber/v3"
)

type ContainerHandler struct {
	containerService *ContainerService
}

func NewContainerHandler(containerService *ContainerService) *ContainerHandler {
	return &ContainerHandler{containerService: containerService}
}

func (ch *ContainerHandler) GetContainers(c fiber.Ctx) error {
	nodeID := c.Query("nodeID")

	containers, err := ch.containerService.GetContainers(nodeID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error", "details": err.Error()})
	}

	return c.JSON(containers)
}

func (ch *ContainerHandler) CreateContainer(c fiber.Ctx) error {
	var container models.Container

	if err := c.Bind().Body(&container); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad Request", "details": err.Error()})
	}

	err := ch.containerService.PutContainer(container)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error", "details": err.Error()})
	}

	return c.Status(201).JSON(container)
}

func (ch *ContainerHandler) UpdateContainer(c fiber.Ctx) error {
	containerID := c.Params("id")
	container, err := ch.containerService.GetContainer(containerID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error", "details": err.Error()})
	}

	if container.ID == "" {
		return c.Status(404).JSON(fiber.Map{"error": "Resource Not Found", "details": fmt.Sprintf("Container with ID=%s not found.", containerID)})
	}

	var patchContainer models.Container
	if err := c.Bind().Body(&patchContainer); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad Request", "details": err.Error()})
	}

	err = container.Patch(&patchContainer)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error", "details": err.Error()})
	}

	err = ch.containerService.PutContainer(container)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error", "details": err.Error()})
	}

	return c.Status(201).JSON(container)
}

func (ch *ContainerHandler) DeleteContainer(c fiber.Ctx) error {
	containerID := c.Params("id")
	container, err := ch.containerService.GetContainer(containerID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error", "details": err.Error()})
	}

	if container.ID == "" {
		return c.Status(404).JSON(fiber.Map{"error": "Resource Not Found", "details": fmt.Sprintf("Container with ID=%s not found.", containerID)})
	}

	err = ch.containerService.DeleteContainer(containerID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error", "details": err.Error()})
	}

	return c.Status(200).Send(nil)
}
