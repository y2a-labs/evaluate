// Generated controller.go.tmpl
package api

import (
	"github.com/go-fuego/fuego"
	"github.com/y2a-labs/evaluate/models"
)

func (rs Resources) RegisterMessageMetadataRoutes(s *fuego.Server) {
	MessageMetadataGroup := fuego.Group(s, "/messageMetadata")

	fuego.Get(MessageMetadataGroup, "/", rs.getAllMessageMetadatas)
	fuego.Post(MessageMetadataGroup, "/", rs.createMessageMetadata)

	fuego.Get(MessageMetadataGroup, "/{id}", rs.getMessageMetadata)
	fuego.Put(MessageMetadataGroup, "/{id}", rs.updateMessageMetadata)
	fuego.Delete(MessageMetadataGroup, "/{id}", rs.deleteMessageMetadata)
}

func (rs Resources) getAllMessageMetadatas(c fuego.ContextNoBody) (*[]models.MessageMetadata, error) {
	return rs.Service.GetAllMessageMetadatas()
}

func (rs Resources) createMessageMetadata(c *fuego.ContextWithBody[models.MessageMetadataCreate]) (*models.MessageMetadata, error) {
	body, err := c.Body()
	if err != nil {
		return &models.MessageMetadata{}, err
	}

	return rs.Service.CreateMessageMetadata(body)
}

func (rs Resources) getMessageMetadata(c fuego.ContextNoBody) (*models.MessageMetadata, error) {
	id := c.PathParam("id")

	return rs.Service.GetMessageMetadata(id)
}

func (rs Resources) updateMessageMetadata(c *fuego.ContextWithBody[models.MessageMetadataUpdate]) (*models.MessageMetadata, error) {
	id := c.PathParam("id")

	body, err := c.Body()
	if err != nil {
		return &models.MessageMetadata{}, err
	}

	new, err := rs.Service.UpdateMessageMetadata(id, body)
	if err != nil {
		return &models.MessageMetadata{}, err
	}

	return new, nil
}

func (rs Resources) deleteMessageMetadata(c *fuego.ContextNoBody) (*models.MessageMetadata, error) {
	id := c.PathParam("id")
	return rs.Service.DeleteMessageMetadata(id)
}