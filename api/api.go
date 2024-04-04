package api

import (
	"net/http"

	statemanager "github.com/0xKowalski1/container-orchestrator/state-manager"
	"github.com/gin-gonic/gin"
)

func Start(state *statemanager.State) {
	router := gin.Default()

	// Should be post
	router.GET("/containers", func(c *gin.Context) {
		createContainer(c, state)
	})

	router.GET("/nodes/:id/desired", func(c *gin.Context) {
		getNodeDesiredState(c, state)
	})

	router.Run("localhost:8080")
}

// /containers

// GET /containers

// POST /containers
func createContainer(c *gin.Context, state *statemanager.State) {
	state.AddContainer("minecraft-server")
}

// GET /containers/{id}

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
