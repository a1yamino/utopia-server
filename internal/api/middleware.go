package api

import (
	"fmt"
	"net/http"
	"strings"
	"utopia-server/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func (s *Server) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "bearer token required"})
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return s.authService.GetJWTSecret(), nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			username, ok := claims["sub"].(string)
			if !ok {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
				return
			}

			user, role, err := s.authService.GetUserWithRole(username)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
				return
			}

			c.Set("user", user)
			c.Set("role", role)
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Next()
	}
}

func (s *Server) RBACMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get role from context
		roleVal, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "role not found in context"})
			return
		}
		role, ok := roleVal.(*models.Role)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "invalid role type in context"})
			return
		}

		// If allow_all is true, skip all checks
		if allowAll, ok := role.Policies["allow_all"].(bool); ok && allowAll {
			c.Next()
			return
		}

		// 2. Bind the request body to GpuClaimSpec
		var spec models.GpuClaimSpec
		if err := c.ShouldBindJSON(&spec); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
			return
		}

		// 3. Check policies
		if maxGpuCount, ok := role.Policies["max_gpu_count"].(float64); ok {
			if spec.Resources.GpuCount > int(maxGpuCount) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "permission denied: GPU count exceeds quota"})
				return
			}
		}

		// Pass the parsed spec to the handler via context
		c.Set("spec", &spec)

		c.Next()
	}
}
