package webroutes

import (
	webhandlers "script_validation/web/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupWebRoutes(app *fiber.App) {
	app.Get("/conversations/:id", webhandlers.ConversationsHandler)

	app.Get("/conversations", webhandlers.GetConversationListHandler)

	app.Post("/conversations/:id/evaluate", webhandlers.EvaluateHandler)
}
