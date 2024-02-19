package apihandlers

import (
	"context"
	database "script_validation"
	"script_validation/models"
)

func CreateConversation(ctx context.Context, input *models.CreateConversationInput) (*models.CreateConversationOutput, error) {
	// Create the conversation
	conversation := models.ConversationModel{
		Conversation: input.Body.Conversation,
	}

	result := database.DB.Create(&conversation)

	if result.Error != nil {
		return nil, result.Error
	}

	// Create the messages
	messages := make([]models.MessageModel, len(input.Body.Messages))

	for i, message := range input.Body.Messages {
		message.ConversationId = conversation.ID
		messages[i] = models.MessageModel{
			Msg: message,
		}
	}

	result = database.DB.Create(&messages)

	if result.Error != nil {
		return nil, result.Error
	}

	// Return the conversation
	return &models.CreateConversationOutput{
		Body: models.CreateConversationOutputBody{
			Conversation: conversation,
			Messages:     messages,
		},
	}, nil
}
