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

	"github.com/gorilla/websocket"
	"github.com/yourusername/payment-monitor/internal/contextbuilder"
	"github.com/yourusername/payment-monitor/internal/llm"
	"github.com/yourusername/payment-monitor/internal/observer"
	wshandler "github.com/yourusername/payment-monitor/internal/websocket"
	"github.com/yourusername/payment-monitor/pkg/config"
	"github.com/yourusername/payment-monitor/pkg/models"
	"github.com/yourusername/payment-monitor/scripts"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

func main() {
	scripts.RunMigrations()

	configPath := flag.String("config", "config/config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Initialize database connection
	db, err := initDB(cfg)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	// Create channels for communication
	alertChan := make(chan *models.Alert, 100)
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize WebSocket hub
	hub := wshandler.NewHub()
	go hub.Run()

	// Initialize components
	observerConfig := &observer.Config{
		Interval:        time.Duration(cfg.Monitoring.Interval) * time.Second,
		Threshold:       cfg.Monitoring.Threshold,
		MinTransactions: cfg.Monitoring.MinTransactions,
		Dimensions:      getEnabledDimensions(cfg),
	}

	observer := observer.NewObserver(db, observerConfig, alertChan, hub)

	contextBuilderConfig := &contextbuilder.Config{
		GitHubToken:       cfg.ContextBuilder.GitHub.Token,
		GitHubRepos:       cfg.ContextBuilder.GitHub.Repos,
		LogPath:           cfg.ContextBuilder.Logs.Path,
		ExperimentURL:     cfg.ContextBuilder.Experiments.Endpoint,
		MaxCommitsPerRepo: 10,
		LookbackHours:     24,
	}

	contextBuilder := contextbuilder.NewContextBuilder(contextBuilderConfig)

	llmConfig := &llm.Config{
		APIKey:     cfg.LLM.APIKey,
		Model:      cfg.LLM.Model,
		Endpoint:   cfg.LLM.Endpoint,
		Deployment: cfg.LLM.Deployment,
		APIVersion: cfg.LLM.APIVersion,
		APIType:    cfg.LLM.APIType,
	}

	analyzer := llm.NewAnalyzer(llmConfig)

	// Start the observer
	go observer.Start(ctx)

	// Start the alert processor
	go processAlerts(ctx, alertChan, contextBuilder, analyzer, hub)

	// Set up HTTP server for WebSocket connections
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
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

	// Start HTTP server
	server := &http.Server{
		Addr: ":8080",
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down server: %v", err)
	}

	cancel()
	log.Println("Shutting down...")
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
	fmt.Println("dsn", dsn)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
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
	hub *wshandler.Hub,
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

				// Send alert to WebSocket clients
				hub.BroadcastAlert(&wshandler.AlertMessage{
					Type:           "alert",
					ID:             alert.ID,
					Dimension:      alert.Dimension,
					Value:          alert.Value,
					CurrentRate:    alert.CurrentRate,
					PreviousRate:   alert.PreviousRate,
					DropPercentage: alert.DropPercentage,
					Timestamp:      alert.Timestamp,
				})

				// TODO: Implement alert notification (e.g., Slack, email, etc.)
			}(alert)
		default:
			fmt.Println("listening for alerts")
			time.Sleep(1 * time.Second)
		}
	}
}
