// embedtext.go
package nomicai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
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

// cosineSimilarity calculates the cosine similarity between two float slices.
func CosineSimilarity(a, b []float64) (float64, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("slices must be of the same length")
	}

	dotProduct := 0.0
	magA := 0.0
	magB := 0.0

	for i := range a {
		dotProduct += a[i] * b[i] // Calculate dot product
		magA += a[i] * a[i]       // Sum for magnitude of a
		magB += b[i] * b[i]       // Sum for magnitude of b
	}

	magA = math.Sqrt(magA) // Calculate magnitude of a
	magB = math.Sqrt(magB) // Calculate magnitude of b

	if magA == 0 || magB == 0 {
		return 0, fmt.Errorf("one or both of the vectors are zero vectors")
	}

	return dotProduct / (magA * magB), nil // Calculate cosine similarity
}

func GetTextSimilarity(text1 string, text2 string) (float64, error) {
	embeddings, err := EmbedText(
		[]string{text1, text2},
		Clustering,
	)
	if err != nil {
		return 0, err
	}
	similarity, err := CosineSimilarity(embeddings.Embeddings[0], embeddings.Embeddings[1])
	if err != nil {
		panic(err)
	}
	return similarity, nil
}

// EmbedText sends texts to the Nomic AI Embed Text API for a specific task and returns the embedded representations.
func EmbedText(texts []string, taskType TaskType) (*EmbedTextResponse, error) {
	apiKey := os.Getenv("NOMICAI_API_KEY")
	apiURL := "https://api-atlas.nomic.ai/v1/embedding/text"

	payload := EmbedTextRequest{
		Texts:    texts,
		TaskType: taskType,
	}
	// Validate the input
	for i, text := range texts {
		if len(text) == 0 {
			fmt.Println(i)
			return nil, fmt.Errorf("text cannot be empty")
		}
	}

	// Validate the API key
	if len(apiKey) == 0 {
		return nil, fmt.Errorf("NOMICAI_API_KEY environment variable not set")
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
	req.Header.Add("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
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
