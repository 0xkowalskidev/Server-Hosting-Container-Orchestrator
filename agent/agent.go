package agent

import (
	"log"
	"time"

	"0xKowalski1/container-orchestrator/api"
	"0xKowalski1/container-orchestrator/config"
	"0xKowalski1/container-orchestrator/models"
	"0xKowalski1/container-orchestrator/runtime"
)

type ApiResponse struct {
	Node struct {
		ID         string             `json:"ID"`
		Containers []models.Container `json:"Containers"`
	} `json:"node"`
}

func Start(cfg *config.Config) {
	// start runtime
	_runtime, err := runtime.NewRuntime("containerd", cfg)

	if err != nil {
		log.Fatalf("Failed to initialize runtime: %v", err)
	}

	apiClient := api.NewApiWrapper(cfg.Namespace)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		node, err := apiClient.GetNode("node-1") //temp
		if err != nil {
			log.Printf("Error checking for nodes desired state: %v", err)
			continue
		}

		desiredContainers := node.Containers

		// List actual containers
		actualContainers, err := _runtime.ListContainers(cfg.Namespace)
		if err != nil {
			log.Printf("Error listing containers: %v", err)
			continue
		}

		// Map actual container IDs for easier lookup
		actualMap := make(map[string]models.Container)
		for _, c := range actualContainers {
			ic, err := _runtime.InspectContainer(cfg.Namespace, c.ID)
			if err != nil {
				log.Printf("Error inspecting container: %v", err)
			}
			actualMap[c.ID] = ic
		}

		for _, desiredContainer := range desiredContainers {
			// Create missing containers
			if _, exists := actualMap[desiredContainer.ID]; !exists {
				// Create container if it does not exist in actual state
				_, err := _runtime.CreateContainer(cfg.Namespace, desiredContainer)
				if err != nil {
					log.Printf("Failed to create container: %v", err)
					continue
				}
			}

			reconcileContainerState(cfg, _runtime, desiredContainer, actualMap[desiredContainer.ID])
		}

		// Stop extra containers
		for _, c := range actualContainers {
			found := false
			for _, d := range desiredContainers {
				if d.ID == c.ID {
					found = true
					break
				}
			}
			if !found {
				_runtime.StopContainer(cfg.Namespace, c.ID, 5) // timeout should come from container config
				_runtime.RemoveContainer(cfg.Namespace, c.ID)
			}
		}
	}
}

func reconcileContainerState(cfg *config.Config, _runtime runtime.Runtime, desiredContainer models.Container, actualContainer models.Container) {
	switch desiredContainer.DesiredStatus {
	case "running":
		if actualContainer.Status != "running" {
			err := _runtime.StartContainer(cfg.Namespace, desiredContainer.ID) // Probably a bug here, if we use actualContainer this fails as ID is missing
			if err != nil {
				log.Fatalf("Failed to start container: %v", err)
			}

		}
	case "stopped":
		if actualContainer.Status != "stopped" {
			err := _runtime.StopContainer(cfg.Namespace, desiredContainer.ID, desiredContainer.StopTimeout)
			if err != nil {
				log.Fatalf("Failed to stop container: %v", err)
			}
		}
	}
}
