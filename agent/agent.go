package agent

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/0xKowalski1/container-orchestrator/runtime"
	statemanager "github.com/0xKowalski1/container-orchestrator/state-manager"
)

type DesiredContainer struct {
	ID            string `json:"id"`
	DesiredStatus string `json:"desiredStatus"`
}

type DesiredState struct {
	Containers []DesiredContainer `json:"containers"`
}

func Start(_runtime runtime.Runtime) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Fetch desired state
		resp, err := http.Get("http://localhost:8080/nodes/node-1/desired")
		if err != nil {
			log.Printf("Error checking for nodes desired state: %v", err)
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			log.Printf("Error reading response body: %v", err)
			continue
		}

		var desired DesiredState
		if err := json.Unmarshal(body, &desired); err != nil {
			log.Printf("Error unmarshaling desired state: %v", err)
			continue
		}

		// List actual containers
		actualContainers, err := _runtime.ListContainers("example")
		if err != nil {
			log.Printf("Error listing containers: %v", err)
			continue
		}

		// Map actual container IDs for easier lookup
		actualMap := make(map[string]statemanager.Container)
		for _, c := range actualContainers {
			ic, err := _runtime.InspectContainer("example", c.ID)
			if err != nil {
				log.Printf("Error inspecting container: %v", err)
			}
			actualMap[c.ID] = ic
		}

		for _, desiredContainer := range desired.Containers {
			// Create missing containers
			if _, exists := actualMap[desiredContainer.ID]; !exists {
				// Create container if it does not exist in actual state
				_containerConfig := runtime.ContainerConfig{ID: desiredContainer.ID, Image: "docker.io/itzg/minecraft-server:latest", Env: []string{"EULA=TRUE"}}

				_, err := _runtime.CreateContainer("example", _containerConfig)
				if err != nil {
					log.Printf("Failed to create container: %v", err)
					continue
				}
			}

			reconcileContainerState(_runtime, desiredContainer, actualMap[desiredContainer.ID])
		}

		// Stop extra containers
		for _, c := range actualContainers {
			found := false
			for _, d := range desired.Containers {
				if d.ID == c.ID {
					found = true
					break
				}
			}
			if !found {
				_runtime.StopContainer("example", c.ID, 5) // timeout should come from container config
				_runtime.RemoveContainer("example", c.ID)
			}
		}
	}
}

func reconcileContainerState(_runtime runtime.Runtime, desiredContainer DesiredContainer, actualContainer statemanager.Container) {
	switch desiredContainer.DesiredStatus {
	case "running":
		if actualContainer.Status != "running" {
			err := _runtime.StartContainer("example", desiredContainer.ID) // Probably a bug here, if we use actualContainer this fails as ID is missing
			if err != nil {
				log.Fatalf("Failed to start container: %v", err)
			}

		}
	case "stopped":
		if actualContainer.Status != "stopped" {
			err := _runtime.StopContainer("example", desiredContainer.ID, 5)
			if err != nil {
				log.Fatalf("Failed to stop container: %v", err)
			}
		}
	}
}
