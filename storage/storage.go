package storage

import (
	"0xKowalski1/container-orchestrator/config"

	"0xKowalski1/container-orchestrator/models"
	"fmt"
	"os"
	"path/filepath"
)

type StorageManager struct {
	cfg *config.Config
}

func NewStorageManager(cfg *config.Config) *StorageManager {
	return &StorageManager{
		cfg: cfg,
	}
}

func (sm *StorageManager) CreateVolume(volumeID string, sizeLimit int64) (models.Volume, error) {
	volumePath := filepath.Join(sm.cfg.StoragePath, volumeID)

	// Check if volume directory already exists
	if _, err := os.Stat(volumePath); !os.IsNotExist(err) {
		return models.Volume{}, fmt.Errorf("volume %s already exists", volumeID)
	}

	// Create the volume directory
	if err := os.MkdirAll(volumePath, 0755); err != nil {
		return models.Volume{}, fmt.Errorf("failed to create volume directory: %v", err)
	}

	// Create and return the volume object
	volume := models.Volume{
		ID:         volumeID,
		MountPoint: volumePath,
		SizeLimit:  sizeLimit,
	}
	return volume, nil
}

func (sm *StorageManager) RemoveVolume(volumeID string) error {
	volumePath := filepath.Join(sm.cfg.StoragePath, volumeID)

	// Check if the volume directory exists
	if _, err := os.Stat(volumePath); os.IsNotExist(err) {
		return fmt.Errorf("volume %s does not exist", volumeID)
	}

	// Remove the volume directory
	if err := os.RemoveAll(volumePath); err != nil {
		return fmt.Errorf("failed to remove volume: %v", err)
	}

	return nil
}

func (sm *StorageManager) ListVolumes() ([]models.Volume, error) {
	var volumes []models.Volume
	entries, err := os.ReadDir(sm.cfg.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			volumeID := entry.Name()
			volumePath := filepath.Join(sm.cfg.StoragePath, volumeID)

			volume := models.Volume{
				ID:         volumeID,
				MountPoint: volumePath,
			}
			volumes = append(volumes, volume)
		}
	}
	return volumes, nil
}
