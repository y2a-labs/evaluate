// Generated controller.go.tmpl
package web

import (
	"script_validation/models"

	"github.com/go-fuego/fuego"
)

func (rs Resources) RegisterMessageRoutes(s *fuego.Server) {
	MessageGroup := fuego.Group(s, "/messages")

	fuego.Get(MessageGroup, "", rs.getAllMessages)
	fuego.Post(MessageGroup, "", rs.createMessage)

	fuego.Get(MessageGroup, "/{id}", rs.getMessage)
	fuego.Get(MessageGroup, "/{id}/edit", rs.getEditMessage)
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

func (rs Resources) getMessage(c fuego.ContextNoBody) (fuego.HTML, error) {
	id := c.PathParam("id")
	msg, err := rs.Service.GetMessage(id)
	if err != nil {
		return "nil", err
	}
	return c.Render("partials/message.partials.html", msg)
}

func (rs Resources) getEditMessage(c fuego.ContextNoBody) (fuego.HTML, error) {
	id := c.PathParam("id")
	msg, err := rs.Service.GetMessage(id)
	if err != nil {
		return "nil", err
	}
	return c.Render("partials/message-edit.partials.html", msg)
}

func (rs Resources) updateMessage(c *fuego.ContextWithBody[models.MessageUpdate]) (fuego.HTML, error) {
	id := c.PathParam("id")

	body, err := c.Body()
	if err != nil {
		return "", err
	}

	message, err := rs.Service.UpdateMessage(id, body)
	if err != nil {
		return "", err
	}

	return c.Render("partials/message.partials.html", message)
}

func (rs Resources) deleteMessage(c *fuego.ContextNoBody) (any, error) {
	id := c.PathParam("id")
	return nil, rs.Service.DeleteMessage(id)
}