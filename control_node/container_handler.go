package controlnode

import (
	"html/template"

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
		return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error", "details": err.Error()})
	}

	if c.Get("HX-Request") == "true" {
		tmplStr := `
			<ul id="containers">
				{{range .}}
				<li>Container ID: {{.ID}}</li>
				{{end}}
			</ul>
		`
		engine := template.New("containers")

		t, _ := engine.Parse(tmplStr)

		return t.Execute(c.Response().BodyWriter(), containers)

	} else {
		return c.JSON(containers)
	}

}

func (ch *ContainerHandler) CreateContainer(c fiber.Ctx) error {
	var container models.Container
	if err := c.Bind().Body(&container); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad Request", "details": err.Error()})
	}

	err := ch.containerService.CreateContainer(container)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error", "details": err.Error()})
	}

	if c.Get("HX-Request") == "true" {
		tmplStr := `
				<li>Container ID: {{.ID}}</li>
		`
		engine := template.New("container")

		t, _ := engine.Parse(tmplStr)

		return t.Execute(c.Response().BodyWriter(), container)

	} else {
		return c.Status(201).JSON(container)
	}
}
