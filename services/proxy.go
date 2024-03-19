package service

import (
	"context"
	"fmt"
	"script_validation/models"
	"strings"

	"github.com/sashabaranov/go-openai"
)

func (s *Service) GetModel(modelName string) (modelID, providerID string, err error) {
	// Check if the modelName can be split with a "/"
	parts := strings.SplitN(modelName, "/", 2)

	// If there was no provider specified
	if len(parts) == 2 {
		providerID = parts[0]
		provider := &models.Provider{BaseModel: models.BaseModel{ID: providerID}}

		// If the provider exists, return the provider and model
		tx := s.Db.First(provider)
		if tx.Error == nil {
			modelID = parts[1]
			return
		}
	}
	// When there was no valid provider specified
	model := &models.LLM{BaseModel: models.BaseModel{ID: modelName}}
	tx := s.Db.First(model)
	if tx.Error != nil {
		return "", "", fmt.Errorf("model not found: %s: %w", modelName, tx.Error)
	}
	providerID = model.ProviderID
	modelID = model.ID
	return
}

func (s *Service) ProxyOpenaiStream(ctx context.Context, req openai.ChatCompletionRequest, providerId string) (*openai.ChatCompletionStream, *models.Conversation, error) {
	// When the provider comes from the headers
	modelId := req.Model
	var err error

	if providerId == "" {
		modelId, providerId, err = s.GetModel(req.Model)
		if err != nil {
			return nil, nil, err
		}
	}

	req.Model = modelId

	provider, ok := s.llmProviders[providerId]
	if !ok {
		return nil, nil, err
	}

	conversation, err := s.CreateConversation(models.ConversationCreate{Messages: req.Messages, LLMID: req.Model})
	if err != nil {
		return nil, nil, err
	}

	stream, err := provider.client.CreateChatCompletionStream(ctx, req)

	if err != nil {
		return nil, nil, err
	}

	return stream, conversation, nil
}

func (s *Service) ProxyOpenaiChat(ctx context.Context, req openai.ChatCompletionRequest, providerId string) (*openai.ChatCompletionResponse, *models.Conversation, error) {
	// When the provider comes from the headers
	modelId := req.Model
	var err error

	if providerId == "" {
		modelId, providerId, err = s.GetModel(req.Model)
		if err != nil {
			return nil, nil, err
		}
	}

	req.Model = modelId

	provider, ok := s.llmProviders[providerId]
	if !ok {
		return nil, nil, err
	}

	conversation, err := s.CreateConversation(models.ConversationCreate{Messages: req.Messages, LLMID: req.Model})
	if err != nil {
		return nil, nil, err
	}

	stream, err := provider.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return &stream, conversation, err
}

func (s *Service) ProxyOpenaiEmbedding(ctx context.Context, req openai.EmbeddingRequest) (*openai.EmbeddingResponse, error) {
	modelID, providerId, err := s.GetModel(string(req.Model))
	if err != nil {
		return nil, err
	}
	req.Model = openai.EmbeddingModel(modelID)

	provider, ok := s.llmProviders[providerId]
	if !ok {
		return nil, fmt.Errorf("provider not found")
	}
	resp, err := provider.client.CreateEmbeddings(ctx, req)
	return &resp, err
}