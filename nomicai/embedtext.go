// embedtext.go
package nomicai

import (
	"bytes"
	"encoding/json"
	"io"
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

// EmbedText sends texts to the Nomic AI Embed Text API for a specific task and returns the embedded representations.
func EmbedText(apiKey string, texts []string, taskType TaskType) (*EmbedTextResponse, error) {
	apiURL := "https://api-atlas.nomic.ai/v1/embedding/text"

	payload := EmbedTextRequest{
		Texts:    texts,
		TaskType: taskType,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		panic(err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	embeddingResponse := &EmbedTextResponse{}
	derr := json.NewDecoder(resp.Body).Decode(embeddingResponse)
	if derr != nil && derr != io.EOF {
		panic(derr)
	}

	if resp.StatusCode != http.StatusOK {
		panic(resp.Status)
	}
	return embeddingResponse, nil
}
