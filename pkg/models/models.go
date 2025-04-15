package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// PaymentStats represents the statistics for a specific dimension
type PaymentStats struct {
	Dimension      string
	Value          string
	Total          int
	Successful     int
	SuccessRate    float64
	PreviousRate   float64
	DropPercentage float64
	Timestamp      time.Time
}

// Alert represents an alert generated when success rate drops
type Alert struct {
	ID             string
	Dimension      string
	Value          string
	CurrentRate    float64
	PreviousRate   float64
	DropPercentage float64
	Timestamp      time.Time
	Context        *AnalysisContext
	Gateway        string
	Method         string
	MerchantID     string
}

// AnalysisContext contains all the context data for LLM analysis
type AnalysisContext struct {
	PaymentStats  *PaymentStats
	RecentChanges []GitHubChange
	LogEntries    []LogEntry
	Experiments   []ExperimentPair
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
// Experiment response structure
type ExperimentResponse struct {
	Experiment struct {
		ID       string `json:"id"`
		Audience string `json:"audience"`
		Status   string `json:"status"`
		// Other fields omitted for brevity
	} `json:"experiment"`
}

// StoredExperiment contains only the data we want to store
type StoredExperiment struct {
	ExperimentID string      `json:"experiment_id"`
	Audience     interface{} `json:"audience"` // Changed from string to interface{} to store parsed JSON
	FetchedAt    string      `json:"fetched_at"`
}

// ExperimentPair represents a pair of current and previous experiment data
type ExperimentPair struct {
	ExperimentID string            `json:"experiment_id"`
	Current      *StoredExperiment `json:"current"`
	Previous     *StoredExperiment `json:"previous"`
}

// Payment represents a payment transaction
type Payment struct {
	ID                string `gorm:"primaryKey"`
	MerchantID        string `gorm:"index"`
	PaymentID         string `gorm:"index"`
	Amount            int64
	Currency          string
	Status            string
	Method            string
	Description       string
	Email             string
	Contact           string
	Notes             string
	AutoCaptured      int16
	CallbackURL       string
	Verified          int16
	Disputed          int16
	OnHold            int16
	CustomerID        string
	GlobalCustomerID  string
	TokenID           string
	LateAuthorized    int16
	GatewayCaptured   int16
	AmountRefunded    int64
	AmountTransferred int64
	RefundStatus      string
	AuthorizedAt      int
	RefundedAt        int
	RefundAt          int
	CapturedAt        int
	AuthenticatedAt   int
	SettledBy         string
	InstrumentID      string
	TerminalID        string
	Gateway           string
	FeeData           JSONB
	Error             JSONB
	AcquirerData      JSONB
	JournalID         string
	Wallet            string
	BaseAmount        int64
}

// TableName specifies the table name for Payment model
func (Payment) TableName() string {
	return "payments"
}

// JSONB is a custom type for JSONB fields
type JSONB map[string]interface{}

// Scan implements the sql.Scanner interface for JSONB
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = JSONB{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}
	return json.Unmarshal(bytes, j)
}

// Value implements the driver.Valuer interface for JSONB
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// GormDataType implements the GORM interface for JSONB
func (JSONB) GormDataType() string {
	return "jsonb"
}
