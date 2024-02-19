package models

import "gorm.io/gorm"

type Conversation struct {
	Name string `json:"name"`
}

type ConversationModel struct {
	gorm.Model
	Conversation
}

type CreateConversationInput struct {
	Body Conversation `json:"body"`
}

type CreateConversationOutput struct {
	Body ConversationModel `json:"body"`
}

type Message struct {
	Role           string `db:"role" json:"role"`
	Content        string `db:"content" json:"content"`
	ConversationId string `db:"conversation_id" json:"conversation_id"`
}

type MessageModel struct {
	gorm.Model
	Message
}
