package agent_test

import (
	"0xKowalski1/container-orchestrator/agent"
	"0xKowalski1/container-orchestrator/config"
	"0xKowalski1/container-orchestrator/storage"
	utils_test "0xKowalski1/container-orchestrator/tests/utils"
)

var fakeStoragePath = "/fakepath"

func setup() (*agent.Agent, *utils_test.MockFileOps, *utils_test.MockCmdRunner) {
	cfg := &config.Config{StoragePath: fakeStoragePath}
	mockFileOps := new(utils_test.MockFileOps)
	mockCmdRunner := new(utils_test.MockCmdRunner)

	storageManager, fileOps, cmdRunner := storage.NewStorageManager(cfg, mockFileOps, mockCmdRunner), mockFileOps, mockCmdRunner

	agent = agent.Agent{
		storage: &storageManager,
	}

	return agent, fileOps, cmdRunner
}
