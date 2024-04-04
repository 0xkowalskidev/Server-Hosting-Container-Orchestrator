package main

import (
	"log"

	"github.com/0xKowalski1/container-orchestrator/agent"
	"github.com/0xKowalski1/container-orchestrator/api"
	"github.com/0xKowalski1/container-orchestrator/runtime"
	"github.com/0xKowalski1/container-orchestrator/schedular"
	statemanager "github.com/0xKowalski1/container-orchestrator/state-manager"
)

func main() {
	// for now nodes will be both controller and worker

	// if control node
	// start state manager
	state := statemanager.Start()

	// start api
	go api.Start(state)

	// start schedular
	schedular.Start(state)

	// start controllers/managers

	// else worker node
	// start runtime
	_runtime, err := runtime.NewRuntime("containerd")

	if err != nil {
		log.Fatalf("Failed to initialize runtime: %v", err)
	}

	// start agent
	//temp join, should be handled by agent
	state.AddNode("node-1")

	agent.Start(_runtime)

	// start networking
	// start local storage

}
