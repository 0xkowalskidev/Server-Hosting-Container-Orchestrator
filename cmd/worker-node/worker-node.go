package main

import (
	"fmt"
	"os"

	"0xKowalski1/container-orchestrator/agent"
	"0xKowalski1/container-orchestrator/config"
	"0xKowalski1/container-orchestrator/networking"
	"0xKowalski1/container-orchestrator/runtime"
	"0xKowalski1/container-orchestrator/storage"
	"0xKowalski1/container-orchestrator/utils"
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
