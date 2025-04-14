package contextbuilder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/yourusername/payment-monitor/pkg/models"
)

type ContextBuilder struct {
	config *Config
	client *http.Client
}

type Config struct {
	GitHubToken   string
	GitHubRepos   []string
	LogPath       string
	ExperimentURL string
	MaxCommitsPerRepo int
	LookbackHours     int
}

func NewContextBuilder(config *Config) *ContextBuilder {
	if config.MaxCommitsPerRepo == 0 {
		config.MaxCommitsPerRepo = 10
	}
	if config.LookbackHours == 0 {
		config.LookbackHours = 24
	}
	return &ContextBuilder{
		config: config,
		client: &http.Client{Timeout: 10 * time.Second},
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
		experiments, err := b.getActiveExperiments(ctx)
		if err != nil {
			fmt.Printf("Error getting experiments: %v\n", err)
		} else {
			analysisContext.Experiments = experiments
		}
	}

	return analysisContext, nil
}

func (b *ContextBuilder) getRecentGitHubChanges(ctx context.Context, repo string) ([]models.GitHubChange, error) {
	lookbackTime := time.Now().Add(-time.Duration(b.config.LookbackHours) * time.Hour)
	
	url := fmt.Sprintf("https://api.github.com/repos/%s/commits?since=%s", 
		repo, 
		lookbackTime.Format(time.RFC3339))
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", b.config.GitHubToken))
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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

	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return nil, err
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
	url := fmt.Sprintf("https://api.github.com/repos/%s/commits/%s", repo, commitSHA)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", b.config.GitHubToken))
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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

	if err := json.NewDecoder(resp.Body).Decode(&commitDetail); err != nil {
		return nil, err
	}

	var fileChanges []string
	for _, file := range commitDetail.Files {
		// if !isRelevantFile(file.Filename) {
		// 	continue
		// }

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
	}

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

func (b *ContextBuilder) getActiveExperiments(ctx context.Context) ([]models.Experiment, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", b.config.ExperimentURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var experiments []models.Experiment
	if err := json.NewDecoder(resp.Body).Decode(&experiments); err != nil {
		return nil, err
	}

	return experiments, nil
} 