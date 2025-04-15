package contextbuilder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/yourusername/payment-monitor/pkg/models"
)

func NewContextBuilder(config *Config, redis *redis.Client) *ContextBuilder {
	if config.MaxCommitsPerRepo == 0 {
		config.MaxCommitsPerRepo = 10
	}
	if config.LookbackHours == 0 {
		config.LookbackHours = 24
	}
	return &ContextBuilder{
		config: config,
		client: &http.Client{Timeout: 10 * time.Second},
		redisClient: redis,
	}
}

func (b *ContextBuilder) BuildContext(ctx context.Context, alert *models.Alert) (*models.AnalysisContext, error) {
	analysisContext := &models.AnalysisContext{
		PaymentStats: &models.PaymentStats{
			Dimension:     alert.Dimension,
			Value:         alert.Value,
			SuccessRate:   alert.CurrentRate,
			PreviousRate:  alert.PreviousRate,
			DropPercentage: alert.DropPercentage,
			Timestamp:     alert.Timestamp,
		},
	}

	// Gather GitHub changes from all repositories
	if b.config.GitHubToken != "" && len(b.config.GitHubRepos) > 0 {
		allChanges := make([]models.GitHubChange, 0)
		for _, repo := range b.config.GitHubRepos {
			changes, err := b.getRecentGitHubChanges(ctx, repo)
			if err != nil {
				fmt.Printf("Error getting GitHub changes for repo %s: %v\n", repo, err)
				continue
			}
			allChanges = append(allChanges, changes...)
		}
		analysisContext.RecentChanges = allChanges
	}

	// Gather log entries
	if b.config.LogPath != "" {
		logs, err := b.getRecentLogs(ctx)
		if err != nil {
			fmt.Printf("Error getting logs: %v\n", err)
		} else {
			analysisContext.LogEntries = logs
		}
	}

	// Gather experiment data
	if b.config.ExperimentURL != "" {
		analysisContext.Experiments = b.getActiveExperiments(ctx)
	}

	return analysisContext, nil
}

