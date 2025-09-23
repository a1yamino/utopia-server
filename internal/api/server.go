package api

import (
	"fmt"
	"net/http"
	"utopia-server/internal/auth"
	"utopia-server/internal/client"
	"utopia-server/internal/config"
	"utopia-server/internal/controller"
	"utopia-server/internal/node"

	"github.com/gin-gonic/gin"
)

// Server is the API server.
type Server struct {
	Router        *gin.Engine
	config        config.ServerConfig
	authService   *auth.Service
	nodeService   *node.Service
	GpuClaimStore controller.GpuClaimStore
	agentClient   *client.AgentClient
}

// NewServer creates a new API server.
func NewServer(config config.ServerConfig, authService *auth.Service, nodeService *node.Service, gpuClaimStore controller.GpuClaimStore, agentClient *client.AgentClient) *Server {
	router := gin.Default() // gin.Default() includes Logger and Recovery middleware.

	server := &Server{
		Router:        router,
		config:        config,
		authService:   authService,
		nodeService:   nodeService,
		GpuClaimStore: gpuClaimStore,
		agentClient:   agentClient,
	}

	router.Static("/ui", "./web/ui")
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/ui")
	})

	server.setupRoutes()

	return server
}

func (s *Server) setupRoutes() {
	api := s.Router.Group("/api")

	// Auth routes
	auth := api.Group("/auth")
	auth.POST("/register", s.handleRegister)
	auth.POST("/login", s.handleLogin)

	// GPU Claim routes
	gpuClaims := api.Group("/gpu-claims")
	gpuClaims.Use(s.AuthMiddleware()) // Protect this group
	gpuClaims.POST("", s.RBACMiddleware(), s.handleCreateGpuClaim)

	// Node routes
	nodes := api.Group("/nodes")
	nodes.POST("/register", s.handleNodeRegister) // No auth middleware
	nodes.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})
	nodes.GET("/:id/status", s.AuthMiddleware(), s.handleGetNodeStatus)

	// Admin routes
	admin := api.Group("/admin")
	admin.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})
}

// Run starts the API server.
func (s *Server) Run() error {
	return s.Router.Run(fmt.Sprintf("%s:%s", s.config.Addr, s.config.Port))
}
