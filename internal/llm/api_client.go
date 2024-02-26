package llm

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"script_validation/models"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

type CustomTransport struct {
	Transport       http.RoundTripper
	CustomerHeaders map[string]string
}

// RoundTrip executes a single HTTP transaction and allows you to modify
// the request before it's sent.
func (c *CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original request
	clonedRequest := req.Clone(req.Context())

	// Add your custom headers here
	//clonedRequest.Header.Add("Helicone-Target-Url", "https://openrouter.ai/api/v1")
	clonedRequest.Header.Add("Helicone-Auth", "Bearer "+os.Getenv("HELICONE_API_KEY"))
	clonedRequest.Header.Set("Content-Type", "application/json")
	//clonedRequest.Header.Set("Authorization", "Bearer "+os.Getenv("LLM_API_KEY"))
	// Add your custom headers here
	for key, value := range c.CustomerHeaders {
		clonedRequest.Header.Add(key, value)
	}

	return c.Transport.RoundTrip(clonedRequest)
}

func GetLLMClient(provider *models.Provider) *openai.Client {
	cid := uuid.New().String()
	customHeaders := map[string]string{
		"Helicone-Property-cid": cid,
	}
	apiKey := os.Getenv(provider.EnvKey)
	fmt.Println("API Key", apiKey)
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = provider.BaseUrl
	config.HTTPClient.Transport = &CustomTransport{Transport: http.DefaultTransport, CustomerHeaders: customHeaders}
	fmt.Println("baseurl", config.BaseURL)
	return openai.NewClientWithConfig(config)
}

func GetLLMResponse(client *openai.Client, messages []models.ChatMessage, model string) (models.ChatMessage, error) {

	oai_messages := []openai.ChatCompletionMessage{}

	for _, message := range messages {
		oai_messages = append(oai_messages, openai.ChatCompletionMessage{
			Role:    message.Role,
			Content: message.Content,
		})
	}

	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model:       model,
		Messages:    oai_messages,
		//Temperature: 0.7,
	})

	if err != nil {
		return models.ChatMessage{}, err
	}

	return models.ChatMessage{
		Role:    resp.Choices[0].Message.Role,
		Content: resp.Choices[0].Message.Content,
	}, nil
}
