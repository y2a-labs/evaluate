package service

import (
	"context"
	"script_validation/models"
	"strings"

	"github.com/sashabaranov/go-openai"
)

func (s *Service) getLLM(modelName string) (modelID, providerID string, err error) {
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
		return "", "", tx.Error
	}
	providerID = model.ProviderID
	modelID = model.ID
	return
}

func (s *Service) ProxyOpenaiStream(ctx context.Context, req openai.ChatCompletionRequest, agentId string) (*openai.ChatCompletionStream, *models.Conversation, error) {
	modelId, providerId, err := s.getLLM(req.Model)
	if err != nil {
		return nil, nil, err
	}
	req.Model = modelId

	provider, ok := s.llmProviders[providerId]
	if !ok {
		return nil, nil, err
	}

	conversation, err := s.CreateConversation(models.ConversationCreate{AgentID: agentId, Messages: req.Messages})
	if err != nil {
		return nil, nil, err
	}

	stream, err := provider.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return stream, conversation, nil
}

func (s *Service) ProxyOpenaiChat(ctx context.Context, req openai.ChatCompletionRequest, agentId string) (*openai.ChatCompletionResponse, *models.Conversation, error) {
	modelID, providerID, err := s.getLLM(req.Model)
	if err != nil {
		return nil, nil, err
	}
	req.Model = modelID

	provider, ok := s.llmProviders[providerID]
	if !ok {
		return nil, nil, err
	}

	conversation, err := s.CreateConversation(models.ConversationCreate{AgentID: agentId, Messages: req.Messages})
	if err != nil {
		return nil, nil, err
	}

	stream, err := provider.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return &stream, conversation, err
}
