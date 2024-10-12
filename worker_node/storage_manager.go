package workernode

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/models"
	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/utils"
)

type StorageManager struct {
	config  Config
	fileOps utils.FileOpsInterface
}

func NewStorageManager(config Config, fileOps utils.FileOpsInterface) *StorageManager {
	return &StorageManager{
		config:  config,
		fileOps: fileOps,
	}
}

func (sm *StorageManager) SyncStorage(desiredContainers []models.Container) error {
	actualVolumes, err := sm.ListVolumes()
	if err != nil {
		return err
	}

	actualMap := make(map[string]models.Volume)
	for _, volume := range actualVolumes {
		actualMap[volume.ID] = volume
	}

	desiredMap := make(map[string]models.Volume)
	for _, desiredContainer := range desiredContainers {
		volume := models.Volume{ID: desiredContainer.ID, SizeLimit: int64(desiredContainer.StorageLimit)}
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

func (sm *StorageManager) RemoveVolume(volumeID string) error {
	volumePath := filepath.Join(sm.config.MountsPath, volumeID)
	volumeFilePath := filepath.Join(sm.config.MountsPath, volumeID) + ".img"

	// Check if the volume directory exists
	if _, err := sm.fileOps.Stat(volumePath); os.IsNotExist(err) {
		return fmt.Errorf("volume %s does not exist", volumeID)
	}

	// No return on errors so that the cleanup continues
	// Unmount the volume
	if err := sm.unmountVolume(volumePath); err != nil {
		log.Printf("failed to unmount volume: %v", err)
	}

	// Remove the volume directory
	if err := sm.fileOps.RemoveAll(volumePath); err != nil {
		log.Printf("failed to remove volume: %v", err)
	}

	// Remove the loopback file, we assume it exists, but dont error if it does not
	if err := sm.fileOps.Remove(volumeFilePath); err != nil {
		log.Printf("failed to remove loopback file: %v", err)
	}

	if err := sm.deleteUser(volumeID); err != nil {
		log.Printf("Failed to delete user: %v", err)
	}

	return nil
}

func (sm *StorageManager) ListVolumes() ([]models.Volume, error) {
	var volumes []models.Volume
	entries, err := sm.fileOps.ReadDir(sm.config.MountsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes directory: %v", err)
	}

	// We only want to return dirs
	for _, entry := range entries {
		if entry.IsDir() {
			volumeID := entry.Name()
			volumePath := filepath.Join(sm.config.MountsPath, volumeID)

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
	volumePath := filepath.Join(sm.config.MountsPath, volumeID)
	volumeFilePath := filepath.Join(sm.config.MountsPath, volumeID) + ".img"

	// Check if volume directory already exists
	if _, err := sm.fileOps.Stat(volumePath); !os.IsNotExist(err) {
		return nil, fmt.Errorf("volume %s already exists", volumeID)
	}

	// Create the volume directory
	if err := sm.fileOps.MkdirAll(volumePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create volume directory: %v", err)
	}

	// Create a fixed-size file
	if err := sm.createFixedSizeFile(volumeFilePath, sizeLimit); err != nil {
		sm.fileOps.RemoveAll(volumePath) //Rollback dir creation
		return nil, err
	}

	// Format the file with a filesystem, e.g., ext4
	if err := sm.formatAsExt4(volumeFilePath); err != nil {
		sm.fileOps.Remove(volumeFilePath) // Rollback file creation
		sm.fileOps.RemoveAll(volumePath)  //Rollback dir creation
		return nil, err
	}

	// Mount the file
	if err := sm.mountVolume(volumeFilePath, volumePath); err != nil {
		sm.fileOps.Remove(volumeFilePath) // Rollback file creation
		sm.fileOps.RemoveAll(volumePath)  //Rollback dir creation
		return nil, err
	}

	// Remove lost and found as intialization expectes volume to be empty.
	lfp := filepath.Join(volumePath, "lost+found") // TODO: This may cause issues with crashes/restart/recovery, explore further
	err := sm.fileOps.RemoveAll(lfp)
	if err != nil {
		sm.unmountVolume(volumePath)      // Rollback the mount
		sm.fileOps.Remove(volumeFilePath) // Rollback file creation
		sm.fileOps.RemoveAll(volumePath)  //Rollback dir creation
		return nil, fmt.Errorf("failed to remove lost and found dir: %v", err)
	}

	err = sm.createUser(volumeID, "password", volumePath)
	if err != nil {
		sm.unmountVolume(volumePath)      // Rollback the mount
		sm.fileOps.Remove(volumeFilePath) // Rollback file creation
		sm.fileOps.RemoveAll(volumePath)  //Rollback dir creation
		return nil, fmt.Errorf("failed to remove create user: %v", err)
	}

	return &models.Volume{
		ID:         volumeID,
		MountPoint: volumePath,
		SizeLimit:  sizeLimit,
	}, nil
}

// TODO: Make utils fpr these so this is easier to test
func (sm *StorageManager) createFixedSizeFile(filePath string, sizeLimit int64) error {
	// Open file with read-write permissions and create it if it doesn't exist
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Use the syscall package to allocate space for the file
	sizeBytes := sizeLimit * 1024 * 1024 * 1024 // Convert GB to bytes
	err = syscall.Fallocate(int(file.Fd()), 0, 0, sizeBytes)
	if err != nil {
		return fmt.Errorf("failed to allocate space: %v", err)
	}

	return nil
}

func (sm *StorageManager) formatAsExt4(filePath string) error {
	cmd := exec.Command("mkfs.ext4", filePath) // TODO: Feels like there should be a better way to do this!
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to format as ext4: %v", err)
	}
	return nil
}

func (sm *StorageManager) mountVolume(filePath, mountPath string) error {
	cmd := exec.Command("mount", "-o", "loop", filePath, mountPath)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to mount volume: %v", err)
	}
	return nil
}
func (sm *StorageManager) unmountVolume(mountPath string) error {
	err := syscall.Unmount(mountPath, 0)
	if err != nil {
		return fmt.Errorf("failed to unmount %s: %v", mountPath, err)
	}
	return nil
}

func (sm *StorageManager) createUser(username, password, dir string) error {
	group := "sftpusers" // TODO get this from config

	userAddCmd := exec.Command("sudo", "useradd", "-M", "-s", "/sbin/nologin", "-G", group, username)
	if output, err := userAddCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create user: %v, output: %s", err, output)
	}

	passwdCmd := exec.Command("sudo", "chpasswd")
	passwdCmd.Stdin = strings.NewReader(fmt.Sprintf("%s:%s", username, password))
	if output, err := passwdCmd.CombinedOutput(); err != nil {
		_ = sm.deleteUser(username)
		return fmt.Errorf("failed to set user password: %v, output: %s", err, output)
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		_ = sm.deleteUser(username)
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Set ownership of the directory to the user
	chownCmd := exec.Command("sudo", "chown", "-R", fmt.Sprintf("%s:%s", username, group), dir)
	if output, err := chownCmd.CombinedOutput(); err != nil {
		_ = sm.deleteUser(username)
		return fmt.Errorf("failed to set directory ownership: %v, output: %s", err, output)
	}

	return nil
}

func (sm *StorageManager) deleteUser(username string) error {
	userDelCmd := exec.Command("sudo", "userdel", username)
	if err := userDelCmd.Run(); err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}
	return nil
}
