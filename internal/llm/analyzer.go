package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/yourusername/payment-monitor/pkg/models"
)

type Analyzer struct {
	config *Config
	client *openai.Client
}

type Config struct {
	APIKey     string
	Model      string
	Endpoint   string
	Deployment string
	APIVersion string
	APIType    string
}

type AnalysisResult struct {
	RootCause       string   `json:"root_cause"`
	Confidence      float64  `json:"confidence"`
	Recommendations []string `json:"recommendations"`
	RelatedChanges  []string `json:"related_changes"`
}

func NewAnalyzer(config *Config) *Analyzer {
	var clientConfig openai.ClientConfig
	if config.APIType == "azure" {
		// Ensure the endpoint has the proper format for Azure OpenAI
		endpoint := config.Endpoint
		if !strings.HasPrefix(endpoint, "https://") {
			endpoint = fmt.Sprintf("https://%s.openai.azure.com", endpoint)
		}
		clientConfig = openai.DefaultAzureConfig(config.APIKey, endpoint)
		clientConfig.APIVersion = config.APIVersion
		clientConfig.AzureModelMapperFunc = func(model string) string {
			return config.Deployment
		}
	} else {
		clientConfig = openai.DefaultConfig(config.APIKey)
		if config.Endpoint != "" {
			clientConfig.BaseURL = config.Endpoint
		}
	}

	return &Analyzer{
		config: config,
		client: openai.NewClientWithConfig(clientConfig),
	}
}

func (a *Analyzer) Analyze(ctx context.Context, context *models.AnalysisContext) (*AnalysisResult, error) {
	prompt := a.buildPrompt(context)

	// For Azure, use the deployment name as the model
	model := a.config.Model
	if a.config.APIType == "azure" {
		model = a.config.Deployment
	}

	resp, err := a.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are an expert code reviewer. Analyze the provided context and identify potential root causes for payment success rate drops.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.7,
			MaxTokens:   1000,
		},
	)

	fmt.Printf("OpenAI API Response: %+v\n", resp)
	if err != nil {
		// Log detailed error information
		fmt.Printf("OpenAI API Error: %v\nModel: %s\n", err, a.config.Model)
		return &AnalysisResult{
			RootCause:  fmt.Sprintf("Error from OpenAI service: %v", err),
			Confidence: 0.5,
			Recommendations: []string{
				"Check API key and permissions",
				"Verify service endpoint configuration",
				"Monitor system logs for detailed errors",
			},
			RelatedChanges: []string{},
		}, nil
	}

	if len(resp.Choices) == 0 {
		return &AnalysisResult{
			RootCause:  "No analysis provided by OpenAI",
			Confidence: 0.5,
			Recommendations: []string{
				"Retry the analysis",
				"Check API configuration",
				"Monitor system performance",
			},
			RelatedChanges: []string{},
		}, nil
	}

	// Try to parse the response as JSON first
	var result AnalysisResult
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &result); err != nil {
		// If JSON parsing fails, log the error and create a placeholder result
		log.Printf("Error parsing LLM JSON response: %v\nRaw response: %s", err, resp.Choices[0].Message.Content)
		result = AnalysisResult{
			RootCause:       "Could not parse analysis result from LLM.",
			Confidence:      0.1,
			Recommendations: []string{"Check logs for the raw LLM response.", "Verify LLM prompt and response format."},
			RelatedChanges:  []string{},
		}
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

Recent GitHub Changes:
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

When referencing GitHub changes, use ONLY the actual commits and PRs provided in the "Recent GitHub Changes" section.
DO NOT make up or guess commit hashes, PR numbers, or changes that are not explicitly listed.

Format your response as a JSON object with the following structure:
{
    "root_cause": "string",
    "confidence": float,
    "recommendations": ["string"],
    "related_changes": ["string"]
}

IMPORTANT: Respond *only* with the valid JSON object requested above. Do not include any introductory text, explanations, summaries, or markdown formatting before or after the JSON.
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
	if len(changes) == 0 {
		return "No recent changes found."
	}

	var formatted string
	for _, change := range changes {
		if strings.HasPrefix(change.CommitID, "PR #") {
			formatted += fmt.Sprintf("- Pull Request %s in %s by %s at %s: %s\n",
				change.CommitID,
				change.Repo,
				change.Author,
				change.Timestamp.Format(time.RFC3339),
				change.Message,
			)
		} else {
			formatted += fmt.Sprintf("- Commit %s in %s by %s at %s: %s\n  Files changed: %s\n",
				change.CommitID,
				change.Repo,
				change.Author,
				change.Timestamp.Format(time.RFC3339),
				change.Message,
				strings.Join(change.FilesChanged, ", "),
			)
		}
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

func (a *Analyzer) formatExperiments(experiments []models.ExperimentPair) string {
	var formatted string
	for _, exp := range experiments {
		formatted += fmt.Sprintf("-ExperimentID: %s -PreviousExpermient: %s -CurrentExperiment: %s\n",
			exp.ExperimentID,
			exp.Previous,
			exp.Current,
		)
	}
	return formatted
}
