package statemanager

import (
	"fmt"
)

// use etcd in the future, this is temp
type Node struct {
	ID         string
	Containers []Container
}

type Container struct {
	ID            string
	DesiredStatus string // running or stopped
	//MarkedForDeletion
	Status string
}

type State struct {
	Nodes                 []Node
	UnscheduledContainers []Container
	listeners             []Listener
}

type EventType string

const (
	NodeAdded        EventType = "NodeAdded"
	NodeRemoved      EventType = "NodeRemoved"
	ContainerAdded   EventType = "ContainerAdded"
	ContainerRemoved EventType = "ContainerRemoved"
)

type Event struct {
	Type EventType
	Data interface{}
}

type Listener func(Event)

func (s *State) Subscribe(listener Listener) {
	s.listeners = append(s.listeners, listener)
}

// emit broadcasts an event to all subscribers.
func (s *State) emit(event Event) {
	for _, listener := range s.listeners {
		listener(event)
	}
}

func Start() *State {
	state := &State{}

	return state
}

// GetNode finds a node from state
func (s *State) GetNode(nodeID string) (Node, error) {
	for _, node := range s.Nodes {
		if node.ID == nodeID {
			// Node found, return its containers as the desired state.
			return node, nil
		}
	}
	return Node{}, fmt.Errorf("Node not found at id: %s", nodeID)
}

// AddNode adds a new node to the state.
func (s *State) AddNode(nodeID string) {
	node := Node{ID: nodeID}
	s.Nodes = append(s.Nodes, node)

	s.emit(Event{Type: NodeAdded, Data: nodeID})
}

// RemoveNode removes a node from the state by ID.
func (s *State) RemoveNode(nodeID string) {
	for i, node := range s.Nodes {
		if node.ID == nodeID {
			s.Nodes = append(s.Nodes[:i], s.Nodes[i+1:]...)
			s.emit(Event{Type: NodeRemoved, Data: nodeID})

			return
		}
	}
}

// AddContainer adds a new container to the specified node.
func (s *State) AddContainer(containerID string) {
	container := Container{ID: containerID, DesiredStatus: "running"}
	s.UnscheduledContainers = append(s.UnscheduledContainers, container)
	s.emit(Event{Type: ContainerAdded, Data: container})
}

// GetContainer finds a container by id, by searching both containers and unscheduled containers
func (s *State) GetContainer(containerID string) (Container, error) {
	// First, search in unscheduled containers.
	for _, container := range s.UnscheduledContainers {
		if container.ID == containerID {
			return container, nil
		}
	}

	// If not found, search in each node's containers.
	for _, node := range s.Nodes {
		for _, container := range node.Containers {
			if container.ID == containerID {
				return container, nil
			}
		}
	}

	return Container{}, fmt.Errorf("Container not found at id: %s", containerID)
}

// RemoveContainer removes a container from the specified node.
func (s *State) RemoveContainer(containerID string) {
	for i, node := range s.Nodes {

		for j, container := range node.Containers {
			if container.ID == containerID {
				s.Nodes[i].Containers = append(s.Nodes[i].Containers[:j], s.Nodes[i].Containers[j+1:]...)
				s.emit(Event{Type: ContainerRemoved, Data: container.ID})
				return
			}

		}
	}
}

// RemoveUnscheduledContainer removes a container from UnscheduledContainers by ID.
func (s *State) RemoveUnscheduledContainer(containerID string) {
	for i, container := range s.UnscheduledContainers {
		if container.ID == containerID {
			s.UnscheduledContainers = append(s.UnscheduledContainers[:i], s.UnscheduledContainers[i+1:]...)
			return
		}
	}
}

type ContainerPatch struct {
	DesiredStatus *string `json:"desiredStatus,omitempty"` //Pointer allows differentiation between an omitted field and an empty value
}

func (s *State) PatchContainer(containerID string, patch ContainerPatch) (Container, error) {
	for i, node := range s.Nodes {
		for j, container := range node.Containers {
			if container.ID == containerID {
				if patch.DesiredStatus != nil {
					s.Nodes[i].Containers[j].DesiredStatus = *patch.DesiredStatus
				}
				return s.Nodes[i].Containers[j], nil // Successfully patched
			}
		}
	}

	for i, container := range s.UnscheduledContainers {
		if container.ID == containerID {
			if patch.DesiredStatus != nil {
				s.UnscheduledContainers[i].DesiredStatus = *patch.DesiredStatus
			}

			return s.UnscheduledContainers[i], nil // Successfully patched
		}
	}

	return Container{}, fmt.Errorf("container with ID %s not found", containerID)
}
