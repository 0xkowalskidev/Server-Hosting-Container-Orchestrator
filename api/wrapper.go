package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	statemanager "0xKowalski1/container-orchestrator/state-manager"
)

// API base URL. Adjust as needed.
const BaseURL = "http://localhost:8080"

// Client represents the API client
type WrapperClient struct {
	HTTPClient *http.Client
}

// NewClient creates a new API client
func NewApiWrapper() *WrapperClient {
	return &WrapperClient{
		HTTPClient: &http.Client{},
	}
}

type ContainerListResponse struct {
	Containers []statemanager.Container `json:"containers"`
}

// CreateContainer creates a new container in the specified namespace
func (c *WrapperClient) CreateContainer(namespace string, req CreateContainerRequest) (*statemanager.Container, error) {
	requestBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/namespaces/%s/containers", BaseURL, namespace)
	response, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API request failed with status code %d", response.StatusCode)
	}

	type ContainerResponse struct {
		Container statemanager.Container `json:"container"`
	}

	var containerResponse ContainerResponse
	if err := json.NewDecoder(response.Body).Decode(&containerResponse); err != nil {
		return nil, err
	}

	fmt.Printf("Container Response: %+v\n", containerResponse.Container)

	return &containerResponse.Container, nil
}

func (c *WrapperClient) ListContainers(namespace string) ([]statemanager.Container, error) {
	url := fmt.Sprintf("%s/namespaces/%s/containers", BaseURL, namespace)
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

func (c *WrapperClient) DeleteContainer(namespace string, containerID string) (string, error) {
	url := fmt.Sprintf("%s/namespaces/%s/containers/%s", BaseURL, namespace, containerID)
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

func (c *WrapperClient) StartContainer(namespace string, containerID string) (string, error) {
	url := fmt.Sprintf("%s/namespaces/%s/containers/%s/start", BaseURL, namespace, containerID)
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

func (c *WrapperClient) StopContainer(namespace string, containerID string) (string, error) {
	url := fmt.Sprintf("%s/namespaces/%s/containers/%s/stop", BaseURL, namespace, containerID)
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
