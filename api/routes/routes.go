package routes

import (
	"net/http"
	"reflect"
	apihandlers "script_validation/api/handlers"
	"script_validation/models"

	huma "github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humafiber"
	"github.com/gofiber/fiber/v2"
)

type Test struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

func NewResponse(config *huma.Config, i interface{}) *huma.Response {
	schema := huma.SchemaFromType(config.Components.Schemas, reflect.TypeOf(i))
	return &huma.Response{
		Content: map[string]*huma.MediaType{
			"application/json": {
				Schema: schema,
			},
		},
	}
}

func NewRequestBody(config *huma.Config, i interface{}) *huma.RequestBody {
	schema := huma.SchemaFromType(config.Components.Schemas, reflect.TypeOf(i))
	return &huma.RequestBody{
		Content: map[string]*huma.MediaType{
			"application/json": {
				Schema: schema,
			},
		},
	}
}

func SetupAPIRoutes(app *fiber.App) {

	config := huma.DefaultConfig("My API", "1.0")

	api := humafiber.New(app, config)

	huma.Register(api, huma.Operation{
		OperationID: "create-evaluation",
		Summary:     "Create LLM Evaluation",
		Method:      http.MethodPost,
		Path:        "/v1/api/conversations/{id}/evaluate",
		Responses: map[string]*huma.Response{
			"200": NewResponse(&config, models.CreateLLMEvaluationResponseValidation{}),
		},
	}, apihandlers.CreateLLMEvaluationAPI)

	huma.Register(api, huma.Operation{
		OperationID: "create-conversation",
		Summary:     "Create a Conversation",
		Method:      http.MethodPost,
		Path:        "/v1/api/conversations",
	}, apihandlers.CreateConversationAPI)

	huma.Register(api, huma.Operation{
		OperationID: "get-conversation",
		Summary:     "Get a Conversation",
		Method:      http.MethodGet,
		Path:        "/v1/api/conversations/{id}",
		Responses: map[string]*huma.Response{
			"200": NewResponse(&config, models.GetConversationResponseValidation{}),
		},
	}, apihandlers.GetConversationAPI)

	huma.Register(api, huma.Operation{
		OperationID: "find-or-create-prompt",
		Summary:     "Find or Create a Prompt",
		Method:      http.MethodPost,
		Path:        "/v1/api/prompts",
	}, apihandlers.FindOrCreatePromptAPI)

	huma.Register(api, huma.Operation{
		OperationID: "get-prompt-by-id",
		Summary:     "Get Prompt by ID",
		Method:      http.MethodPost,
		Path:        "/v1/api/prompts/{id}",
	}, apihandlers.GetPromptByIdAPI)
}
