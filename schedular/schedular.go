package schedular

import (
	"log"

	statemanager "github.com/0xKowalski1/container-orchestrator/state-manager"
)

func Start(state *statemanager.State) {
	state.Subscribe(func(event statemanager.Event) {
		switch event.Type {
		case statemanager.ContainerAdded:
			scheduleContainers(state)
		}
	})
}

// scheduleContainers checks and updates the state to schedule containers to nodes.
func scheduleContainers(state *statemanager.State) {
	for _, container := range state.UnscheduledContainers {
		assigned := false
		for i, node := range state.Nodes {
			if len(node.Containers) < 2 { // Arbitrary rule
				state.Nodes[i].Containers = append(state.Nodes[i].Containers, container)
				assigned = true
				state.RemoveUnscheduledContainer(container.ID)
				break
			}
		}
		if !assigned {
			log.Printf("Failed to assign container %s, all nodes are at capacity", container.ID)
		}
	}
}
