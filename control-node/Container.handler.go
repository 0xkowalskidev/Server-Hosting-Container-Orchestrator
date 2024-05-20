package controlnode

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"0xKowalski1/container-orchestrator/models"
	"github.com/labstack/echo/v4"
)

type ContainerHandler struct {
	ContainerService *ContainerService
	NodeService      *NodeService
}

func NewContainerHandler(containerService *ContainerService, nodeService *NodeService) *ContainerHandler {
	return &ContainerHandler{
		ContainerService: containerService,
		NodeService:      nodeService,
	}
}

func (handler *ContainerHandler) GetContainers(c echo.Context) error {
	containers, err := handler.ContainerService.GetContainers()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"containers": containers,
	})
}

// CreateContainer handles POST /containers
func (handler *ContainerHandler) CreateContainer(c echo.Context) error {
	var req models.CreateContainerRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request"})
	}

	createdContainer, err := handler.ContainerService.CreateContainer(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, echo.Map{
		"container": createdContainer,
	})
}

// UpdateContainer handles PATCH /containers/:id
func (handler *ContainerHandler) UpdateContainer(c echo.Context) error {
	containerID := c.Param("id")
	var req models.UpdateContainerRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request"})
	}

	err := handler.ContainerService.UpdateContainer(containerID, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"success": true,
	})
}

// GetContainer handles GET /containers/:id
func (handler *ContainerHandler) GetContainer(c echo.Context) error {
	containerID := c.Param("id")

	container, err := handler.ContainerService.GetContainer(containerID)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Container not found"})
	}

	return c.JSON(http.StatusOK, container)
}

// DeleteContainer handles DELETE /containers/:id
func (handler *ContainerHandler) DeleteContainer(c echo.Context) error {
	containerID := c.Param("id")
	err := handler.ContainerService.DeleteContainer(containerID, handler.NodeService)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"success": "true"})
}

// StartContainer handles POST /containers/:id/start
func (handler *ContainerHandler) StartContainer(c echo.Context) error {
	containerID := c.Param("id")
	desiredStatus := "running"

	err := handler.ContainerService.UpdateContainer(containerID, models.UpdateContainerRequest{
		DesiredStatus: &desiredStatus,
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Container starting"})
}

// StopContainer handles POST /containers/:id/stop
func (handler *ContainerHandler) StopContainer(c echo.Context) error {
	containerID := c.Param("id")
	desiredStatus := "stopped"

	err := handler.ContainerService.UpdateContainer(containerID, models.UpdateContainerRequest{
		DesiredStatus: &desiredStatus,
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Container stopping"})
}

func (handler *ContainerHandler) StreamContainerLogs(c echo.Context) error {
	containerID := c.Param("id")

	container, err := handler.ContainerService.GetContainer(containerID)
	if err != nil {
		log.Printf("Error getting container: %v", err)
	}

	node, err := handler.NodeService.GetNode(container.NodeID)
	if err != nil {
		log.Printf("Error getting node: %v", err)
	}

	workerAddress := fmt.Sprintf("http://%s:8081", node.NodeIp)
	targetURL, err := url.Parse(workerAddress)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to parse worker address: %v", err))
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.URL.Path = "/containers/" + containerID + "/logs"
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
	}

	proxy.ServeHTTP(c.Response(), c.Request())
	return nil
}

func (handler *ContainerHandler) GetContainerStatus(c echo.Context) error {
	containerID := c.Param("id")

	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("Transfer-Encoding", "chunked")
	c.Response().WriteHeader(http.StatusOK)

	// Flusher to ensure data is sent to the client immediately
	flusher, ok := c.Response().Writer.(http.Flusher)
	if !ok {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "streaming unsupported"})
	}

	statusChan, err := handler.ContainerService.SubscribeToStatus(containerID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to subscribe to container status"})
	}
	defer handler.ContainerService.UnsubscribeFromStatus(containerID, statusChan)

	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	for {
		select {
		case status, ok := <-statusChan:
			if !ok {
				return nil
			}
			fmt.Fprintf(c.Response(), "data: %s\n\n", status)
			flusher.Flush()
		case <-heartbeatTicker.C:
			fmt.Fprintf(c.Response(), ":heartbeat\n\n")
			flusher.Flush()
		case <-c.Request().Context().Done():
			return nil
		}
	}
}
