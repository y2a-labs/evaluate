package models

import "gorm.io/datatypes"

type Embedding struct {
	datatypes.JSONSlice[float64]
}

type Message struct {
	BaseModel
	ChatMessage
	MessageIndex       uint
	ConversationID     string
	MessageEvaluations []MessageEvaluation
	Embedding		  datatypes.JSONSlice[float64]
}

func (message *Message) ConvertToChatMessage() ChatMessage {
	return ChatMessage{
		Role:    message.Role,
		Content: message.Content,
	}
}

type ChatMessage struct {
	Role    string `example:"user" json:"role" enum:"user,system,assistant"`
	Content string `example:"Hello, world!" json:"content"`
}
