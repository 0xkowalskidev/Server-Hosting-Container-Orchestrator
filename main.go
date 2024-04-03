package main

import (
	"log"

	"github.com/0xKowalski1/container-orchestrator/agent"
	"github.com/0xKowalski1/container-orchestrator/api"
	"github.com/0xKowalski1/container-orchestrator/runtime"
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

	// start controllers/managers

	// else worker node
	// start runtime
	_runtime, err := runtime.NewRuntime("containerd")

	if err != nil {
		log.Fatalf("Failed to initialize runtime: %v", err)
	}

	// start agent
	//temp join
	state.AddNode("node-1")

	agent.StartAgent(_runtime)

	// start networking
	// start local storage

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
		log.Fatalf("failed to list containers: %v", err)
	}

	for _, container := range containers {
		log.Println("Container ID:", container.ID)
	}

	ctr, err := _runtime.InspectContainer("example", _container.ID)
	if err != nil {
		log.Fatalf("failed to inspect container: %v", err)
	}
	log.Println("Container ID:", ctr.ID)

	err = _runtime.StopContainer("example", _container.ID, 5)
	if err != nil {
		log.Fatalf("failed to stop container: %v", err)
	}

	err = _runtime.RemoveContainer("example", _container.ID)
	if err != nil {
		log.Fatalf("failed to remove container: %v", err)
	}

}
