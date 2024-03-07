package llm

import (
	"github.com/sashabaranov/go-openai"
)

func NewClient(baseUrl string, apiKey string) *openai.Client {
	// Create a new client
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseUrl
	return openai.NewClientWithConfig(config)
}
