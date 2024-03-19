package nomicai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
)

// TaskType defines the allowed task types for the Embed Text API.
type TaskType string

// Constants for TaskType to enforce allowed values.
const (
	SearchQuery    TaskType = "search_query"
	SearchDocument TaskType = "search_document"
	Classification TaskType = "classification"
	Clustering     TaskType = "clustering"
)

// EmbedTextRequest defines the payload structure for the Embed Text API.
type EmbedTextRequest struct {
	Texts    []string `json:"texts"`
	TaskType TaskType `json:"task_type"`
}

// EmbedTextResponse represents the structure of the response from the Embed Text API.
type EmbedTextResponse struct {
	Embeddings [][]float64    `json:"embeddings"`
	Usage      map[string]int `json:"usage"` // Assuming usage is a map with "total_tokens" key
}

type Client struct {
	apiKey  string
	BaseURL string
}

const defaultBaseURL = "https://api-atlas.nomic.ai/v1"

func NewClient(apiKey string, baseURLs ...string) (*Client, error) {
	baseURL := defaultBaseURL
	if len(baseURLs) > 0 {
		baseURL = baseURLs[0]
	}
	if apiKey == "" {
		return nil, fmt.Errorf("apiKey cannot be empty")
	}

	return &Client{
		apiKey:  apiKey,
		BaseURL: baseURL,
	}, nil
}

func (c *Client) EmbedText(texts []string, taskType TaskType) (*EmbedTextResponse, error) {
	apiURL := c.BaseURL + "/embedding/text"

	payload := EmbedTextRequest{
		Texts:    texts,
		TaskType: taskType,
	}
	// Validate the input
	for _, text := range texts {
		if len(text) == 0 {
			return nil, fmt.Errorf("text cannot be empty, current text: %v", texts)
		}
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.apiKey)

	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	embeddingResponse := &EmbedTextResponse{}
	derr := json.NewDecoder(resp.Body).Decode(embeddingResponse)
	if derr != nil && derr != io.EOF {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: nomicai %s", resp.Status)
	}
	return embeddingResponse, nil
}

func (c *Client) CosineSimilarity(a, b []float32) (float64, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("slices must be of the same length")
	}

	dotProduct := 0.0
	magA := 0.0
	magB := 0.0

	for i := range a {
		af64 := float64(a[i])
		bf64 := float64(b[i])
		dotProduct += af64 * bf64 // Calculate dot product
		magA += af64 * af64       // Sum for magnitude of a
		magB += bf64 * bf64      // Sum for magnitude of b
	}

	magA = math.Sqrt(magA) // Calculate magnitude of a
	magB = math.Sqrt(magB) // Calculate magnitude of b

	if magA == 0 || magB == 0 {
		return 0, nil
	}

	return dotProduct / (magA * magB), nil // Calculate cosine similarity
}