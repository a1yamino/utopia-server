package api

import (
	"net/http"
	"time"

	"utopia-server/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *Server) handleCreateGpuClaim(c *gin.Context) {
	// Retrieve user from context, set by AuthMiddleware
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	user, ok := userVal.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user type in context"})
		return
	}

	// Retrieve spec from context, set by RBACMiddleware
	specVal, exists := c.Get("spec")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "spec not found in context"})
		return
	}
	spec, ok := specVal.(*models.GpuClaimSpec)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid spec type in context"})
		return
	}

	// Create and populate the claim object
	claim := &models.GpuClaim{
		ID:        uuid.NewString(),
		UserID:    user.Username,
		CreatedAt: time.Now(),
		Spec:      *spec,
		Status: models.GpuClaimStatus{
			Phase: models.GpuClaimPhasePending,
		},
	}

	if err := s.GpuClaimStore.CreateGpuClaim(claim); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create gpu claim: " + err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, claim)
}
