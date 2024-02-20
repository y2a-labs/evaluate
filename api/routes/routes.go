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
		OperationID: "create-evaluation",
		Summary:     "Create LLM Evaluation",
		Method:      http.MethodPost,
		Path:        "/v1/api/conversations/{id}/evaluate",
	}, apihandlers.CreateLLMEvaluation)

	huma.Register(api, huma.Operation{
		OperationID: "create-conversation",
		Summary:     "Create a Conversation",
		Method:      http.MethodPost,
		Path:        "/v1/api/conversations",
	}, apihandlers.CreateConversation)

	huma.Register(api, huma.Operation{
		OperationID: "get-conversation",
		Summary:     "Get a Conversation",
		Method:      http.MethodGet,
		Path:        "/v1/api/conversations/{id}",
	}, apihandlers.GetConversation)
}
