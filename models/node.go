package models

type Node struct {
	ID         string      `json:"id"`
	Namespace  string      `json:"namespace"`  // Containerd Namespace
	Containers []Container `json:"containers"` // Storing the entire container in the node struct to reduce api calls, container only stores node id so circular dependcy is avoided
}
