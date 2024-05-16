package main

import (
	"fmt"
	"log"
	"os"

	"0xKowalski1/container-orchestrator/config"
	"0xKowalski1/container-orchestrator/control-node/api"
	"0xKowalski1/container-orchestrator/control-node/schedular"
	statemanager "0xKowalski1/container-orchestrator/control-node/state-manager"
)

func main() {
	cfgPath := "config.json"
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// start state manager
	stateManager, err := statemanager.NewStateManager(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize statemanager: %v", err)
	}
	defer stateManager.Close()

	// start api
	go api.Start(stateManager)

	// start schedular
	go schedular.Start(stateManager)

	// Block main from ending
	done := make(chan struct{})
	<-done
}
