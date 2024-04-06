package models

import "encoding/json"

type Container struct {
	ID            string
	DesiredStatus string // running or stopped
	Status        string
	NamespaceID   string
	NodeID        string
	Image         string
	Env           []string
	StopTimeout   int
}

// Container
type CreateContainerRequest struct {
	ID          string   `json:"id"`
	Image       string   `json:"image"`
	Env         []string `json:"env"`
	StopTimeout int      `json:"stopTimeout"`
}

type UpdateContainerRequest struct {
	DesiredStatus *string `json:"desiredStatus,omitempty"` // Pointer allows differentiation between an omitted field and an empty value
	NodeID        *string `json:"nodeId,omitempty"`
	Status        *string `json:"status,omitempty"`
}

func (c Container) Key() string {
	return "/namespaces/" + c.NamespaceID + "/containers/" + c.ID
}

func (c Container) Value() (string, error) {
	bytes, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
