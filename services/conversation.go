// Generated by server.go.tmpl
package service

import (
	"context"
	"fmt"
	"github.com/y2a-labs/evaluate/models"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

func (s *Service) GetConversation(id string) (*models.Conversation, error) {
	conversation := &models.Conversation{BaseModel: models.BaseModel{ID: id}}
	tx := s.Db.First(conversation)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return conversation, nil
}

func (s *Service) GetConversationWithMessages(id string, selectedVersion int) (*models.Conversation, error) {
	conversation := &models.Conversation{BaseModel: models.BaseModel{ID: id}}
	// Specify the condition for Messages preload
	tx := s.Db.First(conversation)
	if tx.Error != nil {
		return nil, tx.Error
	}

	// Select the version
	if selectedVersion == -1 {
		conversation.SelectedVersion = conversation.Version
	} else {
		conversation.SelectedVersion = selectedVersion
	}

	// Adjusted raw SQL query to fetch the primary messages based on conversation version
	sql := `
WITH RankedMessages AS (
    SELECT m.*,
           ROW_NUMBER() OVER(PARTITION BY m.message_index ORDER BY m.conversation_version DESC) as Rank
    FROM messages m
    WHERE m.conversation_id = ? AND m.conversation_version <= ? AND test_message_id = ''
)
SELECT * FROM RankedMessages WHERE Rank = 1 AND role <> '' ORDER BY message_index ASC;
`

	messages := []*models.Message{}
	if err := s.Db.Raw(sql, id, conversation.SelectedVersion).Scan(&messages).Error; err != nil {
		return nil, err
	}
	conversation.Messages = messages

	return conversation, nil
}

func (s *Service) CreateConversation(input models.ConversationCreate) (*models.Conversation, error) {
	conversation := &models.Conversation{
		Name:             input.Name,
		Description:      input.Description,
		ModelID:          input.LLMID,
		Version:          0,
		IsTest:           input.IsTest,
		LastMessageIndex: len(input.Messages),
	}

	if len(input.Messages) > 0 {
		// Create the messages
		messages := make([]*models.Message, len(input.Messages))
		for i, message := range input.Messages {
			messages[i] = &models.Message{
				BaseModel:           models.BaseModel{ID: uuid.NewString()},
				Role:                message.Role,
				Content:             message.Content,
				MessageIndex:        i,
				ConversationID:      conversation.ID,
				ConversationVersion: 0,
			}
		}

		conversation.Messages = messages
	}

	tx := s.Db.Create(conversation)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return conversation, nil
}

func (s *Service) GetAllConversations() ([]*models.Conversation, error) {
	conversations := []*models.Conversation{}
	tx := s.Db.Limit(25).Order("created_at DESC").Where("is_test = ?", false).Find(&conversations)
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
	if input.IsTest && !conversation.IsTest {
		// If the conversation is being marked as a test, generate embeddings for all of the messages.
		conversation.IsTest = input.IsTest
		messages := []*models.Message{}
		tx := s.Db.Where("conversation_id = ?", conversation.ID).Find(&messages)
		if tx.Error != nil {
			return nil, tx.Error
		}
		go appendMessageEmbeddings(messages, s)
	}
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
	embeddings, err := s.llmProviders["openai"].client.CreateEmbeddings(context.Background(), openai.EmbeddingRequestStrings{
		Model: "text-embedding-3-small",
		Input: texts,
	})
	if err != nil {
		return err
	}
	// add the embeddings to the messages
	messageMetadata := make([]*models.MessageMetadata, len(messages))
	for i, message := range messages {
		messageMetadata[i] = &models.MessageMetadata{
			MessageID: message.ID,
			Embedding: embeddings.Data[0].Embedding,
		}
	}
	// Update the messages in the database
	tx := s.Db.Create(messageMetadata)
	if tx.Error != nil {
		fmt.Println(tx.Error)
	}
	return nil
}

func (s *Service) AddMessagesToConversation(conversation *models.Conversation, inputMessages []models.ChatCompletionMessage) ([]*models.Message, error) {
	conversation.Version++

	// Create the messages from the input
	messages := make([]*models.Message, len(inputMessages))
	for i, message := range inputMessages {
		messages[i] = &models.Message{
			BaseModel:           models.BaseModel{ID: uuid.NewString()},
			Role:                message.Role,
			Content:             message.Content,
			MessageIndex:        conversation.LastMessageIndex + 1,
			ConversationID:      conversation.ID,
			ConversationVersion: conversation.Version,
		}
		conversation.LastMessageIndex++
	}
	//Adds the new messages to the database
	conversation.Messages = append(conversation.Messages, messages...)
	tx := s.Db.Save(conversation)
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
