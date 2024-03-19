package nomicai_test

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"script_validation/internal/nomicai"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmbedTest(t *testing.T) {
	embeddings := [][]float64{{0.1, 0.2}, {0.3, 0.4}}
	usage := map[string]int{"total_tokens": 100}

	response := &nomicai.EmbedTextResponse{
		Embeddings: embeddings,
		Usage:      usage,
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/embedding/text", r.URL.Path, "Expected request to /v1/embedding/text")
		assert.Equal(t, "POST", r.Method, "Expected POST request")
		assert.Equal(t, r.Header.Get("Content-Type"), "application/json", "Expected Content-Type: application/json header")
		w.WriteHeader(http.StatusOK)

		jsonData, _ := json.Marshal(response)

		w.Write(jsonData)
	}))
	defer server.Close()

	client, err := nomicai.NewClient("apiKey", server.URL)
	assert.Nil(t, err, "Expected no error when creating client")
	value, err := client.EmbedText([]string{"test", "test2"}, nomicai.Clustering)
	assert.Nil(t, err, "Should not return an error")
	assert.Equal(t, response, value, "Should return 100 total tokens")

	_, err = client.EmbedText([]string{""}, nomicai.Clustering)
	assert.NotNil(t, err, "Should return an error")

}

func TestCosineSimilarity(t *testing.T) {
	client, err := nomicai.NewClient("apiKey")
	assert.Nil(t, err, "Expected no error when creating client")

	// Test case: vectors of different lengths
	_, err = client.CosineSimilarity([]float32{1, 2}, []float32{1})
	assert.NotNil(t, err, "Expected error for vectors of different lengths")

	// Test case: cosine similarity of identical vectors
	result, err := client.CosineSimilarity([]float32{1, 2, 3}, []float32{1, 2, 3})
	assert.Nil(t, err, "Expected no error for identical vectors")
	assert.Equal(t, 1.0, result, "Expected cosine similarity of 1 for identical vectors")

	// Test case: cosine similarity of orthogonal vectors
	result, err = client.CosineSimilarity([]float32{1, 0}, []float32{0, 1})
	assert.Nil(t, err, "Expected no error for orthogonal vectors")
	assert.Equal(t, 0.0, result, "Expected cosine similarity of 0 for orthogonal vectors")

	// Test case: cosine similarity of arbitrary vectors
	result, err = client.CosineSimilarity([]float32{1, 2, 3}, []float32{4, 5, 6})
	assert.Nil(t, err, "Expected no error for arbitrary vectors")
	expectedResult := (1*4 + 2*5 + 3*6) / (math.Sqrt(1+4+9) * math.Sqrt(16+25+36))
	assert.Equal(t, expectedResult, result, "Expected correct cosine similarity for arbitrary vectors")
}
