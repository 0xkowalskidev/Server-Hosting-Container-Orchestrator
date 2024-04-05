package api

import (
	"net/http"

	statemanager "github.com/0xKowalski1/container-orchestrator/state-manager"
	"github.com/gin-gonic/gin"
)

// GET /nodes
func getNodes(c *gin.Context, state *statemanager.State) {
	c.JSON(http.StatusOK, gin.H{
		"nodes": state.Nodes,
	})
}

// GET /nodes/{id}
func getNode(c *gin.Context, state *statemanager.State) {
	nodeID := c.Param("id")

	for _, node := range state.Nodes {
		if node.ID == nodeID {
			// Node found, return its containers as the desired state.
			c.JSON(http.StatusOK, gin.H{
				"node": node,
			})
			return
		}
	}

	// Node not found.
	c.JSON(http.StatusNotFound, gin.H{
		"error": "Node not found",
	})
}
