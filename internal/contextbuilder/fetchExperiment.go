package contextbuilder

import (
	"context"
	"fmt"
	"log"

	"github.com/yourusername/payment-monitor/pkg/models"
    "github.com/yourusername/payment-monitor/pkg/config"
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
        
        experimentPairs = append(experimentPairs, pair)
    }
    
    if len(experimentPairs) == 0 {
        return nil, fmt.Errorf("no valid experiment pairs collected")
    }
    
    return experimentPairs, nil
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
func (b *ContextBuilder) FetchAndStorePreviousData(expIDs [] config.ExperimentID) {
    
    // Fetch each experiment and store in Redis as PREVIOUS data
    for _, id := range expIDs {
        exp, err := b.FetchExperiment(id.ID)
        if err != nil {
            log.Printf("Error fetching experiment %s: %v", id, err)
            continue
        }
        
        // Store directly as previous data
        b.StorePreviousExperiment(exp)
    }
}