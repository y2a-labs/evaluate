// Generated controller.go.tmpl
package web

import (
	"script_validation/models"

	"github.com/go-fuego/fuego"
)

func (rs Resources) RegisterProviderRoutes(s *fuego.Server) {
	ProviderGroup := fuego.Group(s, "/providers")

	fuego.Get(ProviderGroup, "", rs.getAllProviders)
	fuego.Post(ProviderGroup, "", rs.createProvider)

	fuego.Get(ProviderGroup, "/{id}", rs.getProvider)
	fuego.Put(ProviderGroup, "/{id}", rs.updateProvider)
	fuego.Delete(ProviderGroup, "/{id}", rs.deleteProvider)
	fuego.Post(ProviderGroup, "/{id}/models", rs.pullLLMsFromProvider)
}

func (rs Resources) getAllProviders(c fuego.ContextNoBody) (fuego.HTML, error) {
	providers, err := rs.Service.GetAllProviders()
	if err != nil {
		return "", err
	}
	for i, _ := range providers {
		llms, err := rs.Service.GetLLMByProvider(providers[i].ID)
		if err != nil {
			return "", err
		}
		providers[i].Models = llms
	}

	return c.Render("pages/providers.page.html", providers)
}

func (rs Resources) pullLLMsFromProvider(c fuego.ContextNoBody) (fuego.HTML, error) {
	id := c.PathParam("id")
	models, err := rs.Service.PullLLMsFromProvider(id)
	if err != nil {
		return "", err
	}

	return c.Render("partials/models.partials.html", models)
}

func (rs Resources) createProvider(c *fuego.ContextWithBody[models.ProviderCreate]) (fuego.HTML, error) {
	body, err := c.Body()
	if err != nil {
		return "", err
	}
	provider, err := rs.Service.CreateProvider(body)
	if err != nil {
		return c.Render("partials/error.partials.html", err.Error())
	}

	return c.Render("partials/provider.partials.html", provider) 
}

func (rs Resources) getProvider(c fuego.ContextNoBody) (*models.Provider, error) {
	id := c.PathParam("id")

	return rs.Service.GetProvider(id)
}

func (rs Resources) updateProvider(c *fuego.ContextWithBody[models.ProviderUpdate]) (fuego.HTML, error) {
	id := c.PathParam("id")

	body, err := c.Body()
	if err != nil {
		return "", err
	}

	new, err := rs.Service.UpdateProvider(id, body)
	if err != nil {
		return c.Render("partials/error.partials.html", err.Error())
	}

	return c.Render("partials/provider.partials.html", new)
}

func (rs Resources) deleteProvider(c *fuego.ContextNoBody) (fuego.HTML, error) {
	id := c.PathParam("id")
	_, err := rs.Service.DeleteProvider(id)
	if err != nil {
		return "", err
	}
	return "", err
}
