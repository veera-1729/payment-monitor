package contextbuilder

import (
	"net/http"

	"github.com/go-redis/redis"
	"github.com/yourusername/payment-monitor/pkg/config"
)

type ContextBuilder struct {
	config *Config
	client *http.Client
	redisClient *redis.Client
}

type Config struct {
	GitHubToken   string
	GitHubRepos   []string
	LogPath       string
	ExperimentURL string
	MaxCommitsPerRepo int
	LookbackHours     int
	SplitzToken   string
	ExperimentIds []config.ExperimentID
}