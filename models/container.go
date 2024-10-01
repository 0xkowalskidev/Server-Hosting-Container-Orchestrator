package models

import (
	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/utils"
)

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
	StorageLimit  int             `json:"storage_limit"`                        // Measured in GB
	DesiredStatus ContainerStatus `json:"desired_status" form:"desired_status"` // Desired status for container, node agent will try to match this in container runtime
}

func (c *Container) SetDefaults() {
	if c.DesiredStatus == "" {
		c.DesiredStatus = StatusRunning
	}

	c.StorageLimit = 2 // TODO: TEMP, remove
}

func (c *Container) Patch(patchContainer *Container) error {
	// These values can not be updated
	patchContainer.ID = c.ID
	patchContainer.NodeID = c.NodeID

	err := utils.MergeStructs(c, patchContainer)

	return err
}
