package models

// Base
type Agent struct {
	BaseModel
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Conversations []Conversation `json:"conversations"`
	Prompts       []Prompt       `json:"prompts"`
}

// Create
type AgentCreate struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AgentUpdate struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
