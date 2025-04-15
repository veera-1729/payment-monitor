package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Monitoring struct {
		Interval     int     `yaml:"interval"`
		Thresholds   struct {
			SuccessRateDrop    float64 `yaml:"success_rate_drop"`
			MinTransactions    int     `yaml:"minimum_transactions"`
		} `yaml:"thresholds"`
		Dimensions   []struct {
			Name    string `yaml:"name"`
			Enabled bool   `yaml:"enabled"`
		} `yaml:"dimensions"`
	} `yaml:"monitoring"`

	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
		SSLMode  string `yaml:"sslmode"`
	} `yaml:"database"`

	LLM struct {
		APIKey   string `yaml:"api_key"`
		Model    string `yaml:"model"`
		Endpoint string `yaml:"endpoint"`
	} `yaml:"llm"`

	ContextBuilder struct {
		GitHub struct {
			Enabled bool     `yaml:"enabled"`
			Token   string   `yaml:"token"`
			Repos   []string `yaml:"repos"`
		} `yaml:"github"`
		Logs struct {
			Enabled bool   `yaml:"enabled"`
			Path    string `yaml:"path"`
		} `yaml:"logs"`
		Experiments struct {
			Enabled bool   `yaml:"enabled"`
			ApiUrl string `yaml:"api_url"`
			SplitzToken string`yaml:"splitz_token"`
			ExperimentIds []ExperimentID `yaml:"experiment_ids"`
		} `yaml:"experiments"`
	} `yaml:"context_builder"`

	Redis struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`
}

type ExperimentID struct {
    ID          string `yaml:"id"`
    Name        string `yaml:"name"`
    Description string `yaml:"description"`
}

// LoadConfig loads the configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}