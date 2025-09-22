package api

import (
	"net/http"

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
