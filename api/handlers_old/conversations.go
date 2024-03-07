package apihandlers

import (
	"errors"
	"fmt"
	"os"
	"script_validation/internal/nomicai"
	"script_validation/models"
	"script_validation/web/pages"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/sashabaranov/go-openai"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func (app *App) SetMessageEmbeddings(messages *[]models.Message, embeddingClient *nomicai.Client) {
	texts := make([]string, len(*messages))
	for i, message := range *messages {
		texts[i] = message.Content
	}
	embeddings, err := embeddingClient.EmbedText(texts, nomicai.Clustering)
	if err != nil {
		fmt.Println("Error embedding text: ", err)
		return
	}
	for i, embedding := range embeddings.Embeddings {
		(*messages)[i].Metadata.Embedding = datatypes.NewJSONSlice(embedding)
	}
}

func (app *App) ParseConversationString(conversationString string) ([]openai.ChatCompletionMessage, error) {
	var messages []openai.ChatCompletionMessage
	lines := strings.Split(conversationString, "\n")

	for i, line := range lines {
		if line == "" {
			continue // Skip empty lines
		}

		var parts []string
		if strings.HasPrefix(line, "ASSISTANT:") {
			parts = []string{"ASSISTANT", line[3:]} // Skip "ASSISTANT:" and split
		} else if strings.HasPrefix(line, "USER:") {
			parts = []string{"USER", line[5:]} // Skip "USER:" and split
		} else {
			return nil, errors.New("invalid speaker prefix in line " + strconv.Itoa(i+1) + ": " + line)
		}

		role := ""
		switch parts[0] {
		case "ASSISTANT":
			role = "assistant"
		case "USER":
			role = "user"
		default:
			return nil, errors.New("unknown speaker role in line " + strconv.Itoa(i+1) + ": " + line)
		}

		// Trim space here to ensure there are no leading/trailing spaces in the content
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: strings.TrimSpace(parts[1]),
		})
	}

	return messages, nil
}

func (app *App) CreateConversation(name string, text string) (*models.Conversation, error) {
	// Make sure the last message in the conversation is from the assistant
	chatMessages, err := app.ParseConversationString(text)
	if err != nil {
		return nil, err
	}

	// Create the conversation
	conversation := models.Conversation{Name: name, EvalEndIndex: len(chatMessages) - 1}

	result := app.Db.Create(&conversation)

	if result.Error != nil {
		return nil, result.Error
	}

	// Create the messages
	messages := make([]models.Message, len(chatMessages))

	for i, message := range chatMessages {
		messages[i] = models.Message{
			ConversationID: conversation.ID,
			Content:        message.Content,
			Role:           message.Role,
			MessageIndex:   uint(i),
		}
	}
	embeddingClient := nomicai.NewClient(os.Getenv("NOMICAI_API_KEY"))
	app.SetMessageEmbeddings(&messages, embeddingClient)

	result = app.Db.Create(&messages)

	if result.Error != nil {
		return nil, result.Error
	}
	conversation.Messages = messages

	return &conversation, nil
}

// app.RenderTempl(app.Pages.ConversationPage)
func (app *App) GetConversation(id string) (*models.Conversation, error) {
	// Get the conversation
	conversation := models.Conversation{BaseModel: models.BaseModel{ID: id}}

	result := app.Db.
		Preload("Messages.MessageEvaluations", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC"). //.Limit(5). // Limit to 5 MessageEvaluations
								Preload("MessageEvaluationResults", func(db *gorm.DB) *gorm.DB {
					//return db.Limit(5) // Limit to 5 results per evaluation
					return db
				}).
				Preload("LLM.Provider")
		}).
		First(&conversation, "conversations.id = ?", id)

	if result.Error != nil {
		return nil, result.Error
	}

	// Return the conversation
	return &conversation, nil
}

func (app *App) GetConversationAPI(ctx *fiber.Ctx) error {

	id := ctx.Params("id")

	conversation, err := app.GetConversation(id)
	if err != nil {
		return err
	}
	// If the headers asked for json
	if ctx.Get("Accept") == "application/json" {
		return ctx.JSON(conversation)
	} else {
		return Render(ctx, app.Pages.ConversationPage(conversation))
	}
}

func (app *App) TestAPI(ctx *fiber.Ctx) (*models.Conversation, error) {
	return &models.Conversation{
		Name: ctx.Params("id"),
		Messages: []models.Message{
			{
				Role:         "user",
				Content:      "hello",
				MessageIndex: 0,
			},
		},
	}, nil
}

func (app *App) GetConversationList(limit int) (*[]models.Conversation, error) {
	// Get the conversation
	conversation := []models.Conversation{}

	result := app.Db.Limit(limit).Order("created_at DESC").Find(&conversation)

	if result.Error != nil {
		return nil, result.Error
	}

	// Return the conversation
	return &conversation, nil
}

func (app *App) GetConversationListAPI(ctx *fiber.Ctx) error {
	// Get the conversation
	limit := ctx.Query("limit", "10")
	// turn the limit into an int
	limitInt, _ := strconv.Atoi(limit)
	conversations, err := app.GetConversationList(limitInt)
	if err != nil {
		return err
	}

	return Render(ctx, pages.ConversationListPage(conversations))
}

func (app *App) CreateConversationAPI(ctx *fiber.Ctx) error {

	name := ctx.FormValue("name")
	text := ctx.FormValue("conversation")
	conversation, err := app.CreateConversation(name, text)
	if err != nil {
		return err
	}
	return Render(ctx, pages.ConversationCard(*conversation))
}

func Render(c *fiber.Ctx, component templ.Component, options ...func(*templ.ComponentHandler)) error {
	componentHandler := templ.Handler(component)
	for _, o := range options {
		o(componentHandler)
	}
	return adaptor.HTTPHandler(componentHandler)(c)
}
