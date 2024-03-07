// Generated by server.go.tmpl
package service

import (
	"script_validation/models"
)

func (s *Service) GetMessageMetadata(id string) (*models.MessageMetadata, error) {
	messageMetadata := &models.MessageMetadata{BaseModel: models.BaseModel{ID: id}}
	tx := s.Db.First(messageMetadata)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return messageMetadata, nil
}

func (s *Service) CreateMessageMetadata(input models.MessageMetadataCreate) (*models.MessageMetadata, error) {
	messageMetadata := &models.MessageMetadata{}
	tx := s.Db.Create(messageMetadata)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return messageMetadata, nil
}

func (s *Service) GetAllMessageMetadatas() (*[]models.MessageMetadata, error) {
	messageMetadatas := &[]models.MessageMetadata{}
	tx := s.Db.Find(messageMetadatas)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return messageMetadatas, nil
}

func (s *Service) UpdateMessageMetadata(id string, input models.MessageMetadataUpdate) (*models.MessageMetadata, error) {
	messageMetadata := &models.MessageMetadata{BaseModel: models.BaseModel{ID: id}}
	tx := s.Db.First(messageMetadata)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// Apply the updates to the model
	// messageMetadata.Field = input.Field
	tx = s.Db.Updates(messageMetadata)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return messageMetadata, nil
}

func (s *Service) DeleteMessageMetadata(id string) (*models.MessageMetadata, error) {
	messageMetadata := &models.MessageMetadata{BaseModel: models.BaseModel{ID: id}}

	tx := s.Db.First(messageMetadata)
	if tx.Error != nil {
		return nil, tx.Error
	}
	tx = s.Db.Delete(messageMetadata)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return nil, nil
}

type MessageMetadataManager interface {
	GetMessageMetadata(id string) (*models.MessageMetadata, error)
	CreateMessageMetadata(*models.MessageMetadataCreate) (*models.MessageMetadata, error)
	GetAllMessageMetadatas() ([]*models.MessageMetadata, error)
	UpdateMessageMetadata(id string, input models.MessageMetadataUpdate) (*models.MessageMetadata, error)
	DeleteMessageMetadata(id string) (any, error)
}

/*
Append these structs to your models file if you don't have them already
type MessageMetadata struct {
	// TODO add ressources
	ID string `json:"id"`
}

type MessageMetadataCreate struct {
	// TODO add ressources
	ID string `json:"id"`
}

type MessageMetadataUpdate struct {
	// TODO add ressources
	ID string `json:"id"`
}
*/