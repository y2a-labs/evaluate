package models

type Message struct {
	BaseModel
	ChatMessage
	MessageIndex   uint
	ConversationID string
}

func (message *Message) ConvertToChatMessage() ChatMessage {
	return ChatMessage{
		Role:    message.Role,
		Content: message.Content,
	}
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
