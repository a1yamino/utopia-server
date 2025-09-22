package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"utopia-server/internal/api"
	"utopia-server/internal/auth"
	"utopia-server/internal/client"
	"utopia-server/internal/config"
	"utopia-server/internal/controller"
	"utopia-server/internal/database"
	"utopia-server/internal/node"
	"utopia-server/internal/scheduler"
	"utopia-server/internal/tunnel"

	"github.com/go-sql-driver/mysql"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	// Ensure database exists before migration
	dsnCfg, err := mysql.ParseDSN(cfg.Database.DSN)
	if err != nil {
		log.Fatalf("invalid DSN: %v", err)
	}
	dbName := dsnCfg.DBName
	dsnCfg.DBName = "" // Connect without a specific database

	tempDB, err := sql.Open("mysql", dsnCfg.FormatDSN())
	if err != nil {
		log.Fatalf("could not connect to mysql server to create database: %v", err)
	}

	_, err = tempDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", dbName))
	if err != nil {
		log.Fatalf("could not create database %s: %v", dbName, err)
	}
	if err := tempDB.Close(); err != nil {
		log.Printf("warning: could not close temporary database connection: %v", err)
	}
	log.Printf("Database '%s' checked/created successfully.", dbName)

	// Perform database migration
	if err := database.Migrate(cfg.Database.DSN); err != nil {
		log.Fatalf("could not migrate database: %v", err)
	}

	// Establish database connection
	db, err := database.NewDB(cfg.Database.DSN)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatalf("could not close database: %v", err)
		}
	}(db)

	// Setup tunnel service for frps
	tunnelService := tunnel.NewService(cfg.FRP)
	if err := tunnelService.Start(); err != nil {
		log.Fatalf("could not start tunnel service: %v", err)
	}
	defer tunnelService.Stop()

	authStore := auth.NewMySQLStore(db)
	authService := auth.NewService(authStore, cfg)

	nodeStore := node.NewMySQLStore(db)
	nodeService := node.NewService(nodeStore)

	gpuClaimStore := controller.NewMySQLStore(db)

	// Create the scheduler
	sched := scheduler.NewScheduler(nodeStore)

	// Create and run the controller in a separate goroutine
	agentClient := client.NewAgentClient()
	ctrl := controller.NewController(gpuClaimStore, sched, nodeStore, agentClient)
	stopCh := make(chan struct{})
	defer close(stopCh)
	go ctrl.Run(stopCh)

	// Setup and run discovery service
	discoveryService := node.NewDiscoveryService(
		fmt.Sprintf("http://localhost:%d", cfg.FRP.DashboardPort),
		cfg.FRP.DashboardUser,
		cfg.FRP.DashboardPwd,
		nodeStore,
	)
	go discoveryService.Run(stopCh)

	// Setup and run health check service
	healthCheckService := node.NewHealthCheckService(nodeStore)
	go healthCheckService.Run(stopCh)

	server := api.NewServer(cfg.Server, authService, nodeService, gpuClaimStore)

	go func() {
		listenAddr := fmt.Sprintf("%s:%s", cfg.Server.Addr, cfg.Server.Port)
		log.Printf("Starting server on %s", listenAddr)
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
