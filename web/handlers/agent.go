// Generated controller.go.tmpl
package web

import (
	"fmt"
	"script_validation/models"

	"github.com/go-fuego/fuego"
)

func (rs Resources) RegisterAgentRoutes(s *fuego.Server) {
	AgentGroup := fuego.Group(s, "/agent")

	fuego.Get(AgentGroup, "/", rs.agentsPage)
	fuego.Post(AgentGroup, "/", rs.createAgent)

	fuego.Get(AgentGroup, "/{id}/tests", rs.agentTestPage)
	fuego.Put(AgentGroup, "/{id}/", rs.updateAgent)
	fuego.Delete(AgentGroup, "/{id}/", rs.deleteAgent)
}

func (rs Resources) agentsPage(c fuego.ContextNoBody) (fuego.HTML, error) {
	agents, err := rs.Service.GetAllAgents()
	if err != nil {
		return "pages.AgentListPage(nil)", err
	}
	return c.Render("pages/agents.page.html", agents)
}

func (rs Resources) agentTestPage(c fuego.ContextNoBody) (fuego.HTML, error) {
	id := c.PathParam("id")
	agent := &models.Agent{BaseModel: models.BaseModel{ID: id}}
	tx := rs.Service.Db.Preload("Prompts").
    Preload("Conversations", "is_test = ?", true).
    First(agent)

	if tx.Error != nil {
		return "", tx.Error
	}
	fmt.Println(agent.Conversations)
	return c.Render("pages/agent-tests.page.html", agent)
}

func (rs Resources) createAgent(c *fuego.ContextWithBody[models.AgentCreate]) (any, error) {
	body, err := c.Body()
	if err != nil {
		return nil, err
	}
	if body.Description == "" {
		body.Description = "Untitled Agent Description"
	}
	if body.Name == "" {
		body.Name = "Untitled Agent"
	}

	agent, err := rs.Service.CreateAgent(body)
	if err != nil {
		return nil, err
	}

	return c.Redirect(303, "/agent/"+agent.ID)
}

func (rs Resources) getAgentWithConversationsAndPrompts(c *fuego.ContextNoBody) (fuego.HTML, error) {
	id := c.PathParam("id")
	agent := &models.Agent{BaseModel: models.BaseModel{ID: id}}
	tx := rs.Service.Db.Preload("Prompts").Preload("Conversations").First(agent)
	if tx.Error != nil {
		return "pages.AgentCard(models.Agent{})", tx.Error
	}
	return "pages.AgentPage(agent)", nil
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
	fmt.Println("Deleting agent")
	return rs.Service.DeleteAgent(id)
}
