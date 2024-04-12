package agent

import (
	"0xKowalski1/container-orchestrator/models"
)

func (a *Agent) syncNetworking(desiredPortmaps []models.Portmap) error {
	// Get acutal network network namespaces

	actualNamespaces, err := a.networking.ListNetworkNamespaces()
	if err != nil {
		return err
	}

	// Iterate over desiredPortmaps
	// If no network namespace, make one
	// If no network rule, make one

	// Iterate over network namespaces
	// If namespace exists, but shouldent, delete

	return nil
}
