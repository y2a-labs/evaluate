// Generated by server.go.tmpl
package service

import (
	"fmt"
	"script_validation/internal/nomicai"
	"script_validation/models"
)

func (s *Service) GetConversation(id string) (*models.Conversation, error) {
	conversation := &models.Conversation{BaseModel: models.BaseModel{ID: id}}
	tx := s.Db.First(conversation)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return conversation, nil
}

func (s *Service) GetConversationWithMessages(id string) (*models.Conversation, error) {
	conversation := &models.Conversation{BaseModel: models.BaseModel{ID: id}}
	// Specify the condition for Messages preload
	tx := s.Db.Preload("Messages", "test_message_id IS (?) OR test_message_id IS NULL", "").First(conversation)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return conversation, nil
}

func (s *Service) CreateConversation(input models.ConversationCreate) (*models.Conversation, error) {
	conversation := &models.Conversation{
		Name:        input.Name,
		Description: input.Description,
		AgentID:     input.AgentID,
		PromptID:    input.PromptID,
		IsTest:      input.IsTest,
	}
	tx := s.Db.Create(conversation)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return conversation, nil
}

func (s *Service) GetAllConversations() (*[]models.Conversation, error) {
	conversations := &[]models.Conversation{}
	tx := s.Db.Find(conversations)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return conversations, nil
}

func (s *Service) UpdateConversation(id string, input models.ConversationUpdate) (*models.Conversation, error) {
	conversation := &models.Conversation{BaseModel: models.BaseModel{ID: id}}
	tx := s.Db.First(conversation)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// Apply the updates to the model
	conversation.Name = input.Name
	conversation.Description = input.Description
	tx = s.Db.Updates(conversation)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return conversation, nil
}

func appendMessageEmbeddings(messages []*models.Message, s *Service) error {
	texts := make([]string, len(messages))
	for i, message := range messages {
		texts[i] = message.Content
	}
	// add the text embeddings
	embeddings, err := s.embeddingProviders["nomicai"].client.EmbedText(texts, nomicai.Clustering)
	if err != nil {
		fmt.Println(err)
	}
	// add the embeddings to the messages
	messageMetadata := make([]*models.MessageMetadata, len(messages))
	for i, message := range messages {
		messageMetadata[i] = &models.MessageMetadata{
			MessageID: message.ID,
			Embedding: embeddings.Embeddings[i],
		}
	}
	// Update the messages in the database
	tx := s.Db.Create(messageMetadata)
	if tx.Error != nil {
		fmt.Println(tx.Error)
	}
	return nil
}

func (s *Service) AddMessagesToConversation(conversationId string, inputMessages []models.ChatCompletionMessage) ([]*models.Message, error) {
	// Gets the number of messages in the conversation
	var messageIndex int64
	tx := s.Db.Model(&models.Message{}).Where("conversation_id = ?", conversationId).Where("test_message_id = ?", "").Count(&messageIndex)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// Create the messages from the input
	messages := make([]*models.Message, len(inputMessages))
	for i, message := range inputMessages {
		messages[i] = &models.Message{
			Role:           message.Role,
			Content:        message.Content,
			MessageIndex:   int(messageIndex),
			ConversationID: conversationId,
		}
		messageIndex++
	}
	//Adds the new messages to the database
	tx = s.Db.Create(messages)
	if tx.Error != nil {
		return nil, tx.Error
	}

	// Generate the embeddings in the background
	go appendMessageEmbeddings(messages, s)

	return messages, nil
}

func (s *Service) DeleteConversation(id string) (*models.Conversation, error) {
	conversation := &models.Conversation{BaseModel: models.BaseModel{ID: id}}

	tx := s.Db.First(conversation)
	if tx.Error != nil {
		return nil, tx.Error
	}
	tx = s.Db.Delete(conversation)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return nil, nil
}

type ConversationManager interface {
	GetConversation(id string) (*models.Conversation, error)
	CreateConversation(*models.ConversationCreate) (*models.Conversation, error)
	GetAllConversations() ([]*models.Conversation, error)
	UpdateConversation(id string, input models.ConversationUpdate) (*models.Conversation, error)
	DeleteConversation(id string) (any, error)
	AddMessagesToConversation(conversationId string, input models.ConversationUpdate) ([]models.Message, error)
}

/*
Append these structs to your models file if you don't have them already
type Conversation struct {
	// TODO add ressources
	ID string `json:"id"`
}

type ConversationCreate struct {
	// TODO add ressources
	ID string `json:"id"`
}

type ConversationUpdate struct {
	// TODO add ressources
	ID string `json:"id"`
}
*/