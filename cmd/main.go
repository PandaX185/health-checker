//	@title			Health Checker API
//	@version		1.0
//	@description	A distributed health monitoring system API.

//	@host		localhost:8080
//	@BasePath	/api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"health-checker/internal/app"
	"health-checker/internal/config"
	"health-checker/internal/services"
	"log"
	"os"

	_ "health-checker/docs"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	logger := config.NewLogger(os.Getenv("ENV"))
	defer logger.Sync()

	config.NewDatabase(os.Getenv("DATABASE_URL"))
	if err := config.Migrate(); err != nil {
		logger.Fatal("Database migration failed", zap.Error(err))
	}

	srv := app.NewServer(logger)
	apiGroup := srv.Group("/api/v1")

	services.RegisterServicesRoutes(apiGroup.Group("/services"), services.NewServicesController())

	if err := srv.Run(os.Getenv("PORT")); err != nil {
		logger.Fatal("Failed to run server", zap.Error(err))
	}
}
