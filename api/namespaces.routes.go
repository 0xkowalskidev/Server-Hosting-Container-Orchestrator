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
