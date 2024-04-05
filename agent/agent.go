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

type ApiResponse struct {
	Node struct {
		ID         string                   `json:"ID"`
		Containers []statemanager.Container `json:"Containers"`
	} `json:"node"`
}

func Start(_runtime runtime.Runtime) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Fetch desired statr
		resp, err := http.Get("http://localhost:8080/nodes/node-1") // get node-id from config
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

		var apiResponse ApiResponse
		if err := json.Unmarshal(body, &apiResponse); err != nil {
			log.Printf("Error unmarshaling desired state: %v", err)
			continue
		}

		desiredContainers := apiResponse.Node.Containers

		// List actual containers
		actualContainers, err := _runtime.ListContainers("example") // need to list accross all namespaces
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

		for _, desiredContainer := range desiredContainers {
			// Create missing containers
			if _, exists := actualMap[desiredContainer.ID]; !exists {
				// Create container if it does not exist in actual state
				_, err := _runtime.CreateContainer(desiredContainer.NamespaceID, desiredContainer)
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
			for _, d := range desiredContainers {
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

func reconcileContainerState(_runtime runtime.Runtime, desiredContainer statemanager.Container, actualContainer statemanager.Container) {
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
			err := _runtime.StopContainer(desiredContainer.NamespaceID, desiredContainer.ID, desiredContainer.StopTimeout)
			if err != nil {
				log.Fatalf("Failed to stop container: %v", err)
			}
		}
	}
}
