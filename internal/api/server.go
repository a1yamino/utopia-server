package api

import (
	"utopia-server/internal/auth"
	"utopia-server/internal/controller"
	"utopia-server/internal/node"

	"github.com/gin-gonic/gin"
)

// Config holds the configuration for the API server.
type Config struct {
	ListenAddr string
}

// Server is the API server.
type Server struct {
	router        *gin.Engine
	config        *Config
	authService   *auth.Service
	nodeService   *node.Service
	GpuClaimStore controller.GpuClaimStore
}

// NewServer creates a new API server.
func NewServer(config *Config, authService *auth.Service, nodeService *node.Service, gpuClaimStore controller.GpuClaimStore) *Server {
	router := gin.Default() // gin.Default() includes Logger and Recovery middleware.

	server := &Server{
		router:        router,
		config:        config,
		authService:   authService,
		nodeService:   nodeService,
		GpuClaimStore: gpuClaimStore,
	}

	server.setupRoutes()

	return server
}

func (s *Server) setupRoutes() {
	api := s.router.Group("/api")

	// Auth routes
	auth := api.Group("/auth")
	auth.POST("/register", s.handleRegister)
	auth.POST("/login", s.handleLogin)

	// GPU Claim routes
	gpuClaims := api.Group("/gpu-claims")
	gpuClaims.Use(s.AuthMiddleware()) // Protect this group
	gpuClaims.POST("", s.handleCreateGpuClaim)

	// Node routes
	nodes := api.Group("/nodes")
	nodes.POST("/register", s.handleNodeRegister) // No auth middleware
	nodes.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// Admin routes
	admin := api.Group("/admin")
	admin.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})
}

// Run starts the API server.
func (s *Server) Run() error {
	return s.router.Run(s.config.ListenAddr)
}
