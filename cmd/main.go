package main

import (
	"health-checker/internal/app"
	"health-checker/internal/config"
	"log"
	"os"

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
	srv := app.NewServer(logger)
	if err := srv.Run(os.Getenv("PORT")); err != nil {
		logger.Fatal("Failed to run server", zap.Error(err))
	}
}
