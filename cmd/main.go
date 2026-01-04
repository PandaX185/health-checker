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

	"github.com/jackc/pgx/v5/pgxpool"
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

	var dbPool *pgxpool.Pool
	var err error
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		timeoutCtx, cancelTimeout := context.WithTimeout(context.Background(), 10*time.Second)
		dbPool, err = database.New(timeoutCtx, os.Getenv("DATABASE_URL"), log.Named("Database"))
		if err == nil {
			break
		}
		if i < maxRetries-1 {
			waitTime := time.Duration(i+1) * time.Second
			log.Warn("Database connection failed, retrying...", zap.Error(err), zap.Duration("wait", waitTime))
			time.Sleep(waitTime)
		}
		cancelTimeout()
	}
	if err != nil {
		log.Fatal("Failed to connect to database after retries", zap.Error(err))
	}

	// if err := migrations.Rollback(dbPool); err != nil {
	// 	log.Fatal("Database rollback failed", zap.Error(err))
	// }

	if err := migrations.Migrate(dbPool); err != nil {
		log.Fatal("Database migration failed", zap.Error(err))
	}

	for i := 0; i < maxRetries; i++ {
		if err := database.RdbInstance.Ping(context.Background()).Err(); err == nil {
			break
		}
		if i < maxRetries-1 {
			waitTime := time.Duration(i+1) * time.Second
			log.Warn("Redis connection failed, retrying...", zap.Error(err), zap.Duration("wait", waitTime))
			time.Sleep(waitTime)
		} else {
			log.Fatal("Failed to connect to Redis after retries", zap.Error(err))
		}
	}

	// Initialize event bus and hub
	eventBus := monitor.NewInMemoryEventBus(log.Named("EventBus"))

	hub := monitor.NewWsHub(log.Named("Websocket Hub"))

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	go hub.Run(ctx)

	// Subscribe hub to status change events
	eventBus.Subscribe("StatusChange", func(ctx context.Context, event monitor.Event) {
		if statusChangeEvent, ok := event.(monitor.StatusChangeEvent); ok {
			if err := hub.BroadcastStatusChange(statusChangeEvent); err != nil {
				log.Error("failed to broadcast status change", zap.Error(err))
			}
		}
	})

	monitorRepo := monitor.NewRepository(dbPool)
	monitorService := monitor.NewService(monitorRepo, log.Named("Monitoring service"))
	monitorHandler := monitor.NewHandler(monitorService, hub, log.Named("MonitorHandler"))

	userRepo := auth.NewRepository(dbPool)
	userService := auth.NewService(userRepo, log.Named("User service"))
	authHandler := auth.NewHandler(userService, log.Named("AuthHandler"))

	srv := app.NewServer(log)

	v1 := srv.Group("/api/v1")
	authHandler.RegisterRoutes(v1.Group("/auth"))

	servicesGroup := v1.Group("/services")
	monitorHandler.RegisterRoutes(servicesGroup)

	scheduler := monitor.NewScheduler(database.RdbInstance, monitorRepo, 1, log.Named("Scheduler"))
	go scheduler.Start(ctx)

	worker := monitor.NewWorker(database.RdbInstance, monitorRepo, log.Named("Worker"), eventBus)
	go worker.Run(ctx)

	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	}
	if err := srv.Run(port); err != nil {
		log.Fatal("Failed to run server", zap.Error(err))
	}
}
