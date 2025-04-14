package models

import "time"

// PaymentStats represents the statistics for a specific dimension
type PaymentStats struct {
	Dimension     string
	Value         string
	Total         int
	Successful    int
	SuccessRate   float64
	PreviousRate  float64
	DropPercentage float64
	Timestamp     time.Time
}

// Alert represents an alert generated when success rate drops
type Alert struct {
	ID            string
	Dimension     string
	Value         string
	CurrentRate   float64
	PreviousRate  float64
	DropPercentage float64
	Timestamp     time.Time
	Context       *AnalysisContext
}

// AnalysisContext contains all the context data for LLM analysis
type AnalysisContext struct {
	PaymentStats  *PaymentStats
	RecentChanges []GitHubChange
	LogEntries    []LogEntry
	Experiments   []Experiment
}

// GitHubChange represents a code change from GitHub
type GitHubChange struct {
	Repo         string
	CommitID     string
	Author       string
	Message      string
	Timestamp    time.Time
	FilesChanged []string
}

// LogEntry represents a log entry from payment system
type LogEntry struct {
	Timestamp time.Time
	Level     string
	Message   string
	Context   map[string]interface{}
}

// Experiment represents an active A/B experiment
type Experiment struct {
	ID          string
	Name        string
	StartTime   time.Time
	EndTime     time.Time
	Description string
	Changes     map[string]interface{}
}

// Payment represents a payment transaction
type Payment struct {
	ID         string    `gorm:"primaryKey"`
	Gateway    string    `gorm:"index"`
	Method     string    `gorm:"index"`
	MerchantID string    `gorm:"index"`
	Status     string    `gorm:"index"`
	CreatedAt  time.Time `gorm:"index"`
} 