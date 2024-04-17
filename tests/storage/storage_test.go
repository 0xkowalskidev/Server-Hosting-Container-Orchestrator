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

func TestStorageManager_RemoveNonExistantVolume(t *testing.T) {
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
