package api

import (
	"net/http"

	statemanager "0xKowalski1/container-orchestrator/state-manager"
	"github.com/gin-gonic/gin"
)

// GET /namespaces
func getNamespaces(c *gin.Context, _statemanager *statemanager.StateManager) {
	namespaces, err := _statemanager.ListNamespaces()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"namespaces": namespaces,
	})

}

// POST /namespaces
type CreateNamespaceRequest struct {
	ID string `json:"id"`
}

func createNamespace(c *gin.Context, _statemanager *statemanager.StateManager) {
	var req CreateNamespaceRequest
	// Parse the JSON body to the CreateNamespaceRequest struct.
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err := _statemanager.AddNamespace(statemanager.Namespace{ID: req.ID})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Namespace created",
		"containerID": req.ID,
	})
}

// GET /namespaces/{id}
func getNamespace(c *gin.Context, _statemanager *statemanager.StateManager) {
	namespaceID := c.Param("namespace")

	namespace, err := _statemanager.GetNamespace(namespaceID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Namespace not found"})
		return
	}

	c.JSON(http.StatusOK, namespace)
}
