package controlnode

import (
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
	containers, err := ch.containerService.GetContainers()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error getting containers", "details": err.Error()})
	}

	return c.JSON(containers)

}

func (ch *ContainerHandler) CreateContainer(c fiber.Ctx) error {
	var container models.Container
	if err := c.Bind().Body(&container); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request data", "details": err.Error()})
	}

	err := ch.containerService.CreateContainer(container)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error creating container", "details": err.Error()})
	}

	return c.Status(201).JSON(container)
}
