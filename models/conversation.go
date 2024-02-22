package models

import "gorm.io/datatypes"

type Conversation struct {
	BaseModel
	Name          string
	Messages      []Message
	EvalTestCount int
	EvalPrompt    string
	EvalModels    datatypes.JSONSlice[string]
}

type EvalConfig struct {
	EvalModels    []string
	EvalTestCount int
	EvalPrompt    string
}

type CreateConversationInput struct {
	Body CreateConversationInputBody `json:"body"`
}

type CreateConversationInputBody struct {
	Name     string        `json:"name" example:"My Conversation"`
	Messages []ChatMessage `json:"messages"`
}

type CreateConversationOutput struct {
	Body APIConversationOutput `json:"body"`
}

type APIConversationOutput struct {
	BaseModel
	Name     string           `json:"name"`
	Messages []APIChatMessage `json:"messages"`
}

type APIChatMessage struct {
	ChatMessage
	ID string `json:"id"`
}

type GetConversationInput struct {
	Id string `path:"id"`
}

type GetConversationResponse struct {
	Body *Conversation `json:"body"`
}

type GetConversationResponseBody struct {
	*Conversation
}

type GetConversationResponseValidation struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Messages struct {
		Role      string    `json:"role"`
		Content   string    `json:"content"`
		ID        string    `json:"id"`
		Embedding []float64 `json:"embedding"`
	} `json:"messages"`
}
