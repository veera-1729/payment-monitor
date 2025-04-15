package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/payment-monitor/internal/observer"
	"github.com/yourusername/payment-monitor/internal/seeder"
	"github.com/yourusername/payment-monitor/pkg/config"
	"github.com/yourusername/payment-monitor/pkg/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := initDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create alert channel
	alertChannel := make(chan *models.Alert, 100)

	// Initialize observer
	observerConfig := &observer.Config{
		Interval:        time.Duration(cfg.Monitoring.Interval) * time.Second,
		Threshold:       cfg.Monitoring.Threshold,
		MinTransactions: cfg.Monitoring.MinTransactions,
		Dimensions:      getEnabledDimensions(cfg),
	}
	obs := observer.NewObserver(db, observerConfig, alertChannel)

	// Initialize seeder
	seed := seeder.NewSeeder(db)

	// Create HTTP server
	mux := http.NewServeMux()
	seed.RegisterRoutes(mux)

	// Set server address
	serverAddr := ":8080" // Default address

	server := &http.Server{
		Addr:    serverAddr,
		Handler: mux,
	}

	// Start observer in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go obs.Start(ctx)

	// Start HTTP server in background
	go func() {
		log.Printf("Starting server on %s", serverAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Shutdown server
	log.Println("Shutting down server...")
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}
}

func initDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
	cfg.Database.User,
	cfg.Database.Password,
	cfg.Database.Host,
	cfg.Database.Port,
	cfg.Database.DBName,
	cfg.Database.SSLMode,
)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Auto-migrate models
	if err := db.AutoMigrate(&models.Payment{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %v", err)
	}

	return db, nil
}

func getEnabledDimensions(cfg *config.Config) []string {
	var dimensions []string
	for _, dim := range cfg.Monitoring.Dimensions {
		if dim.Enabled {
			dimensions = append(dimensions, dim.Name)
		}
	}
	return dimensions
} 