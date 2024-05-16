package storage

import (
	"0xKowalski1/container-orchestrator/models"
	"log"
)

func (sm *StorageManager) SyncStorage(desiredVolumes []models.Volume) error {
	// Get list of storage paths
	actualVolumes, err := sm.ListVolumes()
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
			if _, err := sm.CreateVolume(volume.ID, volume.SizeLimit); err != nil {
				log.Printf("failed to create volume %s: %v", volumeID, err)
			}
		}
	}

	// Delete volumes that are in the actual state but not in the desired state (May want to handle this differently?)
	for volumeID := range actualMap {
		if _, desired := desiredMap[volumeID]; !desired {
			if err := sm.RemoveVolume(volumeID); err != nil {
				log.Printf("failed to remove volume %s: %v", volumeID, err)
			}
		}
	}

	// probably want to return a list of errors and unschedule/retry those containers
	return nil
}
