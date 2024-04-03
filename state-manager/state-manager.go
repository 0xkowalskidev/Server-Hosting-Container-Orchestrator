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
	Nodes []Node
}

func Start() *State {
	state := &State{}

	return state
}

// AddNode adds a new node to the state.
func (s *State) AddNode(nodeID string) {
	s.Nodes = append(s.Nodes, Node{ID: nodeID})
}

// RemoveNode removes a node from the state by ID.
func (s *State) RemoveNode(nodeID string) {
	for i, node := range s.Nodes {
		if node.ID == nodeID {
			s.Nodes = append(s.Nodes[:i], s.Nodes[i+1:]...)
			return
		}
	}
}

// AddContainer adds a new container to the specified node.
func (s *State) AddContainer(nodeID, containerID string) {
	for i, node := range s.Nodes {
		if node.ID == nodeID {
			s.Nodes[i].Containers = append(s.Nodes[i].Containers, Container{ID: containerID})
			return
		}
	}
}

// RemoveContainer removes a container from the specified node.
func (s *State) RemoveContainer(nodeID, containerID string) {
	for i, node := range s.Nodes {
		if node.ID == nodeID {
			for j, container := range node.Containers {
				if container.ID == containerID {
					s.Nodes[i].Containers = append(s.Nodes[i].Containers[:j], s.Nodes[i].Containers[j+1:]...)
					return
				}
			}
		}
	}
}
