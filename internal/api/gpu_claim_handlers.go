package api

import (
	"net/http"

	"utopia-server/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *Server) handleCreateGpuClaim(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var spec models.GpuClaimSpec
	if err := c.ShouldBindJSON(&spec); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claim := &models.GpuClaim{
		ID:     uuid.NewString(),
		Spec:   spec,
		UserID: userID.(string),
		Status: models.GpuClaimStatus{
			Phase: models.GpuClaimPhasePending,
		},
	}

	if err := s.GpuClaimStore.CreateGpuClaim(claim); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create gpu claim"})
		return
	}

	c.JSON(http.StatusAccepted, claim)
}
