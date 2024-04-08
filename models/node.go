package models

import "encoding/json"

type Node struct {
	ID          string      `json:"id"`
	Containers  []Container `json:"containers"`
	MemoryLimit int         `json:"memoryLimit"`
	CpuLimit    int         `json:"cpuLimit"`

	MemoryUsed int `json:"memoryUsed"` // Not to be persisted to etcd
	CpuUsed    int `json:"cpuUsed"`    // Not to be persisted to etcd
}

func (n Node) Key() string {
	return "/nodes/" + n.ID
}

func (n Node) Value() (string, error) {
	serializedNode := struct {
		ID          string      `json:"id"`
		Containers  []Container `json:"containers"`
		MemoryLimit int         `json:"memoryLimit"`
		CpuLimit    int         `json:"cpuLimit"`
	}{
		ID:          n.ID,
		Containers:  n.Containers,
		MemoryLimit: n.MemoryLimit,
		CpuLimit:    n.CpuLimit,
	}

	bytes, err := json.Marshal(serializedNode)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
