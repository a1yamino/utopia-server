package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"utopia-server/internal/api"
	"utopia-server/internal/auth"
	"utopia-server/internal/controller"
	"utopia-server/internal/node"
	"utopia-server/internal/scheduler"
	"utopia-server/internal/tunnel"
)

func main() {
	// Setup tunnel service for frps
	tunnelConfig := tunnel.Config{
		BindPort:      7000,
		DashboardPort: 7500,
		AuthToken:     "utopia-auth-token",
	}
	tunnelService := tunnel.NewService(tunnelConfig)
	if err := tunnelService.Start(); err != nil {
		log.Fatalf("could not start tunnel service: %v", err)
	}
	defer tunnelService.Stop()

	authStore := auth.NewMemStore()
	authService := auth.NewService(authStore)

	nodeStore := node.NewMemStore()
	nodeService := node.NewService(nodeStore)

	gpuClaimStore := controller.NewMemStore()

	// Create the scheduler
	sched := scheduler.NewScheduler(nodeStore)

	// Create and run the controller in a separate goroutine
	ctrl := controller.NewController(gpuClaimStore, sched)
	stopCh := make(chan struct{})
	defer close(stopCh)
	go ctrl.Run(stopCh)

	config := &api.Config{
		ListenAddr: ":8080",
	}

	server := api.NewServer(config, authService, nodeService, gpuClaimStore)

	go func() {
		log.Printf("Starting server on %s", config.ListenAddr)
		if err := server.Run(); err != nil {
			log.Fatalf("could not start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
}
