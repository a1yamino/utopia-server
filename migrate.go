package main

import (
	"log"
	"utopia-server/internal/config"
	"utopia-server/internal/database"
)

func main() {
	// Load configuration to get database DSN
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	// Perform database migration
	log.Println("Starting database migration...")
	if err := database.Migrate(cfg.Database.DSN); err != nil {
		log.Fatalf("could not migrate database: %v", err)
	}

	log.Println("Database migration completed successfully.")
}
