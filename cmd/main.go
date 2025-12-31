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
	"os/signal"
	"syscall"
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

	timeoutCtx, cancelTimeout := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelTimeout()

	dbPool, err := database.New(timeoutCtx, os.Getenv("DATABASE_URL"), log.Named("Database"))
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

	monitorRepo := monitor.NewRepository(dbPool)
	monitorService := monitor.NewService(monitorRepo, log.Named("Monitoring service"))
	monitorHandler := monitor.NewHandler(monitorService)

	userRepo := auth.NewRepository(dbPool)
	userService := auth.NewService(userRepo, log.Named("User service"))
	authHandler := auth.NewHandler(userService)

	srv := app.NewServer(log)

	v1 := srv.Group("/api/v1")
	authHandler.RegisterRoutes(v1.Group("/auth"))
	monitorHandler.RegisterRoutes(v1.Group("/services"))

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	scheduler := monitor.NewScheduler(database.RdbInstance, monitorRepo, 1, log.Named("Scheduler"))
	go scheduler.Start(ctx)

	worker := monitor.NewWorker(database.RdbInstance, monitorRepo, log.Named("Worker"))
	go worker.Run(ctx)

	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	}
	if err := srv.Run(port); err != nil {
		log.Fatal("Failed to run server", zap.Error(err))
	}
}
