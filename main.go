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

	_containerConfig := runtime.ContainerConfig{ID: "minecraft-server", Image: "docker.io/itzg/minecraft-server:latest", Env: []string{"EULA=TRUE"}}

	_container, err := _runtime.CreateContainer("example", _containerConfig)

	if err != nil {
		log.Fatalf("Failed to create container: %v", err)
	}

	err = _runtime.StartContainer("example", _container.ID)
	if err != nil {
		log.Fatalf("Failed to create container: %v", err)
	}

	containers, err := _runtime.ListContainers("example")
	if err != nil {
		log.Fatalf("Failed to list containers: %v", err)
	}

	for _, container := range containers {
		log.Println("Container ID:", container.ID)
	}
}
