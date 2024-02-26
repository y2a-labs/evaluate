package apihandlers

import (
	"errors"
	"fmt"
	"os"
	database "script_validation"
	"script_validation/internal/nomicai"
	"script_validation/models"
	"script_validation/views/pages"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func SetMessageEmbeddings(messages *[]models.Message, embeddingClient *nomicai.Client) {
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
		(*messages)[i].Embedding = datatypes.NewJSONSlice(embedding)
	}
}

func ParseConversationString(conversationString string) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage
	lines := strings.Split(conversationString, "\n")

	for i, line := range lines {
		if line == "" {
			continue // Skip empty lines
		}

		var parts []string
		if strings.HasPrefix(line, "AI:") {
			parts = []string{"AI", line[3:]} // Skip "AI:" and split
		} else if strings.HasPrefix(line, "USER:") {
			parts = []string{"USER", line[5:]} // Skip "USER:" and split
		} else {
			return nil, errors.New("invalid speaker prefix in line " + strconv.Itoa(i+1) + ": " + line)
		}

		role := ""
		switch parts[0] {
		case "AI":
			role = "assistant"
		case "USER":
			role = "user"
		default:
			return nil, errors.New("unknown speaker role in line " + strconv.Itoa(i+1) + ": " + line)
		}

		// Trim space here to ensure there are no leading/trailing spaces in the content
		messages = append(messages, models.ChatMessage{
			Role:    role,
			Content: strings.TrimSpace(parts[1]),
		})
	}

	return messages, nil
}

func CreateConversation(name string, text string) (*models.Conversation, error) {
	// Make sure the last message in the conversation is from the assistant
	chatMessages, err := ParseConversationString(text)
	if err != nil {
		return nil, err
	}

	// Create the conversation
	conversation := models.Conversation{Name: name}

	result := database.DB.Create(&conversation)

	if result.Error != nil {
		return nil, result.Error
	}

	// Create the messages
	messages := make([]models.Message, len(chatMessages))

	for i, message := range chatMessages {
		messages[i] = models.Message{
			ConversationID: conversation.ID,
			ChatMessage:    message,
			MessageIndex:   uint(i),
		}
	}
	embeddingClient := nomicai.NewClient(os.Getenv("NOMICAI_API_KEY"))
	SetMessageEmbeddings(&messages, embeddingClient)

	result = database.DB.Create(&messages)

	if result.Error != nil {
		return nil, result.Error
	}
	conversation.Messages = messages

	return &conversation, nil
}

func GetConversation(id string) (*models.Conversation, error) {
	// Get the conversation
	conversation := models.Conversation{BaseModel: models.BaseModel{ID: id}}

	result := database.DB.
		Preload("Messages.MessageEvaluations", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC").Limit(5). // Limit to 5 MessageEvaluations
				Preload("MessageEvaluationResults", func(db *gorm.DB) *gorm.DB {
					return db.Limit(5) // Limit to 5 results per evaluation
				}).
				Preload("LLM")
		}).
		First(&conversation, "conversations.id = ?", id)

	if result.Error != nil {
		return nil, result.Error
	}

	// Return the conversation
	return &conversation, nil
}

func GetConversationAPI(ctx *fiber.Ctx) error {

	id := ctx.Params("id")
	conversation, err := GetConversation(id)
	if err != nil {
		return err
	}

	return Render(ctx, pages.ConversationPage(conversation))
}

func GetConversationList(limit int) (*[]models.Conversation, error) {
	// Get the conversation
	conversation := []models.Conversation{}

	result := database.DB.Limit(limit).Order("created_at DESC").Find(&conversation)

	if result.Error != nil {
		return nil, result.Error
	}

	// Return the conversation
	return &conversation, nil
}

func GetConversationListAPI(ctx *fiber.Ctx) error {
	// Get the conversation
	limit := ctx.Query("limit", "10")
	// turn the limit into an int
	limitInt, _ := strconv.Atoi(limit)
	conversations, err := GetConversationList(limitInt)
	if err != nil {
		return err
	}

	return Render(ctx, pages.ConversationListPage(conversations))
}

func CreateConversationAPI(ctx *fiber.Ctx) error {

	name := ctx.FormValue("name")
	text := ctx.FormValue("conversation")
	conversation, err := CreateConversation(name, text)
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
