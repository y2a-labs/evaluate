// Generated controller.go.tmpl
package web

import (
	"github.com/go-fuego/fuego"
	"script_validation/models"
)

func (rs Resources) RegisterProviderRoutes(s *fuego.Server) {
	ProviderGroup := fuego.Group(s, "/provider")

	fuego.Get(ProviderGroup, "/", rs.getAllProviders)
	fuego.Post(ProviderGroup, "/", rs.createProvider)

	fuego.Get(ProviderGroup, "/{id}", rs.getProvider)
	fuego.Put(ProviderGroup, "/{id}", rs.updateProvider)
	fuego.Delete(ProviderGroup, "/{id}", rs.deleteProvider)
}

func (rs Resources) getAllProviders(c fuego.ContextNoBody) (*[]models.Provider, error) {
	return rs.Service.GetAllProviders()
}

func (rs Resources) createProvider(c *fuego.ContextWithBody[models.ProviderCreate]) (*models.Provider, error) {
	body, err := c.Body()
	if err != nil {
		return &models.Provider{}, err
	}

	return rs.Service.CreateProvider(body)
}

func (rs Resources) getProvider(c fuego.ContextNoBody) (*models.Provider, error) {
	id := c.PathParam("id")

	return rs.Service.GetProvider(id)
}

func (rs Resources) updateProvider(c *fuego.ContextWithBody[models.ProviderUpdate]) (*models.Provider, error) {
	id := c.PathParam("id")

	body, err := c.Body()
	if err != nil {
		return &models.Provider{}, err
	}

	new, err := rs.Service.UpdateProvider(id, body)
	if err != nil {
		return &models.Provider{}, err
	}

	return new, nil
}

func (rs Resources) deleteProvider(c *fuego.ContextNoBody) (*models.Provider, error) {
	id := c.PathParam("id")
	return rs.Service.DeleteProvider(id)
}