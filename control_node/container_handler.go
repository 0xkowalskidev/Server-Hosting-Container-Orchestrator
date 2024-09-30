package controlnode

import (
	"fmt"
	"log"

	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/models"
	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/utils"
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
		return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error", "details": err.Error()})
	}

	if c.Get("HX-Request") == "true" {
		return c.Render("containers", fiber.Map{"Containers": containers})
	} else {
		return c.JSON(containers)
	}

}

func (ch *ContainerHandler) CreateContainer(c fiber.Ctx) error {
	var container models.Container

	log.Println(string(c.Body()))

	if err := c.Bind().Body(&container); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad Request", "details": err.Error()})
	}

	err := ch.containerService.PutContainer(container)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error", "details": err.Error()})
	}

	if c.Get("HX-Request") == "true" {
		return c.Render("container", utils.StructToFiberMap(container))
	} else {
		return c.Status(201).JSON(container)
	}
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

	log.Println(string(c.Body()))
	var patchContainer models.Container
	if err := c.Bind().Body(&patchContainer); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad Request", "details": err.Error()})
	}

	log.Println(patchContainer)

	err = container.Patch(&patchContainer)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error", "details": err.Error()})
	}

	err = ch.containerService.PutContainer(container)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error", "details": err.Error()})
	}

	if c.Get("HX-Request") == "true" {
		return c.Render("container", utils.StructToFiberMap(container))
	} else {
		return c.Status(201).JSON(container)
	}

}
