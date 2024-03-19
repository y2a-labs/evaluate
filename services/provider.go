// Generated by server.go.tmpl
package service

import (
	"context"
	"fmt"
	"script_validation/internal/nomicai"
	"script_validation/models"

	"github.com/sashabaranov/go-openai"
)

func (s *Service) GetProvider(id string) (*models.Provider, error) {
	provider := &models.Provider{BaseModel: models.BaseModel{ID: id}}
	tx := s.Db.First(provider)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return provider, nil
}

func (s *Service) CreateProvider(input models.ProviderCreate) (*models.Provider, error) {
	// Create the encryption key
	aesKey, err := loadOrCreateAESKey(".env")
	if err != nil {
		return nil, err
	}
	encryptedApiKey, err := Encrypt(input.ApiKey, aesKey)
	if err != nil {
		return nil, err
	}

	// Create the provider
	provider := &models.Provider{
		BaseModel:       models.BaseModel{ID: input.Id},
		BaseUrl:         input.BaseUrl,
		Type:            input.Type,
		EncryptedAPIKey: encryptedApiKey,
		Requests:        input.Requests,
		Interval:        input.Interval,
		Unit:            input.Unit,
	}

	// Initialize the provider
	if provider.Type == "llm" {
		client := openai.NewClient(input.ApiKey, provider.BaseUrl)

		// Check to see if it works
		_, err := client.ListModels(context.Background())
		if err != nil {
			return nil, fmt.Errorf("was not able to connect to the Please verify your baseURL and API Key")
		}
		// sk-slpllry-hd4eewa-wfm4hhq-fsdshbi

		s.llmProviders[provider.ID] = &llmProvider{
			Provider: provider,
			client:   client,
		}
	}
	// Initialize the provider
	if provider.Type == "embedding" {
		client, err := nomicai.NewClient(input.ApiKey, provider.BaseUrl)
		if err != nil {
			return nil, err
		}

		s.embeddingProviders[provider.ID] = &embeddingProvider{
			Provider: provider,
			client:   client,
		}
	}

	tx := s.Db.Create(provider)
	if tx.Error != nil {
		return nil, tx.Error
	}

	//Initalizes the rate limiter
	s.limiter.GetLimiter(provider)

	return provider, nil
}

func (s *Service) GetAllProviders() ([]*models.Provider, error) {
	providers := []*models.Provider{}
	tx := s.Db.Find(&providers)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return providers, nil
}

func (s *Service) UpdateProvider(id string, input models.ProviderUpdate) (*models.Provider, error) {
	provider := &models.Provider{BaseModel: models.BaseModel{ID: id}}
	tx := s.Db.First(provider)
	if tx.Error != nil {
		return nil, tx.Error
	}

	// Apply the updates to the model
	if input.ApiKey != "" || input.ApiKey != ".................." {
		// Create the encryption key
		aesKey, err := loadOrCreateAESKey(".env")
		if err != nil {
			return nil, err
		}
		encryptedApiKey, err := Encrypt(input.ApiKey, aesKey)
		if err != nil {
			return nil, err
		}
		provider.EncryptedAPIKey = encryptedApiKey
	}

	if input.BaseUrl != "" {
		provider.BaseUrl = input.BaseUrl
	}

	if input.Type != "" {
		provider.Type = input.Type
	}

	if input.Requests != 0 {
		provider.Requests = input.Requests
	}

	if input.Interval != 0 {
		provider.Interval = input.Interval
	}

	if input.Unit != "" {
		provider.Unit = input.Unit
	}

	tx = s.Db.Save(provider)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return provider, nil
}

func (s *Service) DeleteProvider(id string) (*models.Provider, error) {
	provider := &models.Provider{BaseModel: models.BaseModel{ID: id}}

	tx := s.Db.First(provider)
	if tx.Error != nil {
		return nil, tx.Error
	}
	tx = s.Db.Delete(provider)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return nil, nil
}

type ProviderManager interface {
	GetProvider(id string) (*models.Provider, error)
	CreateProvider(*models.ProviderCreate) (*models.Provider, error)
	GetAllProviders() ([]*models.Provider, error)
	UpdateProvider(id string, input models.ProviderUpdate) (*models.Provider, error)
	DeleteProvider(id string) (any, error)
}