func (b *ContextBuilder) getRecentGitHubChanges(ctx context.Context, repo string) ([]models.GitHubChange, error) {
	lookbackTime := time.Now().Add(-time.Duration(b.config.LookbackHours) * time.Hour)
	
	// Extract owner and repo name from the full URL
	parts := strings.Split(repo, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid repository format: %s. Expected format: owner/repo", repo)
	}
	owner := parts[len(parts)-2]
	repoName := parts[len(parts)-1]
	
	// Remove .git suffix if present
	repoName = strings.TrimSuffix(repoName, ".git")
	
	apiRepoPath := fmt.Sprintf("%s/%s", owner, repoName)
	
	url := fmt.Sprintf("https://api.github.com/repos/%s/commits?since=%s", 
		apiRepoPath, 
		lookbackTime.Format(time.RFC3339))
	
	fmt.Printf("Making GitHub API request to: %s\n", url)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", b.config.GitHubToken))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Read and log the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	fmt.Printf("GitHub API Response Body: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		var errorResponse struct {
			Message          string `json:"message"`
			DocumentationURL string `json:"documentation_url"`
			Status          string `json:"status"`
		}
		if err := json.Unmarshal(body, &errorResponse); err != nil {
			return nil, fmt.Errorf("GitHub API error (status %d): failed to parse error response: %v", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("GitHub API error (status %d): %s. Documentation: %s", 
			resp.StatusCode, 
			errorResponse.Message, 
			errorResponse.DocumentationURL)
	}

	// Create a new reader from the body for the JSON decoder
	reader := bytes.NewReader(body)

	var commits []struct {
		Sha    string `json:"sha"`
		Commit struct {
			Author struct {
				Name  string    `json:"name"`
				Date  time.Time `json:"date"`
			} `json:"author"`
			Message string `json:"message"`
		} `json:"commit"`
	}
	

	if err := json.NewDecoder(reader).Decode(&commits); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if len(commits) > b.config.MaxCommitsPerRepo {
		commits = commits[:b.config.MaxCommitsPerRepo]
	}

	var changes []models.GitHubChange
	for _, commit := range commits {
		fileChanges, err := b.getCommitFileChanges(ctx, repo, commit.Sha)
		if err != nil {
			fmt.Printf("Error getting file changes for commit %s in repo %s: %v\n", commit.Sha, repo, err)
			continue
		}

		if len(fileChanges) == 0 {
			continue
		}

		changes = append(changes, models.GitHubChange{
			Repo:         repo,
			CommitID:     commit.Sha,
			Author:       commit.Commit.Author.Name,
			Message:      commit.Commit.Message,
			Timestamp:    commit.Commit.Author.Date,
			FilesChanged: fileChanges,
		})
	}

	return changes, nil
}

func (b *ContextBuilder) getCommitFileChanges(ctx context.Context, repo, commitSHA string) ([]string, error) {
	// Extract owner and repo name from the full URL
	parts := strings.Split(repo, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid repository format: %s. Expected format: owner/repo", repo)
	}
	owner := parts[len(parts)-2]
	repoName := parts[len(parts)-1]
	
	// Remove .git suffix if present
	repoName = strings.TrimSuffix(repoName, ".git")
	
	apiRepoPath := fmt.Sprintf("%s/%s", owner, repoName)
	
	url := fmt.Sprintf("https://api.github.com/repos/%s/commits/%s", apiRepoPath, commitSHA)
	
	fmt.Printf("Making GitHub API request to: %s\n", url)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", b.config.GitHubToken))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read and log the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	fmt.Printf("GitHub API Response Body: in file changes%s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		var errorResponse struct {
			Message          string `json:"message"`
			DocumentationURL string `json:"documentation_url"`
			Status          string `json:"status"`
		}
		if err := json.Unmarshal(body, &errorResponse); err != nil {
			return nil, fmt.Errorf("GitHub API error (status %d): failed to parse error response: %v", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("GitHub API error (status %d): %s. Documentation: %s", 
			resp.StatusCode, 
			errorResponse.Message, 
			errorResponse.DocumentationURL)
	}

	var commitDetail struct {
		Files []struct {
			Filename    string `json:"filename"`
			Status      string `json:"status"`
			Additions   int    `json:"additions"`
			Deletions   int    `json:"deletions"`
			Changes     int    `json:"changes"`
			Patch       string `json:"patch"`
		} `json:"files"`
	}

	// Create a new reader from the body for the JSON decoder
	reader := bytes.NewReader(body)
	if err := json.NewDecoder(reader).Decode(&commitDetail); err != nil {
		return nil, err
	}

	var fileChanges []string
	for _, file := range commitDetail.Files {
		// if !isRelevantFile(file.Filename) {
		// 	continue
		// }
		fmt.Printf("file: %v\n", file)
		change := fmt.Sprintf("%s (%s): +%d -%d changes", 
			file.Filename, 
			file.Status,
			file.Additions,
			file.Deletions)
		
		if file.Status == "modified" && file.Patch != "" {
			lines := strings.Split(file.Patch, "\n")
			if len(lines) > 3 {
				lines = lines[:3]
			}
			change += "\n" + strings.Join(lines, "\n")
		}
		
		fileChanges = append(fileChanges, change)
		time.Sleep(10 * time.Second)
	}

	fmt.Printf("fileChanges: %v\n", fileChanges)

	return fileChanges, nil
}

func isRelevantFile(filename string) bool {
	relevantPatterns := []string{
		"payment",
		"transaction",
		"order",
		"checkout",
		"gateway",
		"processor",
		"api",
		"service",
		"controller",
	}

	excludePatterns := []string{
		"_test.go",
		"test_",
		"spec",
		"doc",
		"readme",
		"changelog",
	}

	filename = strings.ToLower(filename)
	
	for _, pattern := range excludePatterns {
		if strings.Contains(filename, pattern) {
			return false
		}
	}

	for _, pattern := range relevantPatterns {
		if strings.Contains(filename, pattern) {
			return true
		}
	}
	return false
}

func (b *ContextBuilder) getRecentLogs(ctx context.Context) ([]models.LogEntry, error) {
	// Implement log reading logic based on your logging system
	// This is a placeholder implementation
	return []models.LogEntry{}, nil
} 

// Handler for the /build-context endpoint
func (b *ContextBuilder) getActiveExperiments(ctx context.Context) []models.ExperimentPair {
    
    // Collect experiment pairs
    experimentPairs, err := b.CollectExperimentPairs(ctx)
    if err != nil {
        fmt.Printf("Error collecting experiment pairs: %v \n", err)
    }

	return experimentPairs
}