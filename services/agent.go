// Generated by server.go.tmpl
package service

import (
	"fmt"
	"script_validation/models"
)

func (s *Service) GetAgent(id string) (*models.Agent, error) {
	agent := &models.Agent{BaseModel: models.BaseModel{ID: id}}
	tx := s.Db.First(agent)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return agent, nil
}

func (s *Service) CreateAgent(input models.AgentCreate) (*models.Agent, error) {
	agent := &models.Agent{
		Name: input.Name,
		Description: input.Description,
	}
	tx := s.Db.Create(agent)
	if tx.Error != nil {
		return nil, tx.Error
	}
	prompt := &models.Prompt{AgentID: agent.ID, Content: "You are a helpful assistant.", Version: 0}
	tx = s.Db.Create(prompt)
	if tx.Error != nil {
		return nil, tx.Error
	}
	fmt.Println(prompt)
	agent.Prompts = append(agent.Prompts, *prompt)
	fmt.Println("Agent", agent)
	return agent, nil
}

func (s *Service) GetAllAgents() (*[]models.Agent, error) {
	agents := &[]models.Agent{}
	tx := s.Db.Find(agents)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return agents, nil
}

func (s *Service) UpdateAgent(id string, input models.AgentUpdate) (*models.Agent, error) {
	agent := &models.Agent{BaseModel: models.BaseModel{ID: id}}
	tx := s.Db.First(agent)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// Apply the updates to the model
	agent.Name = input.Name
	agent.Description = input.Description
	tx = s.Db.Updates(agent)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return agent, nil
}

func (s *Service) DeleteAgent(id string) (*models.Agent, error) {
	agent := &models.Agent{BaseModel: models.BaseModel{ID: id}}

	tx := s.Db.First(agent)
	if tx.Error != nil {
		return nil, tx.Error
	}
	tx = s.Db.Delete(agent)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return nil, nil
}

type AgentManager interface {
	GetAgent(id string) (*models.Agent, error)
	CreateAgent(*models.AgentCreate) (*models.Agent, error)
	GetAllAgents() ([]*models.Agent, error)
	UpdateAgent(id string, input models.AgentUpdate) (*models.Agent, error)
	DeleteAgent(id string) (any, error)
}

/*
Append these structs to your models file if you don't have them already
type Agent struct {
	// TODO add ressources
	ID string `json:"id"`
}

type AgentCreate struct {
	// TODO add ressources
	ID string `json:"id"`
}

type AgentUpdate struct {
	// TODO add ressources
	ID string `json:"id"`
}
*/