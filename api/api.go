package api

import (
	"net/http"

	statemanager "github.com/0xKowalski1/container-orchestrator/state-manager"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Start(state *statemanager.State) {
	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowCredentials = true
	router.Use(cors.New(config))

	router.GET("containers", func(c *gin.Context) {
		getContainers(c, state)
	})

	router.GET("/containers/:id", func(c *gin.Context) {
		getContainer(c, state)
	})

	router.POST("/containers", func(c *gin.Context) {
		createContainer(c, state)
	})

	router.GET("nodes/:id/desired", func(c *gin.Context) {
		getNodeDesiredState(c, state)
	})

	router.Run()
}

// /containers

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

	// For this example, add the container to the list of UnscheduledContainers.
	// In a real system, you might instead trigger scheduling logic here.
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

	// First, search in unscheduled containers.
	for _, container := range state.UnscheduledContainers {
		if container.ID == containerID {
			c.JSON(http.StatusOK, container)
			return
		}
	}

	// If not found, search in each node's containers.
	for _, node := range state.Nodes {
		for _, container := range node.Containers {
			if container.ID == containerID {
				c.JSON(http.StatusOK, container)
				return
			}
		}
	}

	// Container not found.
	c.JSON(http.StatusNotFound, gin.H{"error": "Container not found"})
}

// DELETE /containers/{id}

// POST /containers/{id}/start

// POST /containers/{id}/stop

// GET /containers/{id}/logs

// /nodes

// GET /nodes

// GET /nodes/{id}

// GET /nodes/{id}/desired
func getNodeDesiredState(c *gin.Context, state *statemanager.State) {
	nodeID := c.Param("id")

	// Search for the node by ID.
	for _, node := range state.Nodes {
		if node.ID == nodeID {
			// Node found, return its containers as the desired state.
			c.JSON(http.StatusOK, gin.H{
				"containers": node.Containers,
			})
			return
		}
	}

	// Node not found.
	c.JSON(http.StatusNotFound, gin.H{
		"error": "Node not found",
	})
}

// /system

// GET /system/info

// GET /system/health
