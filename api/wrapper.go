package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"0xKowalski1/container-orchestrator/models"
)

// API base URL. Adjust as needed.
const BaseURL = "http://localhost:8080"

// Client represents the API client
type WrapperClient struct {
	HTTPClient *http.Client
	Namespace  string
}

// NewClient creates a new API client
func NewApiWrapper(namespace string) *WrapperClient {
	return &WrapperClient{
		HTTPClient: &http.Client{},
		Namespace:  namespace,
	}
}

type ContainerListResponse struct {
	Containers []models.Container `json:"containers"`
}

// CreateContainer creates a new container in the specified namespace
func (c *WrapperClient) CreateContainer(req models.CreateContainerRequest) (*models.Container, error) {
	requestBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/namespaces/%s/containers", BaseURL, c.Namespace)
	response, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API request failed with status code %d", response.StatusCode)
	}

	type ContainerResponse struct {
		Container models.Container `json:"container"`
	}

	var containerResponse ContainerResponse
	if err := json.NewDecoder(response.Body).Decode(&containerResponse); err != nil {
		return nil, err
	}

	fmt.Printf("Container Response: %+v\n", containerResponse.Container)

	return &containerResponse.Container, nil
}

// UpdateContainer updates an existing container's configuration in the specified namespace.
func (c *WrapperClient) UpdateContainer(containerID string, req models.UpdateContainerRequest) (*models.Container, error) {
	requestBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/namespaces/%s/containers/%s", BaseURL, c.Namespace, containerID)
	request, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := c.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code %d", response.StatusCode)
	}

	var containerResponse struct {
		Container models.Container `json:"container"`
	}

	if err := json.NewDecoder(response.Body).Decode(&containerResponse); err != nil {
		return nil, err
	}

	fmt.Printf("Updated Container Response: %+v\n", containerResponse.Container)

	return &containerResponse.Container, nil
}

func (c *WrapperClient) ListContainers() ([]models.Container, error) {
	url := fmt.Sprintf("%s/namespaces/%s/containers", BaseURL, c.Namespace)
	response, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code %d", response.StatusCode)
	}

	var resp ContainerListResponse // Adjusted to use the new wrapper struct
	if err := json.NewDecoder(response.Body).Decode(&resp); err != nil {
		return nil, err
	}

	return resp.Containers, nil // Return the slice of containers
}

func (c *WrapperClient) DeleteContainer(containerID string) (string, error) {
	url := fmt.Sprintf("%s/namespaces/%s/containers/%s", BaseURL, c.Namespace, containerID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return containerID, err
	}

	response, err := c.HTTPClient.Do(req)
	if err != nil {
		return containerID, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return containerID, fmt.Errorf("API request failed with status code %d", response.StatusCode)
	}

	return containerID, nil
}

func (c *WrapperClient) StartContainer(containerID string) (string, error) {
	url := fmt.Sprintf("%s/namespaces/%s/containers/%s/start", BaseURL, c.Namespace, containerID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return containerID, err
	}

	response, err := c.HTTPClient.Do(req)
	if err != nil {
		return containerID, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return containerID, fmt.Errorf("API request failed with status code %d", response.StatusCode)
	}

	return containerID, nil
}

func (c *WrapperClient) StopContainer(containerID string) (string, error) {
	url := fmt.Sprintf("%s/namespaces/%s/containers/%s/stop", BaseURL, c.Namespace, containerID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return containerID, err
	}

	response, err := c.HTTPClient.Do(req)
	if err != nil {
		return containerID, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return containerID, fmt.Errorf("API request failed with status code %d", response.StatusCode)
	}

	return containerID, nil
}

type NodeResponse struct {
	Node models.Node `json:"node"`
}

func (c *WrapperClient) GetNode(nodeID string) (*models.Node, error) {
	url := fmt.Sprintf("%s/nodes/%s", BaseURL, nodeID)
	response, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code %d", response.StatusCode)
	}

	var resp NodeResponse // Adjusted to use the new wrapper struct
	if err := json.NewDecoder(response.Body).Decode(&resp); err != nil {
		return nil, err
	}

	return &resp.Node, nil // Return the slice of containers

}
