package agent

import (
	"0xKowalski1/container-orchestrator/models"
	"fmt"
)

func (a *Agent) syncStorage(desiredVolumes []models.Volume) error {
	// Get list of storage paths
	actualVolumes, err := a.storage.ListVolumes()
	if err != nil {
		return err
	}

	actualMap := make(map[string]models.Volume)
	for _, volume := range actualVolumes {
		actualMap[volume.ID] = volume
	}

	desiredMap := make(map[string]models.Volume)
	for _, volume := range desiredVolumes {
		desiredMap[volume.ID] = volume
	}

	// Create volumes that are in the desired state but not in the actual state
	for volumeID, volume := range desiredMap {
		if _, exists := actualMap[volumeID]; !exists {
			if _, err := a.storage.CreateVolume(volume.ID, volume.SizeLimit); err != nil {
				return fmt.Errorf("failed to create volume %s: %v", volumeID, err)
			}
		}
	}

	// Delete volumes that are in the actual state but not in the desired state
	for volumeID := range actualMap {
		if _, desired := desiredMap[volumeID]; !desired {
			if err := a.storage.RemoveVolume(volumeID); err != nil {
				return fmt.Errorf("failed to remove volume %s: %v", volumeID, err)
			}
		}
	}

	return nil
}
