package storage_test

import (
	"0xKowalski1/container-orchestrator/config"
	"0xKowalski1/container-orchestrator/models"
	"0xKowalski1/container-orchestrator/storage"
	utils_test "0xKowalski1/container-orchestrator/tests/utils"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var fakeStoragePath = "/fakepath"

func setup() (*storage.StorageManager, *utils_test.MockFileOps, *utils_test.MockCmdRunner) {
	cfg := &config.Config{StoragePath: fakeStoragePath}
	mockFileOps := new(utils_test.MockFileOps)
	mockCmdRunner := new(utils_test.MockCmdRunner)

	return storage.NewStorageManager(cfg, mockFileOps, mockCmdRunner), mockFileOps, mockCmdRunner
}

// RemoveVolume
func TestStorageManager_RemoveVolume(t *testing.T) {
	sm, mockFileOps, mockCmdRunner := setup()

	volume := models.Volume{ID: "volume1", MountPoint: fmt.Sprintf("%s/volume1", fakeStoragePath), SizeLimit: 1}
	imageMountPoint := fmt.Sprintf("%s.img", volume.MountPoint)

	// Mock the scenario where the volume exists
	fakeFileInfo := utils_test.NewFakeFileInfo(volume.ID, volume.SizeLimit*1024*1024*1024, true)
	mockFileOps.On("Stat", volume.MountPoint).Return(fakeFileInfo, nil)
	mockCmdRunner.On("RunCommand", "umount", volume.MountPoint).Return(nil)
	mockFileOps.On("RemoveAll", volume.MountPoint).Return(nil)
	mockFileOps.On("Remove", imageMountPoint).Return(nil)

	// Test removing an existing volume
	err := sm.RemoveVolume(volume.ID)
	assert.NoError(t, err)

	mockFileOps.AssertExpectations(t)
	mockCmdRunner.AssertExpectations(t)
}

func TestStorageManager_RemoveVolume_NonExistantVolume(t *testing.T) {
	sm, mockFileOps, _ := setup()

	volumeID := "noVolume"
	volumePath := fmt.Sprintf("%s/%s", fakeStoragePath, volumeID)

	mockFileOps.On("Stat", volumePath).Return((*utils_test.FakeFileInfo)(nil), os.ErrNotExist)

	err := sm.RemoveVolume(volumeID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")

	mockFileOps.AssertExpectations(t)
}

// ListVolumes
func TestStorageManager_ListVolumes(t *testing.T) {
	sm, mockFileOps, _ := setup()

	fakeDirOne := utils_test.NewFakeDirEntry("volume1", true)
	fakeDirTwo := utils_test.NewFakeDirEntry("volume2", true)

	mockFileOps.On("ReadDir", fakeStoragePath).Return([]os.DirEntry{
		fakeDirOne, fakeDirTwo,
	}, nil)

	volumes, err := sm.ListVolumes()

	assert.NoError(t, err)
	assert.Len(t, volumes, 2)
	assert.Equal(t, fakeDirOne.Name(), volumes[0].ID)
	assert.Equal(t, fmt.Sprintf("%s/%s", fakeStoragePath, fakeDirOne.Name()), volumes[0].MountPoint)
	assert.Equal(t, fakeDirTwo.Name(), volumes[1].ID)
	assert.Equal(t, fmt.Sprintf("%s/%s", fakeStoragePath, fakeDirTwo.Name()), volumes[1].MountPoint)

	mockFileOps.AssertExpectations(t)
}

// CreateVolume
func TestStorageManager_CreateVolume(t *testing.T) {
	sm, mockFileOps, mockCmdRunner := setup()

	expectedVolume := models.Volume{ID: "volume1", MountPoint: fmt.Sprintf("%s/volume1", fakeStoragePath), SizeLimit: 1}

	mountPoint := expectedVolume.MountPoint
	imageMountPoint := fmt.Sprintf("%s.img", expectedVolume.MountPoint)

	mockFileOps.On("Stat", mountPoint).Return(utils_test.FakeFileInfo{}, os.ErrNotExist)
	mockFileOps.On("MkdirAll", mountPoint, os.FileMode(0755)).Return(nil)
	mockCmdRunner.On("RunCommand", "fallocate", "-l", "1073741824", imageMountPoint).Return(nil)
	mockCmdRunner.On("RunCommand", "mkfs.ext4", imageMountPoint).Return(nil)
	mockCmdRunner.On("RunCommand", "mount", "-o", "loop", imageMountPoint, mountPoint).Return(nil)
	mockFileOps.On("RemoveAll", fmt.Sprintf("%s/lost+found", mountPoint)).Return(nil)

	volume, err := sm.CreateVolume(expectedVolume.ID, expectedVolume.SizeLimit)
	assert.NoError(t, err)
	assert.NotNil(t, volume)
	assert.Equal(t, expectedVolume.ID, volume.ID)
	assert.Equal(t, mountPoint, volume.MountPoint)

	mockFileOps.AssertExpectations(t)
	mockCmdRunner.AssertExpectations(t)
}

// CreateVolume rollbacks
func TestStorageManager_CreateVolume_FailureAtCreateFixedSizeFile(t *testing.T) {
	sm, mockFileOps, mockCmdRunner := setup()

	volumeID := "volume1"
	mountPoint := fmt.Sprintf("%s/%s", fakeStoragePath, volumeID)
	imageMountPoint := fmt.Sprintf("%s.img", mountPoint)

	mockFileOps.On("Stat", mountPoint).Return(utils_test.FakeFileInfo{}, os.ErrNotExist)
	mockFileOps.On("MkdirAll", mountPoint, os.FileMode(0755)).Return(nil)
	mockCmdRunner.On("RunCommand", "fallocate", "-l", "1073741824", imageMountPoint).Return(fmt.Errorf("fallocate failed"))

	// Expect RemoveAll to be called to clean up the created directory
	mockFileOps.On("RemoveAll", mountPoint).Return(nil)

	_, err := sm.CreateVolume(volumeID, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fallocate failed")

	mockFileOps.AssertExpectations(t)
	mockCmdRunner.AssertExpectations(t)
}

func TestStorageManager_CreateVolume_FailureAtFormatAsExt4(t *testing.T) {
	sm, mockFileOps, mockCmdRunner := setup()

	volumeID := "volume1"
	mountPoint := fmt.Sprintf("%s/%s", fakeStoragePath, volumeID)
	imageMountPoint := fmt.Sprintf("%s.img", mountPoint)

	mockFileOps.On("Stat", mountPoint).Return(utils_test.FakeFileInfo{}, os.ErrNotExist)
	mockFileOps.On("MkdirAll", mountPoint, os.FileMode(0755)).Return(nil)
	mockCmdRunner.On("RunCommand", "fallocate", "-l", "1073741824", imageMountPoint).Return(nil)
	mockCmdRunner.On("RunCommand", "mkfs.ext4", imageMountPoint).Return(fmt.Errorf("mkfs.ext4 failed"))

	// Expectations for rollback actions
	mockFileOps.On("Remove", imageMountPoint).Return(nil)
	mockFileOps.On("RemoveAll", mountPoint).Return(nil)

	_, err := sm.CreateVolume(volumeID, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mkfs.ext4 failed")

	mockFileOps.AssertExpectations(t)
	mockCmdRunner.AssertExpectations(t)
}

func TestStorageManager_CreateVolume_FailureAtMountVolume(t *testing.T) {
	sm, mockFileOps, mockCmdRunner := setup()

	volumeID := "volume1"
	mountPoint := fmt.Sprintf("%s/%s", fakeStoragePath, volumeID)
	imageMountPoint := fmt.Sprintf("%s.img", mountPoint)

	mockFileOps.On("Stat", mountPoint).Return(utils_test.FakeFileInfo{}, os.ErrNotExist)
	mockFileOps.On("MkdirAll", mountPoint, os.FileMode(0755)).Return(nil)
	mockCmdRunner.On("RunCommand", "fallocate", "-l", "1073741824", imageMountPoint).Return(nil)
	mockCmdRunner.On("RunCommand", "mkfs.ext4", imageMountPoint).Return(nil)
	mockCmdRunner.On("RunCommand", "mount", "-o", "loop", imageMountPoint, mountPoint).Return(fmt.Errorf("mount failed"))

	// Expectations for rollback actions
	mockFileOps.On("Remove", imageMountPoint).Return(nil)
	mockFileOps.On("RemoveAll", mountPoint).Return(nil)

	_, err := sm.CreateVolume(volumeID, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mount failed")

	mockFileOps.AssertExpectations(t)
	mockCmdRunner.AssertExpectations(t)
}

func TestStorageManager_CreateVolume_FailureAtRemovingLostAndFound(t *testing.T) {
	sm, mockFileOps, mockCmdRunner := setup()

	volumeID := "volume1"
	mountPoint := fmt.Sprintf("%s/%s", fakeStoragePath, volumeID)
	imageMountPoint := fmt.Sprintf("%s.img", mountPoint)
	lostFoundPath := fmt.Sprintf("%s/lost+found", mountPoint)

	// Setup to pass all previous steps
	mockFileOps.On("Stat", mountPoint).Return(utils_test.FakeFileInfo{}, os.ErrNotExist)
	mockFileOps.On("MkdirAll", mountPoint, os.FileMode(0755)).Return(nil)
	mockCmdRunner.On("RunCommand", "fallocate", "-l", "1073741824", imageMountPoint).Return(nil)
	mockCmdRunner.On("RunCommand", "mkfs.ext4", imageMountPoint).Return(nil)
	mockCmdRunner.On("RunCommand", "mount", "-o", "loop", imageMountPoint, mountPoint).Return(nil)

	// Simulate failure in removing the lost+found directory
	mockFileOps.On("RemoveAll", lostFoundPath).Return(fmt.Errorf("failed to remove lost+found"))

	// Expectations for rollback actions
	mockCmdRunner.On("RunCommand", "umount", mountPoint).Return(nil) // Expect unmount to be called during rollback
	mockFileOps.On("Remove", imageMountPoint).Return(nil)            // Expect file removal during rollback
	mockFileOps.On("RemoveAll", mountPoint).Return(nil)              // Expect directory removal during rollback

	// Execute the function that's being tested
	_, err := sm.CreateVolume(volumeID, 1)

	// Assert expected error and message
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove lost+found")

	// Verify that all expectations were met
	mockFileOps.AssertExpectations(t)
	mockCmdRunner.AssertExpectations(t)
}
