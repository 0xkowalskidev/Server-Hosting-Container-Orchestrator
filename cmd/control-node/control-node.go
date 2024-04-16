package main

import (
	"fmt"
	"log"
	"os"

	"0xKowalski1/container-orchestrator/api"
	"0xKowalski1/container-orchestrator/config"
	"0xKowalski1/container-orchestrator/schedular"
	statemanager "0xKowalski1/container-orchestrator/state-manager"
)

func main() {
	cfgPath := "config.json"
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// start state manager
	_statemanager, err := statemanager.Start(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize statemanager: %v", err)
	}
	defer _statemanager.Close()

	// start api
	go api.Start(_statemanager)

	// start schedular
	go schedular.Start(_statemanager)

	// Block main from ending
	done := make(chan struct{})
	<-done
}
