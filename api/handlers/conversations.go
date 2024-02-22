package apihandlers

import (
	"context"
	"fmt"
	database "script_validation"
	"script_validation/internal/nomicai"
	"script_validation/models"

	"gorm.io/datatypes"
)

func SetMessageEmbeddings(messages *[]models.Message) {
	texts := make([]string, len(*messages))
	for i, message := range *messages {
		texts[i] = message.Content
	}
	embeddings, err := nomicai.EmbedText(texts, nomicai.Clustering)
	if err != nil {
		fmt.Println("Error embedding text: ", err)
		return
	}
	for i, embedding := range embeddings.Embeddings {
		(*messages)[i].Embedding = datatypes.NewJSONSlice(embedding)
	}
}

func CreateConversation(input *models.CreateConversationInput) (*models.Conversation, error) {
	// Make sure the last message in the conversation is from the assistant
	if len(input.Body.Messages) == 0 || input.Body.Messages[len(input.Body.Messages)-1].Role != "assistant" {
		return nil, fmt.Errorf("err: last message in conversation must be from the assistant")
	}

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
			MessageIndex:   uint(i),
		}
	}

	SetMessageEmbeddings(&messages)

	result = database.DB.Create(&messages)

	if result.Error != nil {
		return nil, result.Error
	}
	conversation.Messages = messages

	return &conversation, nil
}

func CreateConversationAPI(ctx context.Context, input *models.CreateConversationInput) (*models.CreateConversationOutput, error) {

	conversation, err := CreateConversation(input)
	if err != nil {
		return nil, err
	}

	// Create the messages
	messages := make([]models.APIChatMessage, len(input.Body.Messages))

	for i, message := range input.Body.Messages {
		messages[i] = models.APIChatMessage{
			ChatMessage: message,
			ID:          conversation.Messages[i].ID,
		}
	}

	// Return the conversation
	return &models.CreateConversationOutput{
		Body: models.APIConversationOutput{
			BaseModel: models.BaseModel{
				ID:        conversation.ID,
				CreatedAt: conversation.CreatedAt,
				UpdatedAt: conversation.UpdatedAt,
			},
			Name:     conversation.Name,
			Messages: messages,
		},
	}, nil
}

func GetConversation(id string) (*models.Conversation, error) {
	// Get the conversation
	conversation := models.Conversation{}

	result := database.DB.Preload("Messages").First(&conversation, "id = ?", id)

	if result.Error != nil {
		return nil, result.Error
	}

	// Return the conversation
	return &conversation, nil
}

func GetConversationAPI(ctx context.Context, input *models.GetConversationInput) (*models.CreateConversationOutput, error) {
	// Get the conversation
	conversation, err := GetConversation(input.Id)

	if err != nil {
		return nil, err
	}

	// Turn Messages into ChatMessages
	messages := make([]models.APIChatMessage, len(conversation.Messages))

	for i, message := range conversation.Messages {
		messages[i] = models.APIChatMessage{
			ChatMessage: message.ChatMessage,
			ID:          message.ID,
		}
	}

	// Return the conversation
	return &models.CreateConversationOutput{
		Body: models.APIConversationOutput{
			BaseModel: conversation.BaseModel,
			Name:      conversation.Name,
			Messages:  messages,
		},
	}, nil
}
