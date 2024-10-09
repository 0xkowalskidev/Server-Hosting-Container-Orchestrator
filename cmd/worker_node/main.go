package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/utils"
	workernode "github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/worker_node"
	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/hpcloud/tail"
)

func main() {
	var config workernode.Config
	utils.ParseConfigFromEnv(&config)

	runtime, err := workernode.NewContainerdRuntime(config)
	if err != nil {
		log.Fatalf("Failed to initialize runtime: %v", err)
	}

	// Metrics & Logs Api
	app := fiber.New()

	app.Get("/logs/:containerID", func(c fiber.Ctx) error {
		logPath := fmt.Sprintf("%s/%s.log", config.LogsPath, c.Params("containerID"))

		t, err := tail.TailFile(logPath, tail.Config{
			Follow:    true,  // Follow the file for new data
			ReOpen:    true,  // Reopen the file if it gets rotated
			MustExist: false, // Don't fail if file doesn't exist yet, might not want this
			Poll:      true,  // Use polling instead of inotify (better for containers apparently)
		})
		if err != nil {
			return c.Status(500).SendString("Failed to open log file")
		}

		reader, writer := io.Pipe()

		go func() {
			for line := range t.Lines {
				_, err := writer.Write([]byte("data: " + line.Text + "<br> \n\n"))
				if err != nil {
					log.Printf("Error writing to pipe: %v", err)
					writer.Close()
					return
				}
			}
			writer.Close()
		}()

		return c.SendStream(reader)
	})

	app.Get("/metrics/:containerID", func(c fiber.Ctx) error {
		metrics, err := runtime.GetContainerMetrics(context.Background(), c.Params("containerID"), config.ContainerdNamespace)
		if err != nil {
			log.Println(err)
		}
		return c.JSON(metrics)
	})

	app.Get("/status/:containerID", func(c fiber.Ctx) error {
		status, err := runtime.GetContainerStatus(context.Background(), c.Params("containerID"), config.ContainerdNamespace)
		if err != nil {
			log.Println(err)
		}
		return c.JSON(map[string]string{"status": string(status)})
	})

	go app.Listen(":3002")

	storageManager := workernode.NewStorageManager(config, &utils.FileOps{})

	networkManager, err := workernode.NewNetworkManager(config, &utils.FileOps{}, &utils.CmdRunner{})
	if err != nil {
		log.Fatalf("Failed to initialize network manager: %v", err)
	}

	client := resty.New()
	agent := workernode.NewAgent(config, client, runtime, storageManager, networkManager)

	agent.StartAgent()
}
