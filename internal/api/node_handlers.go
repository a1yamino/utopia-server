package api

import (
	"net/http"
	"utopia-server/internal/models"

	"github.com/gin-gonic/gin"
)

func (s *Server) handleNodeRegister(c *gin.Context) {
	var req struct {
		Hostname string `json:"hostname"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.Hostname == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "hostname is required"})
		return
	}

	newNode, err := s.nodeService.CreateNode(req.Hostname)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create node"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"node-id": newNode.ID})
}

func (s *Server) handleGetNodeStatus(c *gin.Context) {
	idStr := c.Param("id")
	// id, err := strconv.ParseUint(idStr, 10, 64)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid node ID"})
	// 	return
	// }

	node, err := s.nodeService.GetNode(idStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}

	if node.Status != models.NodeStatusOnline || node.ControlPort == 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Node is not online"})
		return
	}

	metrics, err := s.agentClient.GetNodeMetrics(node)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to get node metrics from agent"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}
