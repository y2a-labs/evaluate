package apihandlers

import (
	"context"
	database "script_validation"
	"script_validation/models"
)

func CreateConversation(ctx context.Context, input *models.CreateConversationInput) (*models.CreateConversationOutput, error) {

	// Create the conversation
	conversation := models.Conversation{Name: input.Body.Name}

	result := database.DB.Create(&conversation)

	if result.Error != nil {
		return nil, result.Error
	}

	// Create the messages
	messages := make([]models.Message, len(input.Body.Messages))

	for i, message := range input.Body.Messages {
		messages[i] = models.Message{
			ConversationID: conversation.ID,
			ChatMessage:    message,
		}
	}

	result = database.DB.Create(&messages)

	if result.Error != nil {
		return nil, result.Error
	}

	// Return the conversation
	return &models.CreateConversationOutput{
		Body: models.APIConversationOutput{
			BaseModel: conversation.BaseModel,
			Name:      conversation.Name,
			Messages:  input.Body.Messages,
		},
	}, nil
}

func GetConversation(ctx context.Context, input *models.GetConversationInput) (*models.GetConversationOutput, error) {
	// Get the conversation
	conversation := models.Conversation{}

	result := database.DB.Preload("Messages").First(&conversation, "id = ?", input.Id)

	if result.Error != nil {
		return nil, result.Error
	}

	chatMessages := make([]models.ChatMessage, len(conversation.Messages))
	for i, message := range conversation.Messages {
		chatMessages[i] = models.ChatMessage{
			Role:    message.Role,
			Content: message.Content,
		}
	}

	// Return the conversation
	return &models.GetConversationOutput{
		Body: models.APIConversationOutput{
			BaseModel: conversation.BaseModel,
			Name:      conversation.Name,
			Messages:  chatMessages,
		},
	}, nil
}
