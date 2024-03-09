package models

import (
	"fmt"

	"gorm.io/datatypes"
)

type Embedding struct {
	datatypes.JSONSlice[float64]
}

type Message struct {
	BaseModel
	Role           string `example:"user" json:"role" enum:"user,system,assistant"`
	Content        string `example:"Hello, world!" json:"content"`
	MessageIndex   int
	ConversationID string
	PromptID       string
	LLMID          string
	TestMessageID  string
	TestMessages   []*Message `gorm:"foreignKey:TestMessageID" json:"-"` //
	Metadata       *MessageMetadata
	Score          float64 `gorm:"-"`
}

type MessageUpdate struct {
	ID string `json:"id"`
}

type MessageCreate struct {
	ID string `json:"id"`
}

func NewMessage(input *Message) (*Message, error) {
	if input.Role == "" || input.Content == "" || input.ConversationID == "" || input.PromptID == "" {
		return nil, fmt.Errorf("all fields are required")
	}
	return input, nil
}

func NewEvaluationMessage(input *Message) (*Message, error) {
	if input.Role == "" || input.Content == "" || input.ConversationID == "" || input.PromptID == "" {
		return nil, fmt.Errorf("all fields are required")
	}
	if input.Metadata.StartLatencyMs == 0 || input.Metadata.EndLatencyMs == 0 || input.Metadata.OutputTokenCount == 0 || input.Metadata.Embedding == nil {
		return nil, fmt.Errorf("start latency, end latency, and token count are required")
	}
	return input, nil
}
