package agent

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/0xKowalski1/container-orchestrator/runtime"
)

type DesiredState struct {
	Containers []struct {
		ID string `json:"id"`
	} `json:"containers"`
}

func StartAgent(_runtime runtime.Runtime) {
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
		actualMap := make(map[string]bool)
		for _, c := range actualContainers {
			actualMap[c.ID] = true
		}

		// Start missing containers
		for _, c := range desired.Containers {
			if !actualMap[c.ID] {
				// Create container
				_containerConfig := runtime.ContainerConfig{ID: c.ID, Image: "docker.io/itzg/minecraft-server:latest", Env: []string{"EULA=TRUE"}}

				_container, err := _runtime.CreateContainer("example", _containerConfig)
				if err != nil {
					log.Fatalf("Failed to create container: %v", err)
					continue
				}

				err = _runtime.StartContainer("example", _container.ID)
				if err != nil {
					log.Fatalf("Failed to start container: %v", err)
					continue
				}

			}
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
				// Do something
			}
		}
	}
}
