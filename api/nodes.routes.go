package api

import (
	"net/http"

	statemanager "0xKowalski1/container-orchestrator/state-manager"
	"github.com/gin-gonic/gin"
)

// GET /nodes
func getNodes(c *gin.Context, _statemanager *statemanager.StateManager) {
	nodes, err := _statemanager.ListNodes()

	if err != nil {
		//500
	}

	c.JSON(http.StatusOK, gin.H{
		"nodes": nodes,
	})
}

// GET /nodes/{id}
func getNode(c *gin.Context, _statemanager *statemanager.StateManager) {
	nodeID := c.Param("id")

	node, err := _statemanager.GetNode(nodeID)

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
