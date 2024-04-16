package api

import (
	"net/http"

	"0xKowalski1/container-orchestrator/models"
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

// POST /nodes
func joinCluster(c *gin.Context, _statemanager *statemanager.StateManager) {
	//AUTH

	var newNode models.CreateNodeRequest

	// Bind JSON from request body to newNode
	if err := c.BindJSON(&newNode); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node data"})
		return
	}

	// Check if the node already exists
	existingNode, err := _statemanager.GetNode(newNode.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check if node exists"})
		return
	}

	if existingNode != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Node already exists"})
		return
	}

	// Add the new node since it does not exist
	err = _statemanager.AddNode(newNode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add node"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Node added successfully"})
}
