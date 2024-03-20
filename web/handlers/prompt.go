// Generated controller.go.tmpl
package web

import (
	"github.com/go-fuego/fuego"
	"github.com/y2a-labs/evaluate/models"
)

func (rs Resources) RegisterPromptRoutes(s *fuego.Server) {
	PromptGroup := fuego.Group(s, "/prompt")

	fuego.Get(PromptGroup, "/", rs.getAllPrompts)
	fuego.Post(PromptGroup, "/", rs.createPrompt)

	fuego.Get(PromptGroup, "/{id}", rs.getPrompt)
	fuego.Put(PromptGroup, "/{id}", rs.updatePrompt)
	fuego.Delete(PromptGroup, "/{id}", rs.deletePrompt)
}

func (rs Resources) getAllPrompts(c fuego.ContextNoBody) (*[]models.Prompt, error) {
	return rs.Service.GetAllPrompts()
}

func (rs Resources) createPrompt(c *fuego.ContextWithBody[models.PromptCreate]) (*models.Prompt, error) {
	body, err := c.Body()
	if err != nil {
		return &models.Prompt{}, err
	}

	return rs.Service.CreatePrompt(body)
}

func (rs Resources) getPrompt(c fuego.ContextNoBody) (*models.Prompt, error) {
	id := c.PathParam("id")

	return rs.Service.GetPrompt(id)
}

func (rs Resources) updatePrompt(c *fuego.ContextWithBody[models.PromptUpdate]) (*models.Prompt, error) {
	id := c.PathParam("id")

	body, err := c.Body()
	if err != nil {
		return &models.Prompt{}, err
	}

	new, err := rs.Service.UpdatePrompt(id, body)
	if err != nil {
		return &models.Prompt{}, err
	}

	return new, nil
}

func (rs Resources) deletePrompt(c *fuego.ContextNoBody) (*models.Prompt, error) {
	id := c.PathParam("id")
	return rs.Service.DeletePrompt(id)
}