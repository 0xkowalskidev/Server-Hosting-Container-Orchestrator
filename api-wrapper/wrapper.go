package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"0xKowalski1/container-orchestrator/models"
)

const BaseURL = "http://localhost:8080"

// Client represents the API client
type WrapperClient struct {
	HTTPClient *http.Client
	BaseURL    string
}

// NewClient creates a new API client
func NewApiWrapper(baseUrl string) *WrapperClient {
	return &WrapperClient{
		HTTPClient: &http.Client{Timeout: 0},
		BaseURL:    fmt.Sprintf("http://%s:8080", baseUrl),
	}
}

type ContainerListResponse struct {
	Containers []models.Container `json:"containers"`
}

// CreateContainer creates a new container
func (c *WrapperClient) CreateContainer(req models.CreateContainerRequest) (*models.Container, error) {
	requestBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/containers", c.BaseURL)
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

// UpdateContainer updates an existing container's configuration.
func (c *WrapperClient) UpdateContainer(containerID string, req models.UpdateContainerRequest) (*models.Container, error) {
	requestBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/containers/%s", c.BaseURL, containerID)
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
	url := fmt.Sprintf("%s/containers", c.BaseURL)
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

func (c *WrapperClient) GetContainer(containerID string) (*models.Container, error) {
	url := fmt.Sprintf("%s/containers/%s", c.BaseURL, containerID)
	response, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code %d", response.StatusCode)
	}

	var resp models.Container
	if err := json.NewDecoder(response.Body).Decode(&resp); err != nil {
		return nil, err
	}

	return &resp, nil // Return the container
}

func (c *WrapperClient) DeleteContainer(containerID string) error {
	url := fmt.Sprintf("%s/containers/%s", c.BaseURL, containerID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}

	response, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed with status code %d", response.StatusCode)
	}

	return nil
}

func (c *WrapperClient) StartContainer(containerID string) (string, error) {
	url := fmt.Sprintf("%s/containers/%s/start", c.BaseURL, containerID)
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
	url := fmt.Sprintf("%s/containers/%s/stop", c.BaseURL, containerID)
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
	url := fmt.Sprintf("%s/nodes/%s", c.BaseURL, nodeID)
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

	return &resp.Node, nil // Return the node
}

type NodeListResponse struct {
	Nodes []models.Node `json:"nodes"`
}

func (c *WrapperClient) ListNodes() ([]models.Node, error) {
	url := fmt.Sprintf("%s/nodes", c.BaseURL)
	response, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code %d", response.StatusCode)
	}

	var resp NodeListResponse
	if err := json.NewDecoder(response.Body).Decode(&resp); err != nil {
		return nil, err
	}

	return resp.Nodes, nil
}

func (c *WrapperClient) WatchContainer(containerID string, handleData func(string)) error {
	url := fmt.Sprintf("%s/containers/%s/watch", c.BaseURL, containerID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break // Stream closed normally
			}
			return err // Stream error
		}

		handleData(string(line))
	}

	return nil
}

func (c *WrapperClient) StreamContainerLogs(containerID string, handleData func(string)) error {
	url := fmt.Sprintf("%s/containers/%s/logs", c.BaseURL, containerID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break // Stream closed normally.
			}
			return err // Handle errors during the stream.
		}

		handleData(string(line))
	}

	return nil
}

// JoinCluster attempts to join a worker node to the cluster
func (c *WrapperClient) JoinCluster(req models.CreateNodeRequest) (*models.Node, error) {
	requestBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/nodes", c.BaseURL)
	response, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read response body: %v", readErr)
	}

	if response.StatusCode != http.StatusOK {
		apiError := struct {
			Error string `json:"error"`
		}{}
		if err := json.Unmarshal(body, &apiError); err != nil {
			return nil, fmt.Errorf("API request failed with status code %d: unable to parse API error response", response.StatusCode)
		}
		return nil, fmt.Errorf("API request failed: %s", apiError.Error)
	}

	var nodeResponse struct {
		Node models.Node `json:"node"`
	}
	if err := json.Unmarshal(body, &nodeResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &nodeResponse.Node, nil
}
