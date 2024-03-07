// Generated controller.go.tmpl
package api

import (
	"github.com/go-fuego/fuego"
	"script_validation/models"
)

func (rs Resources) RegisterAgentRoutes(s *fuego.Server) {
	AgentGroup := fuego.Group(s, "/agent")

	fuego.Get(AgentGroup, "/", rs.getAllAgents)
	fuego.Post(AgentGroup, "/", rs.createAgent)

	fuego.Get(AgentGroup, "/{id}", rs.getAgent)
	fuego.Put(AgentGroup, "/{id}", rs.updateAgent)
	fuego.Delete(AgentGroup, "/{id}", rs.deleteAgent)
}

func (rs Resources) getAllAgents(c fuego.ContextNoBody) (*[]models.Agent, error) {
	return rs.Service.GetAllAgents()
}

func (rs Resources) createAgent(c *fuego.ContextWithBody[models.AgentCreate]) (*models.Agent, error) {
	body, err := c.Body()
	if err != nil {
		return &models.Agent{}, err
	}

	return rs.Service.CreateAgent(body)
}

func (rs Resources) getAgent(c fuego.ContextNoBody) (*models.Agent, error) {
	id := c.PathParam("id")

	return rs.Service.GetAgent(id)
}

func (rs Resources) updateAgent(c *fuego.ContextWithBody[models.AgentUpdate]) (*models.Agent, error) {
	id := c.PathParam("id")

	body, err := c.Body()
	if err != nil {
		return &models.Agent{}, err
	}

	new, err := rs.Service.UpdateAgent(id, body)
	if err != nil {
		return &models.Agent{}, err
	}

	return new, nil
}

func (rs Resources) deleteAgent(c *fuego.ContextNoBody) (*models.Agent, error) {
	id := c.PathParam("id")
	return rs.Service.DeleteAgent(id)
}