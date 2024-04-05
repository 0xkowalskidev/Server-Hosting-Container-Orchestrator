package main

import (
	"log"

	"0xKowalski1/container-orchestrator/agent"
	"0xKowalski1/container-orchestrator/api"
	"0xKowalski1/container-orchestrator/runtime"
	"0xKowalski1/container-orchestrator/schedular"
	statemanager "0xKowalski1/container-orchestrator/state-manager"
)

func main() {
	// for now nodes will be both controller and worker

	// if control node
	// start state manager
	_statemanager, err := statemanager.Start()
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
	// start runtime
	_runtime, err := runtime.NewRuntime("containerd")

	if err != nil {
		log.Fatalf("Failed to initialize runtime: %v", err)
	}

	//temp join, should be handled by agent
	_, err = _statemanager.GetNode("node-1")
	if err != nil {
		_statemanager.AddNode(statemanager.Node{ID: "node-1"})
	}

	// start agent
	agent.Start(_runtime)

	// start networking
	// start local storage

}
