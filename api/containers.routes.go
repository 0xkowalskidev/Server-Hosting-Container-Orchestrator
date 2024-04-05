package api

import (
	"net/http"

	statemanager "github.com/0xKowalski1/container-orchestrator/state-manager"
	"github.com/gin-gonic/gin"
)

// GET /containers
func getContainers(c *gin.Context, state *statemanager.State) {
	var allContainers []statemanager.Container

	// Iterate over all nodes and aggregate their containers.
	for _, node := range state.Nodes {
		allContainers = append(allContainers, node.Containers...)
	}

	// Return the list of all containers as JSON.
	c.JSON(http.StatusOK, gin.H{
		"containers":            allContainers,
		"unscheduledContainers": state.UnscheduledContainers,
	})
}

// POST /containers
type CreateContainerRequest struct {
	ID string `json:"id"` // Include other fields as necessary, e.g., image name.
}

func createContainer(c *gin.Context, state *statemanager.State) {
	var req CreateContainerRequest
	// Parse the JSON body to the CreateContainerRequest struct.
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	state.AddContainer(req.ID)

	// Respond to indicate successful container creation.
	c.JSON(http.StatusCreated, gin.H{
		"message":     "Container created",
		"containerID": req.ID,
	})
}

// GET /containers/{id}
func getContainer(c *gin.Context, state *statemanager.State) {
	containerID := c.Param("id") // Retrieve the container ID from the URL parameter.

	container, err := state.GetContainer(containerID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Container not found"})
	} else {
		c.JSON(http.StatusOK, container)
	}

}

// DELETE /containers/{id}
func deleteContainer(c *gin.Context, state *statemanager.State) {
	containerID := c.Param("id") // Retrieve the container ID from the URL parameter.

	// Should mark for deletion!
	state.RemoveContainer(containerID)
	state.RemoveUnscheduledContainer(containerID)

	c.JSON(http.StatusOK, gin.H{"success": "true"})
}

// POST /containers/{id}/start
func startContainer(c *gin.Context, state *statemanager.State) {
	containerID := c.Param("id") // Retrieve the container ID from the URL parameter.

	desiredStatus := "running"

	patchedContainer, err := state.PatchContainer(containerID, statemanager.ContainerPatch{
		DesiredStatus: &desiredStatus,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Container starting", "container": patchedContainer})
}

// POST /containers/{id}/stop
func stopContainer(c *gin.Context, state *statemanager.State) {
	containerID := c.Param("id") // Retrieve the container ID from the URL parameter.

	desiredStatus := "stopped"

	patchedContainer, err := state.PatchContainer(containerID, statemanager.ContainerPatch{
		DesiredStatus: &desiredStatus,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Container stopping", "container": patchedContainer})
}

// GET /containers/{id}/logs
func getContainerLogs(c *gin.Context, state *statemanager.State) {
	// containerID := c.Param("id") // Retrieve the container ID from the URL parameter.
}
