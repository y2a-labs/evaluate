package webroutes

import (
	"script_validation/views/pages"
	"script_validation/web"
	webhandlers "script_validation/web/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupWebRoutes(app *fiber.App) {
	// Register GET /greeting/{name}
	app.Get("/", func(c *fiber.Ctx) error {
		return web.Render(c, pages.Home())
	})

	app.Get("/conversations/:id", webhandlers.ConversationsHandler)

	app.Get("/conversations", webhandlers.GetConversationsHandler)
}
