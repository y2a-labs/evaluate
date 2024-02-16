package utils

import (
	"fmt"
	"math"
	"strings"

	"github.com/sashabaranov/go-openai"
)

func ScriptToConversation(script string) []openai.ChatCompletionMessage {
	// Split the conversation by lines
	lines := strings.Split(script, "\n")

	// Initialize variables
	var messages []openai.ChatCompletionMessage
	var currentRole, content string

	// Iterate through each line in the conversation
	for _, line := range lines {
		// Check for the role in the line
		switch {
		case strings.HasPrefix(line, "SYSTEM:"):
			currentRole = "system"
			content = strings.TrimPrefix(line, "SYSTEM:")
		case strings.HasPrefix(line, "USER:"):
			currentRole = "user"
			content = strings.TrimPrefix(line, "USER:")
		case strings.HasPrefix(line, "ASSISTANT:"):
			currentRole = "assistant"
			content = strings.TrimPrefix(line, "ASSISTANT:")
		default:
			// Continuation of the previous message
			content = line
		}

		// Trim whitespace
		content = strings.TrimSpace(content)

		// Skip empty lines
		if content == "" {
			continue
		}

		// Add the message to the payload
		if len(messages) > 0 && messages[len(messages)-1].Role == currentRole {
			// Append content to the last message if the role is the same
			messages[len(messages)-1].Content += " " + content
		} else {
			// Add a new message to the messages
			messages = append(messages, openai.ChatCompletionMessage{Role: currentRole, Content: content})
		}
	}

	return messages
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