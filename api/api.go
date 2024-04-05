package api

import (
	statemanager "github.com/0xKowalski1/container-orchestrator/state-manager"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Start initializes the HTTP server and routes
func Start(state *statemanager.State) {
	router := gin.Default()
	setupMiddlewares(router)
	setupRoutes(router, state)
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
func setupRoutes(router *gin.Engine, state *statemanager.State) {
	// Container routes
	containerGroup := router.Group("/containers")
	{
		containerGroup.GET("", func(c *gin.Context) { getContainers(c, state) })
		containerGroup.GET("/:id", func(c *gin.Context) { getContainer(c, state) })
		containerGroup.POST("", func(c *gin.Context) { createContainer(c, state) })
		containerGroup.DELETE("/:id", func(c *gin.Context) { deleteContainer(c, state) })
		containerGroup.POST("/:id/start", func(c *gin.Context) { startContainer(c, state) })
		containerGroup.POST("/:id/stop", func(c *gin.Context) { stopContainer(c, state) })
		containerGroup.GET("/:id/logs", func(c *gin.Context) { getContainerLogs(c, state) })
	}

	// Node routes
	nodeGroup := router.Group("/nodes")
	{
		nodeGroup.GET("", func(c *gin.Context) { getNodes(c, state) })
		nodeGroup.GET("/:id", func(c *gin.Context) { getNode(c, state) })
	}
}
