package contextbuilder

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/yourusername/payment-monitor/pkg/models"
)

// Store experiment in Redis as previous data
func (b *ContextBuilder) StorePreviousExperiment(exp *models.StoredExperiment) error {
    // Create Redis key for previous data
    prevKey := "experiment:" + exp.ExperimentID + ":prev"
    
    // Marshal experiment data
    expBytes, err := json.Marshal(exp)
    if err != nil {
        return fmt.Errorf("error marshalling experiment: %v", err)
    }
    
    // Store as previous experiment data
    if err := b.redisClient.Set( prevKey, expBytes, 0).Err(); err != nil {
        return fmt.Errorf("error storing previous experiment in Redis: %v", err)
    }
    
    log.Printf("Stored experiment %s as previous data in Redis", exp.ExperimentID)
    return nil
}

// Retrieve previous experiment data from Redis
func (b *ContextBuilder) getPreviousExperiment(id string) (*models.StoredExperiment, error) {
    prevKey := "experiment:" + id + ":prev"
    prevVal, err := b.redisClient.Get(prevKey).Result()
    if err != nil {
        return nil, fmt.Errorf("error retrieving previous experiment data: %v", err)
    }
    
    var prevExp models.StoredExperiment
    if err := json.Unmarshal([]byte(prevVal), &prevExp); err != nil {
        return nil, fmt.Errorf("error unmarshalling previous experiment data: %v", err)
    }
    
    return &prevExp, nil
}