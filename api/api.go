package api

import (
	statemanager "0xKowalski1/container-orchestrator/state-manager"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Start initializes the HTTP server and routes
func Start(_statemanagermanager *statemanager.StateManager) {
	router := gin.Default()
	setupMiddlewares(router)
	setupRoutes(router, _statemanagermanager)
	router.Run()
}

// setupMiddlewares configures any global middlewares
func setupMiddlewares(router *gin.Engine) {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowCredentials = true
	router.Use(cors.New(config))
}

// setupRoutes configures the routes for the API
func setupRoutes(router *gin.Engine, _statemanager *statemanager.StateManager) {

	// Namespace routes
	namespacesGroup := router.Group("/namespaces")
	{
		namespacesGroup.GET("", func(c *gin.Context) { getNamespaces(c, _statemanager) })           // Lists all namespaces
		namespacesGroup.GET("/:namespace", func(c *gin.Context) { getNamespace(c, _statemanager) }) // Retrieves a specific namespace
		namespacesGroup.POST("", func(c *gin.Context) { createNamespace(c, _statemanager) })        // Creates a new namespace

		// Nested container routes
		containersGroup := namespacesGroup.Group("/:namespace/containers")
		{
			containersGroup.GET("", func(c *gin.Context) { getContainers(c, _statemanager) })             // Lists all containers in the specified namespace
			containersGroup.GET("/:id", func(c *gin.Context) { getContainer(c, _statemanager) })          // Retrieves a specific container in the specified namespace
			containersGroup.POST("", func(c *gin.Context) { createContainer(c, _statemanager) })          // Creates a new container in the specified namespace
			containersGroup.DELETE("/:id", func(c *gin.Context) { deleteContainer(c, _statemanager) })    // Deletes a specific container in the specified namespace
			containersGroup.PATCH("/:id", func(c *gin.Context) { updateContainer(c, _statemanager) })     // Updates a specific container in the specified namespace
			containersGroup.POST("/:id/start", func(c *gin.Context) { startContainer(c, _statemanager) }) // Starts a specific container in the specified namespace
			containersGroup.POST("/:id/stop", func(c *gin.Context) { stopContainer(c, _statemanager) })   // Stops a specific container in the specified namespace
			containersGroup.GET("/:id/logs", func(c *gin.Context) { getContainerLogs(c, _statemanager) }) // Retrieves logs for a specific container in the specified namespace
		}
	}

	// Node routes
	nodeGroup := router.Group("/nodes")
	{
		nodeGroup.GET("", func(c *gin.Context) { getNodes(c, _statemanager) })
		nodeGroup.GET("/:id", func(c *gin.Context) { getNode(c, _statemanager) })
	}
}
