package models

type Conversation struct {
	BaseModel
	Name     string    `json:"name"`
	Messages []Message `json:"-"`
}

type CreateConversationInput struct {
	Body CreateConversationInputBody `json:"body"`
}

type CreateConversationInputBody struct {
	Name     string        `json:"name"`
	Messages []ChatMessage `json:"messages"`
}

type CreateConversationOutput struct {
	Body APIConversationOutput `json:"body"`
}

type APIConversationOutput struct {
	BaseModel
	Name     string        `json:"name"`
	Messages []ChatMessage `json:"messages"`
}

type GetConversationInput struct {
	Id string `path:"id"`
}

type GetConversationOutput struct {
	Body APIConversationOutput `json:"body"`
}

type GetConversationOutputBody struct {
	Conversation
}
