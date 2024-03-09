package models

import (
	"github.com/sashabaranov/go-openai"
	"gorm.io/datatypes"
)

type Conversation struct {
	BaseModel
	Name        string `json:"name"`
	Description string `json:"description"`
	Messages    []*Message
	ModelID     string
	PromptID    string `json:"prompt_id"`
	Prompt      Prompt `json:"prompt"`
	AgentID     string

	IsTest     bool
	TestModels datatypes.JSONSlice[TestModels]
	TestCount  int
}

type TestModels struct {
	Provider string
	Model    string
}

type ChatCompletionMessage struct {
	Role    string
	Content string
}

type ConversationCreate struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	AgentID       string
	PromptID      string
	MessageString string
	IsTest        bool
	Messages      []openai.ChatCompletionMessage
}

type ConversationUpdate struct {
	Name        string
	Description string
	Messages    []openai.ChatCompletionMessage
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
	Name               string `json:"name" example:"My Conversation"`
	ConversationString string `json:"conversation_string" example:"USER: Hello\nAI: Hi, what can I help you with today?!\nUSER: Do you know what the best way to make a cake is?\nAI: Yes, I do! I can help you with that."`
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
	Content      string `json:"content"`
	Role         string `json:"role"`
	MessageIndex uint   `json:"message_index"`
	ID           string `json:"id"`
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
