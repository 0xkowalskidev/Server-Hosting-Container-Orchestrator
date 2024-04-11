package storage

import (
	"0xKowalski1/container-orchestrator/config"
	"0xKowalski1/container-orchestrator/models"
	"fmt"
	"os"
	"path/filepath"
)

type StorageManager struct {
	cfg     *config.Config
	Volumes map[string]models.Volume
}

func NewStorageManager(cfg *config.Config) *StorageManager {
	return &StorageManager{
		cfg:     cfg,
		Volumes: make(map[string]models.Volume),
	}
}

func (sm *StorageManager) CreateVolume(volumeID string, sizeLimit int64) (models.Volume, error) {
	volumePath := filepath.Join(sm.cfg.StoragePath, volumeID)

	// Check if volume already exists
	if _, ok := sm.Volumes[volumeID]; ok {
		return models.Volume{}, fmt.Errorf("volume %s already exists", volumeID)
	}

	// Create the volume directory
	if err := os.MkdirAll(volumePath, 0755); err != nil {
		return models.Volume{}, fmt.Errorf("failed to create volume directory: %v", err)
	}

	volume := models.Volume{
		ID:         volumeID,
		MountPoint: volumePath,
		SizeLimit:  sizeLimit,
	}

	// Save the volume info
	sm.Volumes[volumeID] = volume

	return volume, nil
}

func (sm *StorageManager) RemoveVolume(volumeID string) error {
	volume, ok := sm.Volumes[volumeID]
	if !ok {
		return fmt.Errorf("volume %s does not exist", volumeID)
	}

	// Remove the volume directory
	if err := os.RemoveAll(volume.MountPoint); err != nil {
		return fmt.Errorf("failed to remove volume: %v", err)
	}

	// Remove the volume from the manager
	delete(sm.Volumes, volumeID)

	return nil
}
