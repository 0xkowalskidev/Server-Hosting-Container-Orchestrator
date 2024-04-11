package agent

import (
	"0xKowalski1/container-orchestrator/models"
	"log"
)

func (a *Agent) syncContainers(node *models.Node) error {
	desiredContainers := node.Containers

	// List actual containers
	actualContainers, err := a.runtime.ListContainers()
	if err != nil {
		log.Printf("Error listing containers: %v", err)
		return err
	}

	// Map actual container IDs for easier lookup
	actualMap := make(map[string]models.Container)
	for _, c := range actualContainers {
		ic, err := a.runtime.InspectContainer(c.ID)
		if err != nil {
			log.Printf("Error inspecting container: %v", err)
		}
		actualMap[c.ID] = ic
	}

	for _, desiredContainer := range desiredContainers {
		// Create missing containers
		if _, exists := actualMap[desiredContainer.ID]; !exists {
			// Create container if it does not exist in actual state

			a.storage.CreateVolume(desiredContainer.ID, 1000) // Check errors here
			a.networking.SetupContainerNetwork(desiredContainer.ID, desiredContainer.Ports)
			_, err := a.runtime.CreateContainer(desiredContainer)
			if err != nil {
				log.Printf("Failed to create container: %v", err)
				continue
			}
		}

		a.reconcileContainerState(desiredContainer, actualMap[desiredContainer.ID])
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
			// Check errors
			a.runtime.StopContainer(c.ID, c.StopTimeout)
			a.runtime.RemoveContainer(c.ID)
			a.storage.RemoveVolume(c.ID)
			a.networking.CleanupContainerNetwork(c.ID)
		}
	}

	return nil
}

func (a *Agent) reconcileContainerState(desiredContainer models.Container, actualContainer models.Container) {
	switch desiredContainer.DesiredStatus {
	case "running":
		if actualContainer.Status != "running" {

			err := a.runtime.StartContainer(desiredContainer.ID)
			if err != nil {
				log.Fatalf("Failed to start container: %v", err)
			}

		}
	case "stopped":
		if actualContainer.Status != "stopped" {
			err := a.runtime.StopContainer(desiredContainer.ID, desiredContainer.StopTimeout)
			if err != nil {
				log.Fatalf("Failed to stop container: %v", err)
			}
		}
	}
}
