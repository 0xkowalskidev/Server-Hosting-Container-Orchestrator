package schedular

import (
	"log"

	statemanager "github.com/0xKowalski1/container-orchestrator/state-manager"
)

func Start(_statemanager *statemanager.StateManager) {
	// Run schedule once on start, incase there where missed unscheduled events while offlien
	scheduleContainers(_statemanager)

	_statemanager.Subscribe(func(event statemanager.Event) {
		switch event.Type {
		case statemanager.ContainerAdded:
			scheduleContainers(_statemanager)
		}
	})
}

// scheduleContainers checks and updates the _statemanager to schedule containers to nodes.
func scheduleContainers(sm *statemanager.StateManager) {
	unscheduledContainers, err := sm.ListUnscheduledContainers()
	if err != nil {
		log.Printf("Error fetching unscheduled containers: %v", err)
		return
	}

	nodes, err := sm.ListNodes()
	if err != nil {
		log.Printf("Error fetching nodes: %v", err)
		return
	}

	log.Printf("Starting container scheduling...")
	for _, container := range unscheduledContainers {
		log.Printf("Attempting to schedule container %s...", container.ID)
		assigned := false
		for _, node := range nodes {
			if len(node.Containers) < 2 { // Arbitrary scheduling rule

				err := sm.AssignContainerToNode(container.NamespaceID, container.ID, node.ID)
				if err != nil {
					log.Printf("Failed to assign container %s to node %s: %v", container.ID, node.ID, err)
					continue
				}
				assigned = true
				break
			}
		}
		if !assigned {
			log.Printf("Failed to assign container %s, all nodes are at capacity", container.ID)
		}
	}
}
