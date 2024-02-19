package models

type Conversation struct {
	Name string `json:"name"`
}

type ConversationModel struct {
	BaseModel
	Conversation
}

type CreateConversationInput struct {
	Body CreateConversationInputBody `json:"body"`
}

type CreateConversationInputBody struct {
	Conversation
	Messages []Msg `json:"messages"`
}

type CreateConversationOutput struct {
	Body CreateConversationOutputBody `json:"body"`
}

type CreateConversationOutputBody struct {
	Conversation ConversationModel `json:"conversation"`
	Messages     []MessageModel    `json:"messages"`
}
