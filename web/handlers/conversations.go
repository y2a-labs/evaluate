package webhandlers

import (
	"fmt"
	database "script_validation"
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

	// Get the conversation
	conversation := &models.Conversation{}

	r := database.DB.
		Preload("Messages.MessageEvaluations", func(db *gorm.DB) *gorm.DB {
			return db.Where("average_similarity > 0").Order("created_at DESC").Preload("MessageEvaluationResults").Preload("LLM")
		}).
		First(&conversation, "conversations.id = ?", id)

	if r.Error != nil {
		return web.Render(c, components.ErrorMessage(r.Error))
	}

	return web.Render(c, pages.ConversationPage(conversation))
}

func GetConversationListHandler(c *fiber.Ctx) error {
	var conversations []models.Conversation

	r := database.DB.
		Limit(15).
		Find(&conversations)

	if r.Error != nil {
		return web.Render(c, components.ErrorMessage(r.Error))
	}

	return web.Render(c, pages.ConversationListPage(conversations))
}