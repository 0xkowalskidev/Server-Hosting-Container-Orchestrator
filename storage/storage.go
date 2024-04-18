package storage

import (
	"0xKowalski1/container-orchestrator/config"
	"0xKowalski1/container-orchestrator/utils"
	"log"

	"0xKowalski1/container-orchestrator/models"
	"fmt"
	"os"
	"path/filepath"
)

type StorageManager struct {
	cfg       *config.Config
	FileOps   utils.FileOpsInterface
	CmdRunner utils.CmdRunnerInterface
}

func NewStorageManager(cfg *config.Config, fileOps utils.FileOpsInterface, cmdRunner utils.CmdRunnerInterface) *StorageManager {
	return &StorageManager{
		cfg:       cfg,
		FileOps:   fileOps,
		CmdRunner: cmdRunner,
	}
}

func (sm *StorageManager) RemoveVolume(volumeID string) error {
	volumePath := filepath.Join(sm.cfg.StoragePath, volumeID)
	volumeFilePath := filepath.Join(sm.cfg.StoragePath, volumeID) + ".img"

	// Check if the volume directory exists
	if _, err := sm.FileOps.Stat(volumePath); os.IsNotExist(err) {
		return fmt.Errorf("volume %s does not exist", volumeID)
	}

	// No return on errors so that the cleanup continues
	// Unmount the volume
	if err := sm.unmountVolume(volumePath); err != nil {
		log.Printf("failed to unmount volume: %v", err)
	}

	// Remove the volume directory
	if err := sm.FileOps.RemoveAll(volumePath); err != nil {
		log.Printf("failed to remove volume: %v", err)
	}

	// Remove the loopback file, we assume it exists, but dont error if it does not
	if err := sm.FileOps.Remove(volumeFilePath); err != nil {
		log.Printf("failed to remove loopback file: %v", err)
	}

	return nil
}

func (sm *StorageManager) ListVolumes() ([]models.Volume, error) {
	var volumes []models.Volume
	entries, err := sm.FileOps.ReadDir(sm.cfg.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes directory: %v", err)
	}

	// We only want to return dirs
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
	if _, err := sm.FileOps.Stat(volumePath); !os.IsNotExist(err) {
		return nil, fmt.Errorf("volume %s already exists", volumeID)
	}

	// Create the volume directory
	if err := sm.FileOps.MkdirAll(volumePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create volume directory: %v", err)
	}

	// Create a fixed-size file
	if err := sm.createFixedSizeFile(volumeFilePath, sizeLimit); err != nil {
		sm.FileOps.RemoveAll(volumePath) //Rollback dir creation
		return nil, err
	}

	// Format the file with a filesystem, e.g., ext4
	if err := sm.formatAsExt4(volumeFilePath); err != nil {
		sm.FileOps.Remove(volumeFilePath) // Rollback file creation
		sm.FileOps.RemoveAll(volumePath)  //Rollback dir creation
		return nil, err
	}

	// Mount the file
	if err := sm.mountVolume(volumeFilePath, volumePath); err != nil {
		sm.FileOps.Remove(volumeFilePath) // Rollback file creation
		sm.FileOps.RemoveAll(volumePath)  //Rollback dir creation
		return nil, err
	}

	// Remove lost and found as intialization expectes volume to be empty.
	lfp := filepath.Join(volumePath, "lost+found")
	err := sm.FileOps.RemoveAll(lfp)
	if err != nil {
		sm.unmountVolume(volumePath)      // Rollback the mount
		sm.FileOps.Remove(volumeFilePath) // Rollback file creation
		sm.FileOps.RemoveAll(volumePath)  //Rollback dir creation
		return nil, fmt.Errorf("failed to remove lost and found dir: %v", err)
	}

	return &models.Volume{
		ID:         volumeID,
		MountPoint: volumePath,
		SizeLimit:  sizeLimit,
	}, nil
}

func (sm *StorageManager) createFixedSizeFile(filePath string, sizeLimit int64) error {
	sizeBytes := sizeLimit * 1024 * 1024 * 1024 // Convert GB to bytes
	err := sm.CmdRunner.RunCommand("fallocate", "-l", fmt.Sprintf("%d", sizeBytes), filePath)
	if err != nil {
		return fmt.Errorf("failed to create fixed-size file: %v", err)
	}
	return nil
}

func (sm *StorageManager) formatAsExt4(filePath string) error {
	err := sm.CmdRunner.RunCommand("mkfs.ext4", filePath)
	if err != nil {
		return fmt.Errorf("failed to format as ext4: %v", err)
	}
	return nil
}

func (sm *StorageManager) mountVolume(filePath, mountPath string) error {
	err := sm.CmdRunner.RunCommand("mount", "-o", "loop", filePath, mountPath)
	if err != nil {
		return fmt.Errorf("failed to mount volume: %v", err)
	}
	return nil
}

func (sm *StorageManager) unmountVolume(mountPath string) error {
	err := sm.CmdRunner.RunCommand("umount", mountPath)
	if err != nil {
		return fmt.Errorf("failed to unmount %s: %v", mountPath, err)
	}
	return nil
}
