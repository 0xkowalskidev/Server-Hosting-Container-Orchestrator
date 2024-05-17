package controlnode

import (
	"0xKowalski1/container-orchestrator/models"
	"net/http"

	"github.com/labstack/echo/v4"
)

type NodeHandler struct {
	NodeService *NodeService
}

func NewNodeHandler(nodeService *NodeService) *NodeHandler {
	return &NodeHandler{
		NodeService: nodeService,
	}
}

// GetNodes handles GET /nodes
func (handler *NodeHandler) GetNodes(c echo.Context) error {
	nodes, err := handler.NodeService.GetNodes()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": "Something went wrong.",
		})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"nodes": nodes,
	})
}

// GetNode handles GET /nodes/:id
func (handler *NodeHandler) GetNode(c echo.Context) error {
	nodeID := c.Param("id")

	node, err := handler.NodeService.GetNode(nodeID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": "Something went wrong.",
		})
	}

	if node == nil {
		return c.JSON(http.StatusNotFound, echo.Map{
			"error": "Node not found",
		})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"node": node,
	})
}

// JoinCluster handles POST /nodes
func (handler *NodeHandler) JoinCluster(c echo.Context) error {
	var newNode models.CreateNodeRequest

	if err := c.Bind(&newNode); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid node data"})
	}

	existingNode, err := handler.NodeService.GetNode(newNode.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to check if node exists"})
	}

	if existingNode != nil {
		return c.JSON(http.StatusConflict, echo.Map{"error": "Node already exists"})
	}

	err = handler.NodeService.CreateNode(newNode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to add node"})
	}

	return c.JSON(http.StatusOK, echo.Map{"status": "Node added successfully"})
}
