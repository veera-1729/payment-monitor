package contextbuilder

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
		config:      config,
		client:      &http.Client{Timeout: 10 * time.Second},
		redisClient: redis,
	}
}

func (b *ContextBuilder) BuildContext(ctx context.Context, alert *models.Alert) (*models.AnalysisContext, error) {
	analysisContext := &models.AnalysisContext{
		PaymentStats: &models.PaymentStats{
			Dimension:      alert.Dimension,
			Value:          alert.Value,
			SuccessRate:    alert.CurrentRate,
			PreviousRate:   alert.PreviousRate,
			DropPercentage: alert.DropPercentage,
			Timestamp:      alert.Timestamp,
		},
	}

	// Gather GitHub changes if token is provided
	if b.config.GitHubToken != "" {
		changes, err := b.getRecentChanges(ctx)
		if err != nil {
			fmt.Printf("Error getting GitHub changes: %v\n", err)
		} else {
			analysisContext.RecentChanges = changes
		}
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

func (b *ContextBuilder) getRecentChanges(ctx context.Context) ([]models.GitHubChange, error) {
	var allChanges []models.GitHubChange
	since := time.Now().Add(-time.Duration(b.config.LookbackHours) * time.Hour)

	for _, repo := range b.config.GitHubRepos {
		// Get commits
		commits, err := b.getRecentCommits(ctx, repo, since)
		if err != nil {
			fmt.Printf("Error getting commits for repo %s: %v\n", repo, err)
			continue
		}
		allChanges = append(allChanges, commits...)

		// Get pull requests
		prs, err := b.getRecentPRs(ctx, repo, since)
		if err != nil {
			fmt.Printf("Error getting PRs for repo %s: %v\n", repo, err)
			continue
		}
		allChanges = append(allChanges, prs...)
	}

	return allChanges, nil
}

func (b *ContextBuilder) getRecentCommits(ctx context.Context, repo string, since time.Time) ([]models.GitHubChange, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/commits?since=%s&per_page=%d",
		repo, since.Format(time.RFC3339), b.config.MaxCommitsPerRepo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+b.config.GitHubToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s (URL: %s)", resp.StatusCode, string(body), url)
	}

	var commits []struct {
		SHA    string `json:"sha"`
		Commit struct {
			Message string `json:"message"`
			Author  struct {
				Name string `json:"name"`
				Date string `json:"date"`
			} `json:"author"`
		} `json:"commit"`
		Files []struct {
			Filename string `json:"filename"`
		} `json:"files"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return nil, err
	}

	var changes []models.GitHubChange
	for _, commit := range commits {
		timestamp, _ := time.Parse(time.RFC3339, commit.Commit.Author.Date)
		var files []string
		for _, file := range commit.Files {
			files = append(files, file.Filename)
		}

		changes = append(changes, models.GitHubChange{
			Repo:         repo,
			CommitID:     commit.SHA,
			Author:       commit.Commit.Author.Name,
			Message:      commit.Commit.Message,
			Timestamp:    timestamp,
			FilesChanged: files,
		})
	}

	return changes, nil
}

func (b *ContextBuilder) getRecentPRs(ctx context.Context, repo string, since time.Time) ([]models.GitHubChange, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/pulls?state=all&sort=updated&direction=desc&per_page=%d",
		repo, b.config.MaxCommitsPerRepo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+b.config.GitHubToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s (URL: %s)", resp.StatusCode, string(body), url)
	}

	var prs []struct {
		Number    int    `json:"number"`
		Title     string `json:"title"`
		UpdatedAt string `json:"updated_at"`
		User      struct {
			Login string `json:"login"`
		} `json:"user"`
		Head struct {
			SHA string `json:"sha"`
		} `json:"head"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
		return nil, err
	}

	var changes []models.GitHubChange
	for _, pr := range prs {
		timestamp, _ := time.Parse(time.RFC3339, pr.UpdatedAt)
		if timestamp.Before(since) {
			continue
		}

		changes = append(changes, models.GitHubChange{
			Repo:      repo,
			CommitID:  fmt.Sprintf("PR #%d", pr.Number),
			Author:    pr.User.Login,
			Message:   pr.Title,
			Timestamp: timestamp,
		})
	}

	return changes, nil
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
