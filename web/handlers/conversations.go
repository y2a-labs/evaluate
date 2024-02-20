package webhandlers

import (
	"fmt"
	database "script_validation"
	apihandlers "script_validation/api/handlers"
	"script_validation/models"
	"script_validation/views/components"
	"script_validation/views/pages"
	"script_validation/web"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type ConversationWebHandler struct {
	models.Conversation
	Messages []models.Message
}

func ConversationsHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	if len(id) != 36 {
		return web.Render(c, components.ErrorMessage(fmt.Errorf("invalid conversation id: %s", id)))
	}

	conversation, err := apihandlers.GetConversation(nil, &models.GetConversationInput{Id: id})

	if err != nil {
		return web.Render(c, components.ErrorMessage(err))
	}

	return web.Render(c, pages.ConversationPage(conversation.Body))
}

func GetConversationsHandler(c *fiber.Ctx) error {
	var conversations []models.Conversation

	r := database.DB.Limit(15).Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("message_index ASC")
	}).Find(&conversations)

	if r.Error != nil {
		return web.Render(c, components.ErrorMessage(r.Error))
	}

	return web.Render(c, pages.ConversationsPage(conversations))
}