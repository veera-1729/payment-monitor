package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
	"github.com/yourusername/payment-monitor/internal/contextbuilder"
	"github.com/yourusername/payment-monitor/internal/llm"
	"github.com/yourusername/payment-monitor/internal/observer"
	"github.com/yourusername/payment-monitor/internal/seeder"
	wshandler "github.com/yourusername/payment-monitor/internal/websocket"
	"github.com/yourusername/payment-monitor/pkg/config"
	"github.com/yourusername/payment-monitor/pkg/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

func main() {
	// Load configuration
	configPath := flag.String("config", "config/config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
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

	// Initialize WebSocket hub
	hub := wshandler.NewHub()
	go hub.Run()

	// Initialize components
	observerConfig := &observer.Config{
		Interval:        time.Duration(cfg.Monitoring.Interval) * time.Second,
		Threshold:       cfg.Monitoring.Thresholds.SuccessRateDrop,
		MinTransactions: cfg.Monitoring.Thresholds.MinTransactions,
		Dimensions:      getEnabledDimensions(cfg),
	}

	obs := observer.NewObserver(db, observerConfig, alertChannel, hub)

	// Initialize seeder
	seed := seeder.NewSeeder(db)

	// Create HTTP server mux
	mux := http.NewServeMux()
	seed.RegisterRoutes(mux)

	// Add WebSocket handler
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Error upgrading connection: %v", err)
			return
		}

		client := &wshandler.Client{
			Conn: conn,
			Send: make(chan []byte, 256),
		}

		hub.Register <- client
		go client.WritePump()
	})

	// Set server address
	serverAddr := ":8080"

	server := &http.Server{
		Addr:    serverAddr,
		Handler: mux,
	}

	// Initialize context builder
	contextBuilderConfig := &contextbuilder.Config{
		GitHubToken:       cfg.ContextBuilder.GitHub.Token,
		GitHubRepos:       cfg.ContextBuilder.GitHub.Repos,
		LogPath:           cfg.ContextBuilder.Logs.Path,
		ExperimentURL:     cfg.ContextBuilder.Experiments.ApiUrl,
		MaxCommitsPerRepo: 10,
		LookbackHours:     24,
		SplitzToken:       cfg.ContextBuilder.Experiments.SplitzToken,
		ExperimentIds:     cfg.ContextBuilder.Experiments.ExperimentIds,
	}

	contextBuilder := contextbuilder.NewContextBuilder(contextBuilderConfig, initRedis(cfg))
	contextBuilder.FetchAndStorePreviousData(cfg.ContextBuilder.Experiments.ExperimentIds)

	// Initialize LLM analyzer
	llmConfig := &llm.Config{
		APIKey:     cfg.LLM.APIKey,
		Model:      cfg.LLM.Model,
		Endpoint:   cfg.LLM.Endpoint,
		Deployment: cfg.LLM.Deployment,
		APIVersion: cfg.LLM.APIVersion,
		APIType:    cfg.LLM.APIType,
	}

	analyzer := llm.NewAnalyzer(llmConfig)

	// Start components
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go obs.Start(ctx)
	go processAlerts(ctx, alertChannel, contextBuilder, analyzer, hub)

	// Start HTTP server in background
	go func() {
		log.Printf("Starting server on %s", serverAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Shutdown server
	log.Println("Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
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

func initRedis(cfg *config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
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

func processAlerts(ctx context.Context, alertChan chan *models.Alert, contextBuilder *contextbuilder.ContextBuilder, analyzer *llm.Analyzer, hub *wshandler.Hub) {
	for {
		select {
		case <-ctx.Done():
			return
		case alert := <-alertChan:
			if alert == nil {
				continue
			}

			// Build context for the alert
			alertContext, err := contextBuilder.BuildContext(context.TODO(), alert)
			if err != nil {
				log.Printf("Error building context: %v", err)
				continue
			}

			// Analyze the alert with context
			analysis, err := analyzer.Analyze(context.TODO(), alertContext)
			if err != nil {
				log.Printf("Error analyzing alert: %v", err)
				continue
			}

			// Print detailed analysis results
			log.Printf("===== ANALYSIS RESULTS =====")
			log.Printf("Alert for: %s - %s", alert.Dimension, alert.Value)
			log.Printf("Root Cause: %s", analysis.RootCause)
			log.Printf("Confidence: %.2f", analysis.Confidence)
			log.Printf("Recommendations:")
			for i, rec := range analysis.Recommendations {
				log.Printf("  %d. %s", i+1, rec)
			}
			if len(analysis.RelatedChanges) > 0 {
				log.Printf("Related Changes:")
				for i, change := range analysis.RelatedChanges {
					log.Printf("  %d. %s", i+1, change)
				}
			}
			log.Printf("============================")

			// Update alert with analysis
			alert.RootCause = analysis.RootCause
			alert.Confidence = analysis.Confidence
			alert.Recommendations = analysis.Recommendations

			// Broadcast alert to WebSocket clients
			if hub != nil {
				// Create a properly formatted alert message
				alertMsg := &wshandler.AlertMessage{
					Type:            "alert",
					ID:              alert.ID,
					Dimension:       alert.Dimension,
					Value:           alert.Value,
					CurrentRate:     alert.CurrentRate,
					PreviousRate:    alert.PreviousRate,
					DropPercentage:  alert.DropPercentage,
					Timestamp:       alert.Timestamp,
					RootCause:       analysis.RootCause,
					Confidence:      analysis.Confidence,
					Recommendations: analysis.Recommendations,
				}
				hub.BroadcastAlert(alertMsg)
			}
		}
	}
}
