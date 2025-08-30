// Package docs for swagger generation
// @title Airline Booking API
// @version 1.0
// @description Smart seat reservation system with hold and purchase functionality
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.example.com/support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.basic BasicAuth

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"airline-booking/internal/api"
	"airline-booking/internal/config"
	"airline-booking/internal/db"
	"airline-booking/internal/es"
	"airline-booking/internal/jobs"
	"airline-booking/internal/repository"
	"airline-booking/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Setup logger
	logger := setupLogger(cfg.Log)
	defer logger.Sync()

	logger.Info("Starting Airline Booking API",
		zap.String("version", "1.0.0"),
		zap.String("env", cfg.App.Env))

	// Initialize database
	database, err := db.NewDatabase(&cfg.Database, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer database.Close()

	// Run migrations
	// if err := database.RunMigrations("migrations"); err != nil {
	//	logger.Fatal("Failed to run migrations", zap.Error(err))
	// }

	// Initialize Elasticsearch client
	esClient, err := es.NewClient(&cfg.Elasticsearch, logger)
	if err != nil {
		logger.Fatal("Failed to connect to Elasticsearch", zap.Error(err))
	}

	// Create Elasticsearch index
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := esClient.CreateIndex(ctx); err != nil {
		logger.Fatal("Failed to create Elasticsearch index", zap.Error(err))
	}

	// Initialize repositories
	seatRepo := repository.NewSeatRepository(database, logger)
	ticketRepo := repository.NewTicketRepository(database, logger)
	flightRepo := repository.NewFlightRepository(database, logger)

	// Initialize services
	bookingService := service.NewBookingService(
		seatRepo,
		ticketRepo,
		flightRepo,
		esClient,
		database,
		cfg,
		logger,
	)

	// Initialize cleanup job
	cleanupJob := jobs.NewCleanupJob(bookingService, logger)
	if err := cleanupJob.Start(); err != nil {
		logger.Fatal("Failed to start cleanup job", zap.Error(err))
	}
	defer cleanupJob.Stop()

	// Initialize API handlers and router
	bookingHandler := api.NewBookingHandler(bookingService, logger)
	router := api.NewRouter(bookingHandler, cfg, logger)
	router.Setup()

	// Setup HTTP server
	server := &http.Server{
		Addr:         cfg.App.Host + ":" + cfg.App.Port,
		Handler:      router.GetEngine(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting HTTP server", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Give outstanding requests a deadline for completion
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

func setupLogger(cfg config.LogConfig) *zap.Logger {
	var zapConfig zap.Config

	if cfg.Format == "json" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Set log level
	switch cfg.Level {
	case "debug":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	logger, err := zapConfig.Build()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	return logger
}
