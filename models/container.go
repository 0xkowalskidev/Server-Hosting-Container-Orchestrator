package models

import "encoding/json"

type Port struct {
	HostPort      int    `json:"hostPort"`
	ContainerPort int    `json:"containerPort"`
	Protocol      string `json:"protocol"` // tcp or udp
}

// Used by the agent network syncer, needs to be tied to the container id
type Portmap struct {
	HostPort      int    `json:"hostPort"`
	ContainerPort int    `json:"containerPort"`
	Protocol      string `json:"protocol"` // tcp or udp
	ID            string `json:"id"`
}

type Container struct {
	ID            string
	DesiredStatus string // running or stopped
	Status        string
	NamespaceID   string
	NodeID        string
	Image         string
	Env           []string
	StopTimeout   int
	MemoryLimit   int
	CpuLimit      int
	StorageLimit  int
	Ports         []Port
}

// Container
type CreateContainerRequest struct {
	ID           string   `json:"id"`
	Image        string   `json:"image"`
	Env          []string `json:"env"`
	StopTimeout  int      `json:"stopTimeout"`
	MemoryLimit  int      `json:"memoryLimit"`
	CpuLimit     int      `json:"cpuLimit"`
	StorageLimit int      `json:"storageLimit"`
	Ports        []Port   `json:"ports"`
}

type UpdateContainerRequest struct {
	DesiredStatus *string `json:"desiredStatus,omitempty"` // Pointer allows differentiation between an omitted field and an empty value
	NodeID        *string `json:"nodeId,omitempty"`
	Status        *string `json:"status,omitempty"`
	MemoryLimit   int     `json:"memoryLimit"`
	CpuLimit      int     `json:"cpuLimit"`
	StorageLimit  int     `json:"storageLimit"`
	Ports         []Port  `json:"ports"`
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
