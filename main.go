package main

import (
	"fmt"
	"log"
	"os"

	"0xKowalski1/container-orchestrator/agent"
	"0xKowalski1/container-orchestrator/api"
	"0xKowalski1/container-orchestrator/config"
	"0xKowalski1/container-orchestrator/models"
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

	// for now nodes will be both controller and worker

	// if control node
	// start state manager
	_statemanager, err := statemanager.Start(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize statemanager: %v", err)
	}
	defer _statemanager.Close()

	// start api
	go api.Start(_statemanager)

	// start schedular
	schedular.Start(_statemanager)

	// start controllers/managers

	// else worker node
	//temp join, should be handled by agent through api calls
	_, err = _statemanager.GetNode("node-1")
	// 4 core, 16GB
	if err != nil {
		_statemanager.AddNode(models.Node{ID: "node-1", MemoryLimit: 16, CpuLimit: 4})
	}

	// start agent
	agent.Start(cfg)

	// start networking
	// start local storage

}
