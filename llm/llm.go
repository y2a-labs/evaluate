package llm

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
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
	clonedRequest.Header.Add("Helicone-Target-Url", "https://openrouter.ai/api/v1")
	clonedRequest.Header.Add("Helicone-Auth", "Bearer "+os.Getenv("LLM_PROXY_API_KEY"))
	clonedRequest.Header.Set("Content-Type", "application/json")
	clonedRequest.Header.Set("Authorization", "Bearer "+os.Getenv("LLM_API_KEY"))
	// Add your custom headers here
	for key, value := range c.CustomerHeaders {
		clonedRequest.Header.Add(key, value)
	}

	return c.Transport.RoundTrip(clonedRequest)
}

type RateLimit struct {
	requests int
}

func GetRateLimit() {

}

func GetLLMClient(customHeaders map[string]string) *openai.Client {
	godotenv.Load()
	cid := uuid.New().String()
	customHeaders["Helicone-Property-cid"] = cid
	apiKey := os.Getenv("LLM_API_KEY")
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://gateway.hconeai.com/api/v1"
	config.HTTPClient.Transport = &CustomTransport{Transport: http.DefaultTransport, CustomerHeaders: customHeaders}
	return openai.NewClientWithConfig(config)
}

func GetLLMResponse(client *openai.Client, messages []openai.ChatCompletionMessage, model string) (openai.ChatCompletionMessage, error) {

	// Throw an error if the last message is from the assistant
	if len(messages) > 0 && messages[len(messages)-1].Role == "assistant" {
		return openai.ChatCompletionMessage{}, fmt.Errorf("err: the last message must be from the user")
	}

	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		Temperature: 0.7,
	})
	if err != nil {
		panic(err)
	}
	return resp.Choices[0].Message, nil
}

func GetUsersMessageVariant(client *openai.Client, message string) openai.ChatCompletionMessage {
	user_variant_messages := []openai.ChatCompletionMessage{
		{Role: "system", Content: "Create a variant of the following message."},
		{Role: "user", Content: "Create a variant of the following message. You are not an assistant. Do not ask if the user has any questions.\n\n" + message},
	}
	resp, err := GetLLMResponse(client, user_variant_messages, "openchat/openchat-7b")
	if err != nil {
		panic(err)
	}
	return resp
}
