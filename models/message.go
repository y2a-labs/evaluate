package models

import (
	"gorm.io/datatypes"
)

type Embedding struct {
	datatypes.JSONSlice[float64]
}
type Message struct {
	BaseModel
	Role                string `example:"user" json:"role"`
	Content             string `example:"Hello, world!" json:"content"`
	MessageIndex        int
	ConversationID      string
	PromptID            string
	LLMID               string
	ConversationVersion int `gorm:"default:0"`
	TestMessageID       string
	TestMessages        []*Message `gorm:"foreignKey:TestMessageID" json:"-"` //
	Metadata            *MessageMetadata
	Score               float64 `gorm:"-"`
	Count               int `gorm:"-"`
}

type MessageUpdate struct {
	Content string
	Role    string
}

type MessageCreate struct {
	ID string `json:"id"`
}
