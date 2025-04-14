package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yourusername/payment-monitor/pkg/models"
)

type Analyzer struct {
	config *Config
	client *http.Client
}

type Config struct {
	APIKey    string
	Model     string
	Endpoint  string
}

type AnalysisResult struct {
	RootCause    string   `json:"root_cause"`
	Confidence   float64  `json:"confidence"`
	Recommendations []string `json:"recommendations"`
	RelatedChanges []string `json:"related_changes"`
}

func NewAnalyzer(config *Config) *Analyzer {
	return &Analyzer{
		config: config,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (a *Analyzer) Analyze(ctx context.Context, context *models.AnalysisContext) (*AnalysisResult, error) {
	prompt := a.buildPrompt(context)
	
	requestBody := map[string]interface{}{
		"model": a.config.Model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are an expert payment system analyst. Analyze the provided context and identify potential root causes for payment success rate drops.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.7,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.config.Endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.config.APIKey))

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	var result AnalysisResult
	if err := json.Unmarshal([]byte(response.Choices[0].Message.Content), &result); err != nil {
		return nil, fmt.Errorf("error parsing LLM response: %v", err)
	}

	return &result, nil
}

func (a *Analyzer) buildPrompt(context *models.AnalysisContext) string {
	prompt := fmt.Sprintf(`
Payment Success Rate Analysis Request:

Dimension: %s
Value: %s
Current Success Rate: %.2f%%
Previous Success Rate: %.2f%%
Drop Percentage: %.2f%%
Timestamp: %s

Recent Changes:
%s

Recent Logs:
%s

Active Experiments:
%s

Please analyze this information and provide:
1. The most likely root cause of the success rate drop
2. Your confidence level in this analysis (0-1)
3. Recommended actions to address the issue
4. Any related code changes that might be contributing to the problem

Format your response as a JSON object with the following structure:
{
    "root_cause": "string",
    "confidence": float,
    "recommendations": ["string"],
    "related_changes": ["string"]
}
`,
		context.PaymentStats.Dimension,
		context.PaymentStats.Value,
		context.PaymentStats.SuccessRate,
		context.PaymentStats.PreviousRate,
		context.PaymentStats.DropPercentage,
		context.PaymentStats.Timestamp.Format(time.RFC3339),
		a.formatGitHubChanges(context.RecentChanges),
		a.formatLogs(context.LogEntries),
		a.formatExperiments(context.Experiments),
	)

	return prompt
}

func (a *Analyzer) formatGitHubChanges(changes []models.GitHubChange) string {
	var formatted string
	for _, change := range changes {
		formatted += fmt.Sprintf("- %s by %s at %s: %s\n",
			change.CommitID,
			change.Author,
			change.Timestamp.Format(time.RFC3339),
			change.Message,
		)
	}
	return formatted
}

func (a *Analyzer) formatLogs(logs []models.LogEntry) string {
	var formatted string
	for _, log := range logs {
		formatted += fmt.Sprintf("- [%s] %s: %s\n",
			log.Timestamp.Format(time.RFC3339),
			log.Level,
			log.Message,
		)
	}
	return formatted
}

func (a *Analyzer) formatExperiments(experiments []models.Experiment) string {
	var formatted string
	for _, exp := range experiments {
		formatted += fmt.Sprintf("- %s (%s): %s\n",
			exp.Name,
			exp.ID,
			exp.Description,
		)
	}
	return formatted
} 