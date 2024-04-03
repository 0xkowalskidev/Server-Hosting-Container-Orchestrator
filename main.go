package main

import (
	"log"

	"github.com/0xKowalski1/container-orchestrator/runtime"
)

func main() {
	// for now nodes will be both controller and worker
	//state := statemanager.Start()
	//api.Start()

	// if control node
	// start api
	// start schedular
	// start controllers/managers
	// start state manager

	// else worker node
	// start runtime
	// start agent
	// start networking
	// start local storage

	// Create runtime
	_runtime, err := runtime.NewRuntime("containerd")

	if err != nil {
		log.Fatalf("Failed to initialize runtime: %v", err)
	}

	containers, err := _runtime.ListContainers("example")
	if err != nil {
		log.Fatalf("Failed to list containers: %v", err)
	}

	for _, container := range containers {
		log.Println("Container ID:", container.ID)
	}
}
