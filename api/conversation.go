// Generated controller.go.tmpl
package api

import (
	"github.com/go-fuego/fuego"
	"script_validation/models"
)

func (rs Resources) RegisterConversationRoutes(s *fuego.Server) {
	ConversationGroup := fuego.Group(s, "/conversation")

	fuego.Get(ConversationGroup, "/", rs.getAllConversations)
	fuego.Post(ConversationGroup, "/", rs.createConversation)

	fuego.Get(ConversationGroup, "/{id}", rs.getConversation)
	fuego.Put(ConversationGroup, "/{id}", rs.updateConversation)
	fuego.Delete(ConversationGroup, "/{id}", rs.deleteConversation)
}

func (rs Resources) getAllConversations(c fuego.ContextNoBody) (*[]models.Conversation, error) {
	return rs.Service.GetAllConversations()
}

func (rs Resources) createConversation(c *fuego.ContextWithBody[models.ConversationCreate]) (*models.Conversation, error) {
	body, err := c.Body()
	if err != nil {
		return &models.Conversation{}, err
	}

	return rs.Service.CreateConversation(body)
}

func (rs Resources) getConversation(c fuego.ContextNoBody) (*models.Conversation, error) {
	id := c.PathParam("id")

	return rs.Service.GetConversation(id)
}

func (rs Resources) updateConversation(c *fuego.ContextWithBody[models.ConversationUpdate]) (*models.Conversation, error) {
	id := c.PathParam("id")

	body, err := c.Body()
	if err != nil {
		return &models.Conversation{}, err
	}

	new, err := rs.Service.UpdateConversation(id, body)
	if err != nil {
		return &models.Conversation{}, err
	}

	return new, nil
}

func (rs Resources) deleteConversation(c *fuego.ContextNoBody) (*models.Conversation, error) {
	id := c.PathParam("id")
	return rs.Service.DeleteConversation(id)
}