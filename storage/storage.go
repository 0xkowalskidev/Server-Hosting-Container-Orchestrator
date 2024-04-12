package storage

import (
	"0xKowalski1/container-orchestrator/config"
	"log"
	"os/exec"

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

func (sm *StorageManager) RemoveVolume(volumeID string) error {
	volumePath := filepath.Join(sm.cfg.StoragePath, volumeID)
	volumeFilePath := filepath.Join(sm.cfg.StoragePath, volumeID) + ".img"

	// Check if the volume directory exists
	if _, err := os.Stat(volumePath); os.IsNotExist(err) {
		return fmt.Errorf("volume %s does not exist", volumeID)
	}

	// Unmount the volume
	if err := unmountVolume(volumePath); err != nil {
		log.Printf("failed to unmount volume: %v", err)
	}

	// Remove the volume directory
	if err := os.RemoveAll(volumePath); err != nil {
		log.Printf("failed to remove volume: %v", err)
	}

	// Remove the loopback file
	if err := os.Remove(volumeFilePath); err != nil {
		log.Printf("failed to remove loopback file: %v", err)
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

func (sm *StorageManager) CreateVolume(volumeID string, sizeLimit int64) (*models.Volume, error) {
	volumePath := filepath.Join(sm.cfg.StoragePath, volumeID)
	volumeFilePath := filepath.Join(sm.cfg.StoragePath, volumeID) + ".img"

	// Check if volume directory already exists
	if _, err := os.Stat(volumePath); !os.IsNotExist(err) {
		return nil, fmt.Errorf("volume %s already exists", volumeID)
	}

	// Create the volume directory
	if err := os.MkdirAll(volumePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create volume directory: %v", err)
	}

	// Create a fixed-size file
	if err := createFixedSizeFile(volumeFilePath, sizeLimit); err != nil {
		return nil, err
	}

	// Format the file with a filesystem, e.g., ext4
	if err := formatAsExt4(volumeFilePath); err != nil {
		return nil, err
	}

	// Mount the file
	if err := mountVolume(volumeFilePath, volumePath); err != nil {
		return nil, err
	}

	// Remove lost and found as we wont be using it (and dont want users to see it)
	lfp := filepath.Join(volumePath, "lost+found")
	err := os.RemoveAll(lfp)
	if err != nil {
		return nil, fmt.Errorf("failed to remove lost and found dir: %v", err)
	}

	// Create and return the volume object
	volume := models.Volume{
		ID:         volumeID,
		MountPoint: volumePath,
		SizeLimit:  sizeLimit,
	}
	return &volume, nil
}

func createFixedSizeFile(filePath string, sizeLimit int64) error {
	log.Printf("%d", sizeLimit)
	sizeBytes := sizeLimit * 1024 * 1024 * 1024 // Convert GB to bytes
	cmd := exec.Command("fallocate", "-l", fmt.Sprintf("%d", sizeBytes), filePath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create fixed-size file: %v", err)
	}
	return nil
}

func formatAsExt4(filePath string) error {
	cmd := exec.Command("mkfs.ext4", filePath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to format as ext4: %v", err)
	}
	return nil
}

func mountVolume(filePath, mountPath string) error {
	cmd := exec.Command("mount", "-o", "loop", filePath, mountPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to mount volume: %v", err)
	}
	return nil
}

func unmountVolume(mountPath string) error {
	cmd := exec.Command("umount", mountPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to unmount %s: %v", mountPath, err)
	}
	return nil
}
