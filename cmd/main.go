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
	"context"
	"health-checker/internal/app"
	"health-checker/internal/app/auth"
	"health-checker/internal/database"
	"health-checker/internal/logger"
	"health-checker/internal/migrations"
	"health-checker/internal/monitor"
	"log"
	"os"
	"time"

	_ "health-checker/docs"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	log := logger.New(os.Getenv("ENV"))
	defer log.Sync()

	log.Info("Starting Health Checker Application")

	dbPool, err := database.New(context.Background(), os.Getenv("DATABASE_URL"), log)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer database.Close()

	// if err := migrations.Rollback(dbPool); err != nil {
	// 	log.Fatal("Database rollback failed", zap.Error(err))
	// }

	if err := migrations.Migrate(dbPool); err != nil {
		log.Fatal("Database migration failed", zap.Error(err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := database.Rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("Failed to connect to Redis", zap.Error(err))
	}

	monitorRepo := monitor.NewRepository(dbPool)
	monitorService := monitor.NewService(monitorRepo)
	monitorHandler := monitor.NewHandler(monitorService)

	userRepo := auth.NewRepository(dbPool)
	userService := auth.NewService(userRepo)
	authHandler := auth.NewHandler(userService)

	srv := app.NewServer(log)

	v1 := srv.Group("/api/v1")
	authHandler.RegisterRoutes(v1.Group("/auth"))
	monitorHandler.RegisterRoutes(v1.Group("/services"))

	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	}
	if err := srv.Run(port); err != nil {
		log.Fatal("Failed to run server", zap.Error(err))
	}
}
