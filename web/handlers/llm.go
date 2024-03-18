// Generated controller.go.tmpl
package web

import (
	"github.com/go-fuego/fuego"
	"script_validation/models"
)

func (rs Resources) RegisterLLMRoutes(s *fuego.Server) {
	LLMGroup := fuego.Group(s, "/llms")

	fuego.Get(LLMGroup, "", rs.getLLMs)
	fuego.Post(LLMGroup, "", rs.createLLM)

	fuego.Get(LLMGroup, "/{id}", rs.getLLM)
	fuego.Put(LLMGroup, "/{id}", rs.updateLLM)
	fuego.Delete(LLMGroup, "/{id}", rs.deleteLLM)
}

func (rs Resources) getLLMs(c fuego.ContextNoBody) (fuego.HTML, error) {
	provider := c.QueryParam("provider")
	llms, err := rs.Service.GetLLMByProvider(provider)
	if err != nil {
		return "", err
	}
	return c.Render("partials/llms-select.partials.html",llms)
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