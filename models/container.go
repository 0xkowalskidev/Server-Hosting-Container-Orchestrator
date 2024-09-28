package models

type Container struct {
	ID        string `json:"id"`
	NodeID    string `json:"node_id"`
	Namespace string `json:"namespace"`
	Image     string `json:"image"`
}
