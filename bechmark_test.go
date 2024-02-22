package database

import (
	"script_validation/models"
	"testing"
)

var conversation = models.Conversation{
	Name: "Test Conversation",
	Messages: []models.Message{
		{
			BaseModel: models.BaseModel{
				ID: "1",
			},
			ChatMessage: models.ChatMessage{
				Role:    "system",
				Content: "test prompt",
			},
			MessageEvaluations: make([]models.MessageEvaluation, 4),
		},
		{
			BaseModel: models.BaseModel{
				ID: "2",
			},
			ChatMessage: models.ChatMessage{
				Role:    "user",
				Content: "test content",
			},
			MessageEvaluations: make([]models.MessageEvaluation, 4),
		},
		{
			BaseModel: models.BaseModel{
				ID: "4",
			},
			ChatMessage: models.ChatMessage{
				Role:    "assistant",
				Content: "test content",
			},
			MessageEvaluations: make([]models.MessageEvaluation, 4),
		},
		{
			BaseModel: models.BaseModel{
				ID: "5",
			},
			ChatMessage: models.ChatMessage{
				Role:    "user",
				Content: "test content",
			},
			MessageEvaluations: make([]models.MessageEvaluation, 4),
		},
		{
			BaseModel: models.BaseModel{
				ID: "6",
			},
			ChatMessage: models.ChatMessage{
				Role:    "assistant",
				Content: "test content",
			},
			MessageEvaluations: make([]models.MessageEvaluation, 4),
		},
	},

}

var eval = []models.MessageEvaluation{
	{
		MessageID: "4",
		MessageEvaluationResults: []models.MessageEvaluationResult{},
	},
	{
		MessageID: "4",
		MessageEvaluationResults: []models.MessageEvaluationResult{},
	},
	{
		MessageID: "4",
		MessageEvaluationResults: []models.MessageEvaluationResult{},
	},
	{
		MessageID: "4",
		MessageEvaluationResults: []models.MessageEvaluationResult{},
	},
	{
		MessageID: "6",
		MessageEvaluationResults: []models.MessageEvaluationResult{},
	},
	{
		MessageID: "6",
		MessageEvaluationResults: []models.MessageEvaluationResult{},
	},
	{
		MessageID: "6",
		MessageEvaluationResults: []models.MessageEvaluationResult{},
	},
	{
		MessageID: "6",
		MessageEvaluationResults: []models.MessageEvaluationResult{},
	},
}

func BenchmarkOriginal(b *testing.B) {
	evaluations := eval
	for i := 0; i < b.N; i++ {
		for i, message := range conversation.Messages {
			if message.ChatMessage.Role != "assistant" {
				continue
			}
			// Loop through the evaluations and find the one that matches the message
			for _, evaluation := range evaluations {
				if evaluation.MessageID == message.ID {
					conversation.Messages[i].MessageEvaluations = append(conversation.Messages[i].MessageEvaluations, evaluation)
				}
			}
		}
	}
}
func BenchmarkModified(b *testing.B) {
	// Create a map of evaluations by MessageID
	evalMap := make(map[string][]models.MessageEvaluation)
	for i := 0; i < len(eval); i += 4 {
		messageID := eval[i].MessageID
		evalMap[messageID] = eval[i : i+4]
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for i, message := range conversation.Messages {
			if message.ChatMessage.Role != "assistant" {
				continue
			}
			// Look up the evaluations in the map
			if evals, ok := evalMap[message.ID]; ok {
				conversation.Messages[i].MessageEvaluations = evals
			}
		}
	}
}