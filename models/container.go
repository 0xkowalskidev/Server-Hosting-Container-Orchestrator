package models

type Container struct {
	ID     string `json:"id"`
	NodeID string `json:"node_id"`
	Image  string `json:"image"`
}
