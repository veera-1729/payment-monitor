package analyzer

import (
	"strings"

	"github.com/sashabaranov/go-openai"
)

type Config struct {
	APIKey         string
	Endpoint       string
	DeploymentName string
}

type Analyzer struct {
	client         *openai.Client
	deploymentName string
}

func NewAnalyzer(config Config) (*Analyzer, error) {
	clientConfig := openai.DefaultAzureConfig(config.APIKey, config.Endpoint)

	// Ensure endpoint has proper scheme
	if !strings.HasPrefix(config.Endpoint, "https://") {
		config.Endpoint = "https://" + config.Endpoint
	}

	// Set the base URL for Azure OpenAI
	clientConfig.BaseURL = config.Endpoint
	clientConfig.APIVersion = "2023-05-15" // Use a stable API version
	clientConfig.AzureModelMapperFunc = func(model string) string {
		return config.DeploymentName // Use the deployment name from config
	}

	client := openai.NewClientWithConfig(clientConfig)
	return &Analyzer{
		client:         client,
		deploymentName: config.DeploymentName,
	}, nil
}
