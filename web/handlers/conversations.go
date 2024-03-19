// Generated controller.go.tmpl
package web

import (
	"script_validation/models"

	"github.com/go-fuego/fuego"
)

func (rs Resources) RegisterConversationRoutes(s *fuego.Server) {
	ConversationGroup := fuego.Group(s, "/conversations")
	fuego.Post(ConversationGroup, "", rs.createTest)
	fuego.Get(ConversationGroup, "", rs.getConversationList)

	fuego.Get(ConversationGroup, "/{id}", rs.getConversation)
	fuego.Put(ConversationGroup, "/{id}", rs.updateConversation)
	fuego.Delete(ConversationGroup, "/{id}", rs.deleteConversation)
}

func (rs Resources) getConversation(c fuego.ContextNoBody) (fuego.HTML, error) {
	id := c.PathParam("id")
	conversation, err := rs.Service.GetConversationWithMessages(id, -1)
	if err != nil {
		return "", err
	}

	return c.Render("pages/conversation.page.html", conversation)
}

func (rs Resources) getConversationList(c fuego.ContextNoBody) (fuego.HTML, error) {
	conversations, err := rs.Service.GetAllConversations()
	if err != nil {
		return "", err
	}
	for i := range conversations {
		conversations[i].CreatedAtString = conversations[i].CreatedAt.Format("January 2 03:04 PM")
	}
	return c.Render("pages/conversations.page.html", conversations)
}

func (rs Resources) updateConversation(c *fuego.ContextWithBody[models.ConversationUpdate]) (any, error) {
	id := c.PathParam("id")

	body, err := c.Body()
	if err != nil {
		return "", err
	}

	conversation, err := rs.Service.UpdateConversation(id, body)
	if err != nil {
		return "", err
	}
	if body.IsTest {
		return c.Redirect(303, "/tests/"+conversation.ID)
	}

	return conversation, nil
}

func (rs Resources) deleteConversation(c *fuego.ContextNoBody) (*models.Conversation, error) {
	id := c.PathParam("id")
	return rs.Service.DeleteConversation(id)
}
