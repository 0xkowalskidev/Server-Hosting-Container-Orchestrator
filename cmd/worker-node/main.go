package main

import (
	"fmt"
	"os"

	"0xKowalski1/container-orchestrator/config"
	"0xKowalski1/container-orchestrator/utils"
	"0xKowalski1/container-orchestrator/worker-node/agent"
	"0xKowalski1/container-orchestrator/worker-node/networking"
	"0xKowalski1/container-orchestrator/worker-node/runtime"
	"0xKowalski1/container-orchestrator/worker-node/storage"
)

func main() {

	cfgPath := "config.json"
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	runtime, err := runtime.NewContainerdRuntime(cfg)

	if err != nil {
		fmt.Printf("Error initializing runtime: %v\n", err)
		os.Exit(1)
	}

	storage := storage.NewStorageManager(cfg, &utils.FileOps{}, &utils.CmdRunner{})

	networking := networking.NewNetworkingManager(cfg, &utils.CmdRunner{})

	agent := agent.NewAgent(cfg, runtime, storage, networking)

	// start agent
	agent.Start()
}
