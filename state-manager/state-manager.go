package statemanager

import ()

// use etcd in the future, this is temp
type Node struct {
	ID         string
	Containers []Container
}

type Container struct {
	ID string
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
	container := Container{ID: containerID}
	s.UnscheduledContainers = append(s.UnscheduledContainers, container)
	s.emit(Event{Type: ContainerAdded, Data: container})
}

// RemoveContainer removes a container from the specified node.
func (s *State) RemoveContainer(nodeID, containerID string) {
	for i, node := range s.Nodes {
		if node.ID == nodeID {
			for j, container := range node.Containers {
				if container.ID == containerID {
					s.Nodes[i].Containers = append(s.Nodes[i].Containers[:j], s.Nodes[i].Containers[j+1:]...)
					s.emit(Event{Type: ContainerRemoved, Data: container.ID})
					return
				}
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
