package apihandlers

import (
	"context"
	database "script_validation"
	"script_validation/models"
)

func CreateConversation(ctx context.Context, input *models.CreateConversationInput) (*models.CreateConversationOutput, error) {
	// Create the conversation
	data := models.ConversationModel{
		Conversation: input.Body,
	}

	result := database.DB.Create(&data)

	if result.Error != nil {
		return nil, result.Error
	}

	// Return the conversation
	return &models.CreateConversationOutput{
		Body: data,
	}, nil
}
