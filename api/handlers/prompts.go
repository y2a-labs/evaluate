package apihandlers

import (
	"context"
	"fmt"
	database "script_validation"
	"script_validation/models"
)

func FindOrCreatePrompt(text string) (*models.Prompt, error) {
	prompt := models.Prompt{Content: text}
	if text == "" {
		return nil, fmt.Errorf("err: prompt content cannot be empty")
	}
	// Finds the prompt in the database
	result := database.DB.First(&prompt, "content = ?", text)
	if result.Error != nil {
		/*
			// Get the embedding
			embedding, err := nomicai.EmbedText([]string{text}, nomicai.Clustering)
			if err != nil {
				return nil, err
			}
			prompt.Embedding = embedding.Embeddings[0]
		*/
		result_2 := database.DB.Create(&prompt)
		if result_2.Error != nil {
			return nil, result_2.Error
		}
	}
	return &prompt, nil
}

func FindOrCreatePromptAPI(context context.Context, request *models.FindOrCreatePromptRequest) (*models.FindOrCreatePromptResponse, error) {
	prompt, err := FindOrCreatePrompt(request.Body.Content)

	if err != nil {
		return nil, err
	}
	return &models.FindOrCreatePromptResponse{
		Body: models.FindOrCreatePromptResponseBody{
			Content:   prompt.Content,
			BaseModel: prompt.BaseModel,
		},
	}, nil
}
