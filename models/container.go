package models

type ContainerStatus string

const (
	StatusRunning ContainerStatus = "running"
	StatusStopped ContainerStatus = "stopped"
	StatusUnknown ContainerStatus = "unknown"
)

type Container struct {
	ID            string          `json:"id"`
	NodeID        string          `json:"node_id"`
	Image         string          `json:"image"`
	DesiredStatus ContainerStatus `json:"desired_status"` // Desired status for container, node agent will try to match this in container runtime
}

func (c *Container) SetDefaults() {
	if c.DesiredStatus == "" {
		c.DesiredStatus = StatusRunning
	}
}
