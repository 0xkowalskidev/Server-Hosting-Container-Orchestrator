package main

import (
	"fmt"
	"os"

	"0xKowalski1/container-orchestrator/config"
	controlnode "0xKowalski1/container-orchestrator/control-node"

	"github.com/labstack/echo/v4"

	echomiddleware "github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load config
	cfgPath := "config.json" // Take me from env
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		fmt.Printf("Error loading config: %v", err)
		os.Exit(1)
	}

	e := echo.New()

	// Etcd
	etcdClient, err := controlnode.NewEtcdClient()
	if err != nil {
		fmt.Printf("Error creating Etcd Client: %v", err)
		os.Exit(1)
	}

	// Services
	containerService := controlnode.NewContainerService(cfg, etcdClient)
	nodeService := controlnode.NewNodeService(cfg, etcdClient, containerService)

	// Handlers
	containerHandler := controlnode.NewContainerHandler(containerService, nodeService)
	nodeHandler := controlnode.NewNodeHandler(nodeService)

	// Middleware
	e.Use(echomiddleware.Logger())

	// Cors
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	// New Schedular
	controlnode.NewSchedular(etcdClient, containerService, nodeService)

	// Routes

	/// Nodes
	e.GET("/nodes", nodeHandler.GetNodes)
	e.GET("/nodes/:id", nodeHandler.GetNode)
	e.POST("/nodes", nodeHandler.JoinCluster)

	// Containers
	e.GET("/containers", containerHandler.GetContainers)
	e.GET("/containers/:id", containerHandler.GetContainer)
	e.POST("/containers", containerHandler.CreateContainer)
	e.DELETE("/containers/:id", containerHandler.DeleteContainer)
	e.PATCH("/containers/:id", containerHandler.UpdateContainer)
	e.POST("/containers/:id/start", containerHandler.StartContainer)
	e.POST("/containers/:id/stop", containerHandler.StopContainer)
	e.GET("/containers/:id/logs", containerHandler.StreamContainerLogs)
	e.GET("/containers/:id/watch", containerHandler.GetContainerStatus)

	fmt.Printf("Listening on :%d", 8080)
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", 8080)))
}
