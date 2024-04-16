package api

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"0xKowalski1/container-orchestrator/models"
	statemanager "0xKowalski1/container-orchestrator/state-manager"

	"github.com/gin-gonic/gin"
)

// GET /containers
func getContainers(c *gin.Context, _statemanager *statemanager.StateManager) {
	containers, err := _statemanager.ListContainers()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"containers": containers,
	})

}

// POST /containers
func createContainer(c *gin.Context, _statemanager *statemanager.StateManager) {
	var req models.CreateContainerRequest
	// Parse the JSON body to the CreateContainerRequest struct.
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	createdContainer, err := _statemanager.AddContainer(req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"container": createdContainer,
	})
}

// PATCH /containers
func updateContainer(c *gin.Context, _statemanager *statemanager.StateManager) {
	containerID := c.Param("id")

	var req models.UpdateContainerRequest
	// Parse the JSON body to the UpdateContainerRequest struct.
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err := _statemanager.PatchContainer(containerID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

// GET /containers/{id}
func getContainer(c *gin.Context, _statemanager *statemanager.StateManager) {
	containerID := c.Param("id")

	container, err := _statemanager.GetContainer(containerID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Container not found"})
		return
	}

	c.JSON(http.StatusOK, container)
}

// DELETE /containers/{id}
func deleteContainer(c *gin.Context, _statemanager *statemanager.StateManager) {
	containerID := c.Param("id")

	// Should mark for deletion!
	err := _statemanager.RemoveContainer(containerID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "true"})
}

// POST /containers/{id}/start
func startContainer(c *gin.Context, _statemanager *statemanager.StateManager) {
	containerID := c.Param("id")

	desiredStatus := "running"

	err := _statemanager.PatchContainer(containerID, models.UpdateContainerRequest{
		DesiredStatus: &desiredStatus,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Container starting"})
}

// POST /containers/{id}/stop
func stopContainer(c *gin.Context, _statemanager *statemanager.StateManager) {
	containerID := c.Param("id")

	desiredStatus := "stopped"

	err := _statemanager.PatchContainer(containerID, models.UpdateContainerRequest{
		DesiredStatus: &desiredStatus,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Container stopping"})
}

// GET /containers/{id}/logs
func streamContainerLogs(c *gin.Context, _statemanager *statemanager.StateManager) {
	containerID := c.Param("id")      // Retrieve the container ID from the URL parameter.
	namespace := c.Param("namespace") // GET ME FROM CONFIG

	// Get Container
	container, err := _statemanager.GetContainer(containerID)
	if err != nil {
		//handle
		log.Printf("Error getting container: %v", err)
	}

	// Get Node
	node, err := _statemanager.GetNode(container.NodeID)
	if err != nil {
		//handle
		log.Printf("Error getting node: %v", err)
	}

	workerAddress := fmt.Sprintf("http://%s:8081", node.NodeIp)

	// Construct the target URL to the worker node's log service.
	targetURL, err := url.Parse(workerAddress)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to parse worker address: %v", err)
		return
	}

	// Set up the reverse proxy.
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.URL.Path = "/namespaces/" + namespace + "/containers/" + containerID + "/logs"
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
	}

	// Use the reverse proxy to handle the request.
	proxy.ServeHTTP(c.Writer, c.Request)
}

// getContainerStatus streams the status of a container using Server-Sent Events (SSE).
func getContainerStatus(c *gin.Context, stateManager *statemanager.StateManager) {
	containerID := c.Param("id") // Retrieve the container ID from the URL parameter.

	// Set headers related to event streaming.
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	// Subscribe to the status updates for the specified container.
	statusChan, err := stateManager.SubscribeToStatus(containerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to subscribe to container status"})
		return
	}
	defer stateManager.UnsubscribeFromStatus(containerID, statusChan)

	// Use a ticker for sending heartbeats to keep the connection alive.
	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	// Loop for streaming updates.
	for {
		select {
		case status, ok := <-statusChan:
			if !ok {
				// Channel was closed, stop streaming.
				return
			}
			// Send the status update.
			c.SSEvent("status", status)
			c.Writer.Flush()
		case <-heartbeatTicker.C:
			// Send a comment as a heartbeat to keep the connection alive.
			c.SSEvent("", ":heartbeat")
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			// Client closed the connection, stop streaming.
			return
		}
	}
}
