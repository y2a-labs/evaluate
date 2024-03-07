// Generated controller.go.tmpl
package web

import (
	"github.com/go-fuego/fuego"
	"script_validation/models"
)

func (rs Resources) RegisterLLMRoutes(s *fuego.Server) {
	LLMGroup := fuego.Group(s, "/lLM")

	fuego.Get(LLMGroup, "/", rs.getAllLLMs)
	fuego.Post(LLMGroup, "/", rs.createLLM)

	fuego.Get(LLMGroup, "/{id}", rs.getLLM)
	fuego.Put(LLMGroup, "/{id}", rs.updateLLM)
	fuego.Delete(LLMGroup, "/{id}", rs.deleteLLM)
}

func (rs Resources) getAllLLMs(c fuego.ContextNoBody) (*[]models.LLM, error) {
	return rs.Service.GetAllLLMs()
}

func (rs Resources) createLLM(c *fuego.ContextWithBody[models.LLMCreate]) (*models.LLM, error) {
	body, err := c.Body()
	if err != nil {
		return &models.LLM{}, err
	}

	return rs.Service.CreateLLM(body)
}

func (rs Resources) getLLM(c fuego.ContextNoBody) (*models.LLM, error) {
	id := c.PathParam("id")

	return rs.Service.GetLLM(id)
}

func (rs Resources) updateLLM(c *fuego.ContextWithBody[models.LLMUpdate]) (*models.LLM, error) {
	id := c.PathParam("id")

	body, err := c.Body()
	if err != nil {
		return &models.LLM{}, err
	}

	new, err := rs.Service.UpdateLLM(id, body)
	if err != nil {
		return &models.LLM{}, err
	}

	return new, nil
}

func (rs Resources) deleteLLM(c *fuego.ContextNoBody) (*models.LLM, error) {
	id := c.PathParam("id")
	return rs.Service.DeleteLLM(id)
}