package webhandlers

import (
	"context"
	database "script_validation"
	apihandlers "script_validation/api/handlers"
	"script_validation/models"
	"script_validation/views/components"
	"script_validation/views/pages"
	"script_validation/web"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func EvaluateHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	form, err := c.MultipartForm()
	if err != nil {
		web.Render(c, components.ErrorMessage(err))
	}
	prompt := form.Value["prompt"]
	testCount := form.Value["test_count"]
	llms := form.Value["models"]

	testCountInt, err := strconv.Atoi(testCount[0])
	if err != nil {
		// handle error
		return web.Render(c, components.ErrorMessage(err))
	}

	// Get the conversation
	conversation := &models.Conversation{
		BaseModel:   models.BaseModel{ID: id},
		EvalTestCount: testCountInt,
		EvalPrompt:   prompt[0],
		EvalModels:  llms,
	}

	database.DB.Save(&conversation)


	_, err = apihandlers.CreateLLMEvaluation(context.Background(), &models.CreateLLMEvaluationRequest{
		ID: id,
		Body: models.CreateLLMEvaluationRequestBody{
			Prompt:    prompt[0],
			TestCount: testCountInt,
			Models:    llms,
		},
	})

	if err != nil {
		return err
	}

	r := database.DB.
		Preload("Messages.MessageEvaluations", func(db *gorm.DB) *gorm.DB {
			return db.Where("average_similarity > 0").Order("created_at DESC").Limit(5).Preload("MessageEvaluationResults").Preload("LLM")
		}).
		First(&conversation, "conversations.id = ?", id)

	if r.Error != nil {
		return web.Render(c, components.ErrorMessage(r.Error))
	}

	return web.Render(c, pages.Messages(conversation.Messages))
}
