package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"time"

	controlpanel "github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/control_panel"
	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/models"
	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/gofiber/template/html/v2"
)

func main() {
	// Config
	//var config controlnode.Config
	//utils.ParseConfigFromEnv(&config)

	wrapper := controlpanel.NewWrapper("http://localhost:3001/api") // Get me from config

	// HTTP Server
	engine := html.New("./control_panel/templates", ".html") // Template engine
	app := fiber.New(fiber.Config{
		Views: engine,
	})
	app.Use(compress.New())

	/// Control Panel Routes
	//// Static Files
	app.Use("/static*", static.New("./control_panel/static", static.Config{
		CacheDuration: 60 * time.Second, // Cache file handlers for 1 minute
		MaxAge:        86400,            // Cache files on the client for 1 day
		Compress:      true,             // Compress and cache static files
	}))

	//// Routes
	app.Get("/", func(c fiber.Ctx) error {
		if c.Get("HX-Request") == "true" {
			return c.Render("home_page", nil)
		} else {
			return c.Render("home_page", nil, "layout")
		}
	})

	app.Get("/gameservers", func(c fiber.Ctx) error {
		containers, err := wrapper.GetContainers()
		if err != nil {
			// TODO: Do something else here
			return c.Status(500).JSON(fiber.Map{"error": "Error getting containers", "details": err.Error()})
		}

		if c.Get("HX-Request") == "true" {
			return c.Render("gameservers_page", fiber.Map{"Gameservers": containers})
		} else {
			return c.Render("gameservers_page", fiber.Map{"Gameservers": containers}, "layout")
		}
	})

	app.Get("/gameservers/:id", func(c fiber.Ctx) error {
		containerID := c.Params("id")

		if containerID == "" { // TODO: do something else
			return c.Status(404).JSON(fiber.Map{"error": "Resource Not Found", "details": fmt.Sprintf("Container with ID=%s not found.", containerID)})
		}

		container, err := wrapper.GetContainer(containerID)
		if err != nil {
			// TODO: Do something else here
			return c.Status(500).JSON(fiber.Map{"error": "Error getting containers", "details": err.Error()})
		}

		if c.Get("HX-Request") == "true" {
			return c.Render("gameserver_page", utils.StructToFiberMap(container))
		} else {
			return c.Render("gameserver_page", utils.StructToFiberMap(container), "layout")
		}
	})

	app.Get("/gameservers/:id/console", func(c fiber.Ctx) error {
		containerID := c.Params("id")

		if containerID == "" { // TODO: do something else
			return c.Status(404).JSON(fiber.Map{"error": "Resource Not Found", "details": fmt.Sprintf("Container with ID=%s not found.", containerID)})
		}

		if c.Get("HX-Request") == "true" {
			return c.Render("gameserver_console", fiber.Map{"ID": containerID})
		} else {
			return c.Render("gameserver_console", fiber.Map{"ID": containerID}, "layout") // TODO this wont work with gameserver_page layout
		}

	})

	app.Post("/gameservers/:id/console", func(c fiber.Ctx) error {
		containerID := c.Params("id")

		if containerID == "" { // TODO: do something else
			return c.Status(404).JSON(fiber.Map{"error": "Resource Not Found", "details": fmt.Sprintf("Container with ID=%s not found.", containerID)})
		}

		command := c.FormValue("rcon_command")
		if command == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Command cannot be empty"})
		}

		rconAPI := fmt.Sprintf("http://localhost:3002/rcon/%s", containerID) // TODO: TEMP

		req, err := http.NewRequest("POST", rconAPI, bytes.NewBufferString(command))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create request"})
		}
		req.Header.Set("Content-Type", "text/plain")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to send request"})
		}
		defer resp.Body.Close()

		return nil
	})

	// TODO All these worker node api calls should proxy through the control node
	app.Get("/gameservers/:id/logs", func(c fiber.Ctx) error {
		containerID := c.Params("id")

		if containerID == "" { // TODO: do something else
			return c.Status(404).JSON(fiber.Map{"error": "Resource Not Found", "details": fmt.Sprintf("Container with ID=%s not found.", containerID)})
		}

		nodeLogsAPI := fmt.Sprintf("http://localhost:3002/logs/%s", containerID) // TODO: TEMP

		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")

		reader, writer := io.Pipe()

		go func() {
			defer writer.Close()

			resp, err := http.Get(nodeLogsAPI)
			if err != nil {
				log.Printf("Error fetching logs from node API: %v", err)
				writer.CloseWithError(err)
				return
			}
			defer resp.Body.Close()

			buf := make([]byte, 1024) // Create a buffer to read data in chunks
			for {
				n, err := resp.Body.Read(buf)
				if err != nil {
					if err != io.EOF {
						log.Printf("Error reading from response body: %v", err)
						writer.CloseWithError(err)
						return
					}
					break
				}
				if _, err := writer.Write(buf[:n]); err != nil {
					log.Printf("Error writing to pipe: %v", err)
					writer.CloseWithError(err)
					return
				}
			}

		}()

		return c.SendStream(reader) // TODO probably want to handle not hx requests
	})

	app.Get("/gameservers/:id/metrics", func(c fiber.Ctx) error {
		containerID := c.Params("id")

		if containerID == "" { // TODO: do something else
			return c.Status(404).JSON(fiber.Map{"error": "Resource Not Found", "details": fmt.Sprintf("Container with ID=%s not found.", containerID)})
		}

		containerMetricsEndpoint := fmt.Sprintf("http://localhost:3002/metrics/%s", containerID) // TODO: TEMP

		resp, err := http.Get(containerMetricsEndpoint)
		if err != nil {
			// TODO: Handle this
			log.Printf("Error fetching logs from node API: %v", err)
			return c.Status(500).Send(nil)
		}
		defer resp.Body.Close()

		// TODO: Use resty or something for this, maybe put in wrapper
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response body: %v", err)
			return c.Status(500).Send(nil)
		}

		var metrics models.Metrics

		err = json.Unmarshal(body, &metrics)
		if err != nil {
			log.Printf("Error parsing JSON: %v", err)
			return c.Status(500).Send(nil)
		}

		// Prepare the metrics
		metrics.MemoryUsage = math.Min(math.Floor((metrics.MemoryUsage*10))/10, metrics.MemoryLimit)
		metrics.CPUUsage = math.Min(math.Floor(metrics.CPUUsage), 100)

		return c.Render("gameserver_metrics", utils.StructToFiberMap(metrics))
	})

	app.Get("/gameservers/:id/status", func(c fiber.Ctx) error {
		containerID := c.Params("id")

		if containerID == "" { // TODO: do something else
			return c.Status(404).JSON(fiber.Map{"error": "Resource Not Found", "details": fmt.Sprintf("Container with ID=%s not found.", containerID)})
		}

		containerStatusEndpoint := fmt.Sprintf("http://localhost:3002/status/%s", containerID) // TODO: TEMP

		resp, err := http.Get(containerStatusEndpoint)
		if err != nil {
			// TODO: Handle this
			log.Printf("Error fetching logs from node API: %v", err)
			return c.Status(500).Send(nil)
		}
		defer resp.Body.Close()

		// TODO: Use resty or something for this, maybe put in wrapper
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response body: %v", err)
			return c.Status(500).Send(nil)
		}

		type Response struct {
			Status string
		}
		var response Response

		err = json.Unmarshal(body, &response)
		if err != nil {
			log.Printf("Error parsing JSON: %v", err)
			return c.Status(500).Send(nil)
		}

		return c.Render("gameserver_status", utils.StructToFiberMap(response))
	})

	app.Post("/gameservers", func(c fiber.Ctx) error {
		var body models.Container

		if err := c.Bind().Body(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Bad Request", "details": err.Error()}) // TODO: do something else
		}
		container, err := wrapper.CreateContainer(body)
		if err != nil {
			// TODO: Do something else here
			return c.Status(500).JSON(fiber.Map{"error": "Error getting containers", "details": err.Error()})
		}

		if c.Get("HX-Request") == "true" {
			return c.Render("gameserver", utils.StructToFiberMap(container))
		} else {
			return c.Redirect().To("/gameservers")
		}
	})

	app.Patch("/gameservers/:id", func(c fiber.Ctx) error {
		containerID := c.Params("id")

		if containerID == "" { // TODO: do something else
			return c.Status(404).JSON(fiber.Map{"error": "Resource Not Found", "details": fmt.Sprintf("Container with ID=%s not found.", containerID)})
		}

		var body models.Container

		if err := c.Bind().Body(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Bad Request", "details": err.Error()}) // TODO: do something else
		}

		container, err := wrapper.UpdateContainer(containerID, body)
		if err != nil {
			// TODO: Do something else here
			return c.Status(500).JSON(fiber.Map{"error": "Error getting containers", "details": err.Error()})
		}

		if c.Get("HX-Request") == "true" {
			return c.Render("gameserver", utils.StructToFiberMap(container))
		} else {
			return c.Redirect().To("/gameservers")
		}
	})

	app.Delete("/gameservers/:id", func(c fiber.Ctx) error {
		containerID := c.Params("id")

		if containerID == "" { // TODO: do something else
			return c.Status(404).JSON(fiber.Map{"error": "Resource Not Found", "details": fmt.Sprintf("Container with ID=%s not found.", containerID)})
		}

		err := wrapper.DeleteContainer(containerID)
		if err != nil {
			// TODO: Do something else here
			return c.Status(500).JSON(fiber.Map{"error": "Error getting containers", "details": err.Error()})
		}

		if c.Get("HX-Request") == "true" {
			return c.Status(200).Send(nil)
		} else {
			return c.Redirect().To("/gameservers")
		}
	})

	app.Get("/nodes", func(c fiber.Ctx) error {
		nodes, err := wrapper.GetNodes()
		if err != nil {
			// TODO: Do something else here
			return c.Status(500).JSON(fiber.Map{"error": "Error getting containers", "details": err.Error()})
		}

		if c.Get("HX-Request") == "true" {
			return c.Render("nodes_page", fiber.Map{"Nodes": nodes})
		} else {
			return c.Render("nodes_page", fiber.Map{"Nodes": nodes}, "layout")
		}
	})

	log.Fatal(app.Listen(":3000")) // TODO: Get this from config
}
