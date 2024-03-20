// Generated controller.go.tmpl
package web

import (
	"github.com/y2a-labs/evaluate/models"

	"github.com/go-fuego/fuego"
)

func (rs Resources) RegisterLLMRoutes(s *fuego.Server) {
	LLMGroup := fuego.Group(s, "/models")

	fuego.Get(LLMGroup, "", rs.getLLMs)
	fuego.Post(LLMGroup, "", rs.createLLM)
	fuego.Delete(LLMGroup, "", rs.deleteLLM2)

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
	return c.Render("partials/llms-select.partials.html", llms)
}

func (rs Resources) createLLM(c *fuego.ContextWithBody[models.LLMCreate]) (fuego.HTML, error) {
	body, err := c.Body()
	if err != nil {
		return "", err
	}
	llm, err := rs.Service.CreateLLM(body)
	if err != nil {
		return "", err
	}
	return fuego.HTML("<option value='" + llm.ID + "'>" + llm.ID + "</option>"), nil
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

type ID struct {
	ID string `form:"id"`
}

func (rs Resources) deleteLLM2(c *fuego.ContextWithBody[ID]) (fuego.HTML, error) {
	body, err := c.Body()
	if err != nil {
		return "", err
	}
	_, err = rs.Service.DeleteLLM(body.ID)
	if err != nil {
		return "", err
	}
	return "", nil
}
