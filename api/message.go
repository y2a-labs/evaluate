// Generated controller.go.tmpl
package api

import (
	"github.com/go-fuego/fuego"
	"script_validation/models"
)

func (rs Resources) RegisterMessageRoutes(s *fuego.Server) {
	MessageGroup := fuego.Group(s, "/message")

	fuego.Get(MessageGroup, "/", rs.getAllMessages)
	fuego.Post(MessageGroup, "/", rs.createMessage)

	fuego.Get(MessageGroup, "/{id}", rs.getMessage)
	fuego.Put(MessageGroup, "/{id}", rs.updateMessage)
	fuego.Delete(MessageGroup, "/{id}", rs.deleteMessage)
}

func (rs Resources) getAllMessages(c fuego.ContextNoBody) (*[]models.Message, error) {
	return rs.Service.GetAllMessages()
}

func (rs Resources) createMessage(c *fuego.ContextWithBody[models.MessageCreate]) (*models.Message, error) {
	body, err := c.Body()
	if err != nil {
		return &models.Message{}, err
	}

	return rs.Service.CreateMessage(body)
}

func (rs Resources) getMessage(c fuego.ContextNoBody) (*models.Message, error) {
	id := c.PathParam("id")

	return rs.Service.GetMessage(id)
}

func (rs Resources) updateMessage(c *fuego.ContextWithBody[models.MessageUpdate]) (*models.Message, error) {
	id := c.PathParam("id")

	body, err := c.Body()
	if err != nil {
		return &models.Message{}, err
	}

	new, err := rs.Service.UpdateMessage(id, body)
	if err != nil {
		return &models.Message{}, err
	}

	return new, nil
}

func (rs Resources) deleteMessage(c *fuego.ContextNoBody) (*models.Message, error) {
	id := c.PathParam("id")
	return rs.Service.DeleteMessage(id)
}