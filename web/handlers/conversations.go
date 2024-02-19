package webhandlers

import (
	"fmt"
	"script_validation/views/components"
	"script_validation/views/pages"
	"script_validation/web"

	"github.com/gofiber/fiber/v2"
)

func ConversationsHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" || len(id) != 15 {
		return web.Render(c, components.ErrorMessage(fmt.Errorf("invalid conversation id: %s", id)))
	}

	return web.Render(c, pages.ConversationsPage())
}
