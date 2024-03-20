// Generated by server.go.tmpl
package service

import (
	"github.com/y2a-labs/evaluate/models"

	"github.com/google/uuid"
)

func (s *Service) GetMessage(id string) (*models.Message, error) {
	message := &models.Message{BaseModel: models.BaseModel{ID: id}}
	tx := s.Db.Preload("Metadata").First(message)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return message, nil
}

func (s *Service) CreateMessage(input models.MessageCreate) (*models.Message, error) {
	message := &models.Message{}
	tx := s.Db.Create(message)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return message, nil
}

func (s *Service) GetAllMessages() (*[]models.Message, error) {
	messages := &[]models.Message{}
	tx := s.Db.Find(messages)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return messages, nil
}

func (s *Service) UpdateMessage(id string, input models.MessageUpdate) (*models.Message, error) {
	message, err := s.GetMessage(id)
	if err != nil {
		return nil, err
	}
	conversation, err := s.GetConversation(message.ConversationID)
	if err != nil {
		return nil, err
	}

	// Increment the version
	conversation.Version++

	// Apply the updates to the message
	if input.Role != "" {
		message.Role = input.Role
	}
	if input.Content != "" {
		message.Content = input.Content
	}

	// Generate a new message ID and update the version
	newMessage := message
	newMessage.ConversationVersion = conversation.Version
	newMessage.ID = uuid.NewString()

	s.Db.Save(conversation)
	s.Db.Create(newMessage)

	go appendMessageEmbeddings([]*models.Message{newMessage}, s)

	return newMessage, nil
}

func (s *Service) DeleteMessage(id string) error {
	message, err := s.GetMessage(id)
	if err != nil {
		return err
	}
	conversation, err := s.GetConversation(message.ConversationID)
	if err != nil {
		return err
	}
	// Increment the version
	conversation.Version++
	s.Db.Save(conversation)

	// Generate a new message ID and update the version
	newMessage := message
	newMessage.ConversationVersion = conversation.Version
	newMessage.ID = uuid.NewString()
	newMessage.Role = ""

	s.Db.Create(newMessage)

	return nil
}

type MessageManager interface {
	GetMessage(id string) (*models.Message, error)
	CreateMessage(*models.MessageCreate) (*models.Message, error)
	GetAllMessages() ([]*models.Message, error)
	UpdateMessage(id string, input models.MessageUpdate) (*models.Message, error)
	DeleteMessage(id string) (any, error)
}

/*
Append these structs to your models file if you don't have them already
type Message struct {
	// TODO add ressources
	ID string `json:"id"`
}

type MessageCreate struct {
	// TODO add ressources
	ID string `json:"id"`
}

type MessageUpdate struct {
	// TODO add ressources
	ID string `json:"id"`
}
*/
