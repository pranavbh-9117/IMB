// @title Institute Management Backend API
// @version 1.0
// @description API documentation for the Institute Management Platform.
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

package main

import (
	"fmt"
	"os"

	_ "github.com/pranavbh-9117/IMB/docs"
	"github.com/pranavbh-9117/IMB/internal/app"
)

func main() {
	// Initialize Application (Config, Logger, DB, Migrations, Seed, Routes)
	application, err := app.NewApp()
	if err != nil {
		fmt.Printf("Failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	// Start Server
	if err := application.Start(); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}
}
