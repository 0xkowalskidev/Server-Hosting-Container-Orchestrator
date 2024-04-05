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

	node, err := state.GetNode(nodeID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Node not found",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"node": node,
		})
	}
}
