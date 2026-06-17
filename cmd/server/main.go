package main

import (
	"log"

	"github.com/pranavbh-9117/IMB/internal/migration"
	"github.com/pranavbh-9117/IMB/internal/seed"
	"github.com/pranavbh-9117/IMB/pkg/config"
	"github.com/pranavbh-9117/IMB/pkg/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	if err := database.HealthCheck(db); err != nil {
		log.Fatalf("Database health check failed: %v", err)
	}
	log.Println("Database connected successfully")

	if err := migration.Run(db); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Database migration completed")

	if err := seed.Run(db, cfg.Seed); err != nil {
		log.Fatalf("Seed failed: %v", err)
	}
	log.Println("Database seed completed")
}