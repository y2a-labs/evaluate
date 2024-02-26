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
	conversation, err := GetConversationWithEval(id, 50)
	if err != nil {
		return web.Render(c, components.ErrorMessage(err))
	}

	return web.Render(c, pages.ConversationPage(conversation))
}

func GetConversationWithEval(id string, limit int) (*models.Conversation, error) {
	conversation := &models.Conversation{}

	r := database.DB.
		Preload("Messages.MessageEvaluations", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC").Order("average_similarity DESC").
				Preload("MessageEvaluationResults", func(db *gorm.DB) *gorm.DB {
					return db.Limit(limit) // Limit to 5 results per evaluation
				}).
				Preload("LLM").
				Preload("Prompt")
		}).
		First(&conversation, "conversations.id = ?", id)

	if r.Error != nil {
		return nil, r.Error
	}

	return conversation, nil
}
