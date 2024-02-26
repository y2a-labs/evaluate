package apihandlers

import (
	"context"
	"os"
	database "script_validation"
	"script_validation/internal/nomicai"
	"script_validation/models"
	"time"
)

func UpdateMessage(message *models.Message, embeddingClient *nomicai.Client) (*models.Message, error) {
	embedding, err := embeddingClient.EmbedText([]string{message.Content}, nomicai.Clustering)
	if err != nil {
		return nil, err
	}
	message.Embedding = embedding.Embeddings[0]
	tx := database.DB.Updates(&message)
	if tx.Error != nil {
		return nil, tx.Error
	}
	tx = database.DB.First(&message, "id = ?", message.ID)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return message, nil
}


func UpdateMessageAPI(ctx context.Context, request *UpdateMessageRequest) (*UpdateMessageResponse, error) {
	message := &models.Message{BaseModel: models.BaseModel{ID: request.ID},
		ChatMessage: models.ChatMessage{
			Content: request.Body.Content,
		}}
	embeddingClient := nomicai.NewClient(os.Getenv("NOMICAI_API_KEY"))

	message, err := UpdateMessage(message, embeddingClient)
	if err != nil {
		return nil, err
	}
	return &UpdateMessageResponse{
		Body: UpdateMessageResponseBody{
			Content:      message.Content,
			Role:         message.Role,
			ID:           message.ID,
			MessageIndex: message.MessageIndex,
			CreatedAt:    message.CreatedAt,
			UpdatedAt:    message.UpdatedAt,
		},
	}, nil
}

type UpdateMessageRequest struct {
	ID   string                   `path:"id" validate:"required"`
	Body UpdateMessageRequestBody `json:"body"`
}

type UpdateMessageRequestBody struct {
	Content string `json:"content"`
}

type UpdateMessageResponse struct {
	Body UpdateMessageResponseBody `json:"body"`
}

type UpdateMessageResponseBody struct {
	Content      string    `json:"content"`
	Role         string    `json:"role"`
	ID           string    `json:"id"`
	MessageIndex uint      `json:"message_index"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}