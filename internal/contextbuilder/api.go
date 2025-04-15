package contextbuilder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"github.com/yourusername/payment-monitor/pkg/models"
)

// Fetch a single experiment from the API
func (b *ContextBuilder) FetchExperiment(id string) (*models.StoredExperiment, error) {
	// Create request payload
	payload := map[string]interface{}{
		"id":      id,
		"expands": []string{"workflow", "state_change_log"},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshalling payload: %v", err)
	}

	// Create request
	req, err := http.NewRequest("POST", b.config.ExperimentURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", b.config.SplitzToken)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	// Parse response
	var expResp models.ExperimentResponse
	if err := json.Unmarshal(body, &expResp); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %v", err)
	}
	
	// Check if experiment is activated
    if expResp.Experiment.Status != "activated" {
        return nil, fmt.Errorf("experiment %s is not activated (status: %s)", id, expResp.Experiment.Status)
    }

	// Parse audience string into proper JSON
    var audienceJSON interface{}
    if expResp.Experiment.Audience != "" {
        if err := json.Unmarshal([]byte(expResp.Experiment.Audience), &audienceJSON); err != nil {
            log.Printf("Warning: Failed to parse audience JSON for experiment %s: %v", id, err)
            // Keep the original audience string if parsing fails
            audienceJSON = expResp.Experiment.Audience
        }
    }
	

	// Create stored experiment
	storedExp := &models.StoredExperiment{
		ExperimentID: expResp.Experiment.ID,
		Audience:     audienceJSON,
		FetchedAt:    time.Now().Format(time.RFC3339),
	}

	return storedExp, nil
}
