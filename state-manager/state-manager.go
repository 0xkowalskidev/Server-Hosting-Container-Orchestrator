package statemanager

import ()

// use etcd in the future, this is temp
type Node struct {
}

type Container struct {
}

type State struct {
	Nodes      []Node
	Containers []Container
}

func Start() State {
	state := State{}

	return state
}
