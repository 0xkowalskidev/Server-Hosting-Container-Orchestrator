package controlpanel

import (
	"fmt"

	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/models"
	"github.com/go-resty/resty/v2"
)

type Wrapper struct {
	baseURL string
	client  *resty.Client
}

func NewWrapper(baseURL string) *Wrapper {
	client := resty.New()
	client.SetBaseURL(baseURL)
	client.SetHeader("Content-Type", "application/json")

	return &Wrapper{
		baseURL: baseURL,
		client:  client,
	}
}

func (w *Wrapper) get(endpoint string, result interface{}) error {
	resp, err := w.client.R().
		SetResult(result).
		Get(endpoint)
	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("error: %s", resp.Status())
	}

	return nil
}

func (w *Wrapper) post(endpoint string, body interface{}, result interface{}) error {
	resp, err := w.client.R().
		SetBody(body).
		SetResult(result).
		Post(endpoint)
	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("error: %s", resp.Status())
	}

	return nil
}

func (w *Wrapper) patch(endpoint string, resourceID string, body interface{}, result interface{}) error {
	resp, err := w.client.R().
		SetBody(body).
		SetResult(result).
		Patch(fmt.Sprintf("%s/%s", endpoint, resourceID))
	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("error: %s", resp.Status())
	}

	return nil
}

// Delete is reserved
func (w *Wrapper) remove(endpoint string, resourceID string) error {
	resp, err := w.client.R().Delete(fmt.Sprintf("%s/%s", endpoint, resourceID))

	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("error: %s", resp.Status())
	}

	return nil
}

// Containers
func (w *Wrapper) GetContainers() ([]models.Container, error) {
	var containers []models.Container

	err := w.get("/containers", &containers)
	if err != nil {
		return containers, err
	}

	return containers, nil
}

func (c *Wrapper) CreateContainer(body models.Container) (models.Container, error) {
	var container models.Container

	err := c.post("/containers", body, &container)
	if err != nil {
		return container, err
	}

	return container, nil
}

func (c *Wrapper) UpdateContainer(containerID string, body models.Container) (models.Container, error) {
	var container models.Container

	err := c.patch("/containers", containerID, body, &container)
	if err != nil {
		return container, err
	}

	return container, nil
}

func (c *Wrapper) DeleteContainer(containerID string) error {
	err := c.remove("/containers", containerID)
	if err != nil {
		return err
	}

	return nil
}

// Nodes
func (w *Wrapper) GetNodes() ([]models.Node, error) {
	var nodes []models.Node

	err := w.get("/nodes", &nodes)
	if err != nil {
		return nodes, err
	}

	return nodes, nil
}
