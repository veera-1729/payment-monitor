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

	"github.com/go-redis/redis"
	"github.com/yourusername/payment-monitor/internal/contextbuilder"
	"github.com/yourusername/payment-monitor/internal/llm"
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
		Threshold:       cfg.Monitoring.Thresholds.SuccessRateDrop,
		MinTransactions: cfg.Monitoring.Thresholds.MinTransactions,
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
	contextBuilderConfig := &contextbuilder.Config{
		GitHubToken:   cfg.ContextBuilder.GitHub.Token,
		GitHubRepos:   cfg.ContextBuilder.GitHub.Repos,
		LogPath:       cfg.ContextBuilder.Logs.Path,
		ExperimentURL: cfg.ContextBuilder.Experiments.ApiUrl,
		MaxCommitsPerRepo: 10,
		LookbackHours:     24,
		SplitzToken:   cfg.ContextBuilder.Experiments.SplitzToken,
		ExperimentIds: cfg.ContextBuilder.Experiments.ExperimentIds,
	}

	contextBuilder := contextbuilder.NewContextBuilder(contextBuilderConfig, initRedis(cfg))

	contextBuilder.FetchAndStorePreviousData(cfg.ContextBuilder.Experiments.ExperimentIds)

	// Handle graceful shutdown
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

func processAlerts(
	ctx context.Context,
	alertChan <-chan *models.Alert,
	contextBuilder *contextbuilder.ContextBuilder,
	analyzer *llm.Analyzer,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case alert := <-alertChan:
			go func(alert *models.Alert) {
				// Build context
				fmt.Println("Building context for alert", alert)
				context, err := contextBuilder.BuildContext(ctx, alert)
				if err != nil {
					log.Printf("Error building context: %v", err)
					return
				}

				// Analyze with LLM
				result, err := analyzer.Analyze(ctx, context)
				if err != nil {
					log.Printf("Error analyzing with LLM: %v", err)
					return
				}

				// Log the analysis result
				log.Printf("Analysis for alert %s:\nRoot Cause: %s\nConfidence: %.2f\nRecommendations: %v\n",
					alert.ID,
					result.RootCause,
					result.Confidence,
					result.Recommendations,
				)

				// TODO: Implement alert notification (e.g., Slack, email, etc.)
			}(alert)
		default:
			fmt.Println("listening for alerts")
			time.Sleep(1 * time.Second)
		}
	}
} 

func initRedis(cfg *config.Config) (*redis.Client) {
    // Initialize Redis client
    rdb := redis.NewClient(&redis.Options{
        Addr:     cfg.Redis.Host + ":" + fmt.Sprint(cfg.Redis.Port),
        Password: cfg.Redis.Password, // no password set
        DB:       cfg.Redis.DB,  // use default DB
    })
	return rdb
}
