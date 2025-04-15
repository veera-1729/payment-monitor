package contextbuilder

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/yourusername/payment-monitor/pkg/config"
	"github.com/yourusername/payment-monitor/pkg/models"
)

// Collect current and previous experiment data for all experiments
func (b *ContextBuilder) CollectExperimentPairs (ctx context.Context) ([]models.ExperimentPair, error) {
    
    // Create a list to hold experiment pairs
    experimentPairs := []models.ExperimentPair{}
    
    // Get all experiment IDs as a slice
    ids := b.config.ExperimentIds
    
    // For each experiment ID
    for _, id := range ids {
        pair, err := b.processExperiment(id.ID)
        if err != nil {
            log.Printf("Skipping experiment %s: %v", id, err)
            continue
        }
        
        // Check if there's a difference between current and previous data
        if hasExperimentChanged(pair) {
            log.Printf("Found change in experiment %s, adding to results", id.ID)
            experimentPairs = append(experimentPairs, pair)
        } else {
            log.Printf("No change detected in experiment %s, skipping", id.ID)
        }
    }
    
    if len(experimentPairs) == 0 {
        return nil, fmt.Errorf("no valid experiment pairs collected")
    }
    
    return experimentPairs, nil
}

// hasExperimentChanged determines if there's a difference between current and previous experiment data
func hasExperimentChanged(pair models.ExperimentPair) bool {
    // Ensure both current and previous exist
    if pair.Current == nil || pair.Previous == nil {
        // If either is nil, consider it a change
        return true
    }
    
    // Convert audience to JSON strings for comparison
    currentJSON, err1 := json.Marshal(pair.Current.Audience)
    previousJSON, err2 := json.Marshal(pair.Previous.Audience)
    
    // If marshaling fails, consider it a change
    if err1 != nil || err2 != nil {
        return true
    }
    
    // Compare the JSON strings
    return string(currentJSON) != string(previousJSON)
}

// Process a single experiment: get previous data, fetch current data, update storage
func (b *ContextBuilder) processExperiment(id string,) (models.ExperimentPair, error) {
    // Get previous data from Redis
    prevExp, err := b.getPreviousExperiment(id)
    if err != nil {
        return models.ExperimentPair{}, fmt.Errorf("no previous data found: %v", err)
    }
    
    // Fetch current data
    currentExp, err := b.FetchExperiment(id)
    if err != nil {
        return models.ExperimentPair{}, fmt.Errorf("error fetching current data: %v", err)
    }
    
    // Create pair
    pair := models.ExperimentPair{
        ExperimentID: id,
        Current:      currentExp,
        Previous:     prevExp,
    }
    
    return pair, nil
}

// Fetch data for all experiments and store as previous data
func (b *ContextBuilder) FetchAndStorePreviousData(expIDs []config.ExperimentID) {    
    // Fetch each experiment and store in Redis as PREVIOUS data
    for _, id := range expIDs {
        exp, err := b.FetchExperiment(id.ID)
        if err != nil {
            log.Printf("Error fetching experiment %s: %v", id.ID, err)
            continue
        }
        
        // Store directly as previous data
        err = b.StorePreviousExperiment(exp)
        if err != nil {
            log.Printf("Error storing experiment %s: %v", id.ID, err)
        }
    }
    
    log.Printf("Finished processing all %d experiments", len(expIDs))
}