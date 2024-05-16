package agent

import (
	"0xKowalski1/container-orchestrator/models"
	"fmt"
)

func (a *Agent) syncNetworking(desiredContainers []models.Container) error {
	actualNamespaces, err := a.networking.ListNetworkNamespaces()
	if err != nil {
		return err
	}

	desiredMap := make(map[string]models.Container)
	for _, container := range desiredContainers {
		desiredMap[container.ID] = container
	}

	actualMap := make(map[string]bool)
	for _, namespace := range actualNamespaces { // Namespace is a containerID
		actualMap[namespace] = true
	}

	for containerID, container := range desiredMap {
		if !actualMap[containerID] {
			err := a.networking.SetupContainerNetwork(containerID, container.Ports)
			if err != nil {
				return fmt.Errorf("Failed to setup container network: %v", err)
			}

		}
	}

	for namespace := range actualMap {
		if _, desired := desiredMap[namespace]; !desired {
			err := a.networking.CleanupContainerNetwork(namespace)
			if err != nil {
				return fmt.Errorf("Failed to cleanup container network: %v", err)
			}
		}
	}

	return nil
}
