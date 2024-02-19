package routes

import (
	"net/http"
	apihandlers "script_validation/api/handlers"

	huma "github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humafiber"
	"github.com/gofiber/fiber/v2"
)

func SetupAPIRoutes(app *fiber.App) {
	api := humafiber.New(app, huma.DefaultConfig("My API", "1.0"))

	huma.Register(api, huma.Operation{
		OperationID: "post-chat",
		Summary:     "Post a chat",
		Method:      http.MethodPost,
		Path:        "/v1/api/script-chat",
	}, apihandlers.PostScriptChat)

	huma.Register(api, huma.Operation{
		OperationID: "evaluate-conversation",
		Summary:     "Evaluate a Conversation",
		Method:      http.MethodPost,
		Path:        "/v1/api/conversations/{id}/evaluate",
	}, apihandlers.PostScriptChatValidation)

	huma.Register(api, huma.Operation{
		OperationID: "create-conversation",
		Summary:     "Create a Conversation",
		Method:      http.MethodPost,
		Path:        "/v1/api/conversations",
	}, apihandlers.CreateConversation)
}
