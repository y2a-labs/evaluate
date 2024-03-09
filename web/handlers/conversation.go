// Generated controller.go.tmpl
package web

import (
	"errors"
	"fmt"
	"script_validation/models"
	service "script_validation/services"
	"strconv"
	"strings"

	"github.com/go-fuego/fuego"
	"gorm.io/gorm"
)

func (rs Resources) RegisterConversationRoutes(s *fuego.Server) {
	ConversationGroup := fuego.Group(s, "/tests")
	fuego.Get(ConversationGroup, "/", rs.getAllConversations)
	fuego.Post(ConversationGroup, "/", rs.createTest)
	fuego.Post(ConversationGroup, "/{id}/messages", rs.addMessagesToConversation)
	fuego.Post(ConversationGroup, "/{id}/run", rs.runTest)
	fuego.Get(ConversationGroup, "/{id}/results", rs.getTestResults)

	fuego.Get(ConversationGroup, "/{id}/", rs.getTest)
	fuego.Put(ConversationGroup, "/{id}/", rs.updateConversation)
	fuego.Put(ConversationGroup, "/{id}/appendmodel", rs.appendTestModel)
	fuego.Put(ConversationGroup, "/{id}/removemodel", rs.removeTestModel)
	fuego.Delete(ConversationGroup, "/{id}", rs.deleteConversation)
}

type RunTestInput struct {
	RunCount int    `form:"run_count"`
	PromptID string `form:"prompt_id"`
}


func (rs Resources) getTestResults(c fuego.ContextNoBody) (fuego.HTML, error) {
	conversationID := c.PathParam("id")
	promptID := c.QueryParam("promptID")
	results, err := rs.Service.GetTestResults(conversationID, promptID)
	if err != nil {
		return "", err
	}
	return c.Render("pages/test-results.page.html", results)
}

func (rs Resources) runTest(c *fuego.ContextWithBody[RunTestInput]) (any, error) {
	conversationID := c.PathParam("id")
	body, err := c.Body()
	if err != nil {
		return nil, err
	}
	_, err = rs.Service.ExecuteTestWorkflow(service.ExecuteTestInput{
		Context:        c.Context(),
		RunCount:       body.RunCount,
		PromptID:       body.PromptID,
		ConversationID: conversationID,
	})
	if err != nil {
		return nil, err
	}
	return c.Redirect(303, "/conversation/"+conversationID+"/results/")
}

func (rs Resources) getAllConversations(c fuego.ContextNoBody) (*[]models.Conversation, error) {
	return rs.Service.GetAllConversations()
}

type AddMessages struct {
	Content string `form:"content"`
}

func (rs Resources) addMessagesToConversation(c *fuego.ContextWithBody[AddMessages]) (fuego.HTML, error) {
	id := c.PathParam("id")
	body, err := c.Body()
	if err != nil {
		return "nil", err
	}
	inputMessages, err := parseConversationString(body.Content)
	if err != nil {
		return "nil", err
	}
	fmt.Println("role", inputMessages[0].Role, "Content", inputMessages[0].Content)
	messages, err := rs.Service.AddMessagesToConversation(id, inputMessages)
	if err != nil {
		return "nil", err
	}
	return c.Render("partials/messages.partials.html", messages)
}

func (rs Resources) createTest(c *fuego.ContextWithBody[models.ConversationCreate]) (any, error) {
	body, err := c.Body()
	if err != nil {
		return "", err
	}
	body.IsTest = true
	if body.Description == "" {
		body.Description = "Untitled Description"
	}
	if body.Name == "" {
		body.Name = "Untitled Name"
	}
	if body.PromptID == "" {
		prompt := &models.Prompt{}
		tx := rs.Service.Db.Where("agent_id = ?", body.AgentID).Where("base_prompt_id = ?", "").First(prompt)
		if tx.Error != nil {
			return "", tx.Error
		}
		body.PromptID = prompt.ID
	}
	conversation, err := rs.Service.CreateConversation(body)
	if err != nil {
		return "", err
	}
	return c.Redirect(303, "/conversation/"+conversation.ID)
}

type ModelInput struct {
	Provider string
	Model    string
}

func (rs Resources) appendTestModel(c *fuego.ContextWithBody[ModelInput]) (fuego.HTML, error) {
	id := c.PathParam("id")
	body, err := c.Body()
	if err != nil {
		return "", err
	}
	conversation := &models.Conversation{BaseModel: models.BaseModel{ID: id}}
	// Find the conversation
	tx := rs.Service.Db.Select("test_models").First(conversation)
	if tx.Error != nil {
		return "", tx.Error
	}

	model := models.TestModels{Provider: body.Provider, Model: body.Model}

	// Check if there's any item with the same Provider and Model
	for _, testModel := range conversation.TestModels {
		if testModel.Provider == model.Provider && testModel.Model == model.Model {
			return "", errors.New("a model with the same provider and model name already exists")
		}
	}

	if conversation.TestModels == nil || len(conversation.TestModels) == 0 {
		conversation.TestModels = []models.TestModels{model}
	} else {
		conversation.TestModels = append(conversation.TestModels, model)
	}

	tx = rs.Service.Db.Model(conversation).Update("test_models", conversation.TestModels)
	if tx.Error != nil {
		return "", tx.Error
	}

	return c.Render("partials/selected-model-row.partials.html", model)
}

func (rs Resources) removeTestModel(c *fuego.ContextWithBody[ModelInput]) (fuego.HTML, error) {
	id := c.PathParam("id")
	body, err := c.Body()
	if err != nil {
		return "", err
	}
	conversation := &models.Conversation{BaseModel: models.BaseModel{ID: id}}
	// Find the conversation
	tx := rs.Service.Db.Select("test_models").First(conversation)
	if tx.Error != nil {
		return "", tx.Error
	}

	// Create a new slice excluding the matching item
	newTestModels := []models.TestModels{}
	for _, testModel := range conversation.TestModels {
		if testModel.Provider != body.Provider || testModel.Model != body.Model {
			newTestModels = append(newTestModels, testModel)
		}
	}

	conversation.TestModels = newTestModels

	tx = rs.Service.Db.Model(conversation).Update("test_models", conversation.TestModels)
	if tx.Error != nil {
		return "", tx.Error
	}

	return "", nil
}

func (rs Resources) getTest(c fuego.ContextNoBody) (fuego.HTML, error) {
	id := c.PathParam("id")
	conversation := &models.Conversation{BaseModel: models.BaseModel{ID: id}}
	tx := rs.Service.Db.Preload("Prompt").Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Where("test_message_id = ?", "").Order("message_index asc")
	}).First(conversation)
	if tx.Error != nil {
		return "", tx.Error
	}

	return c.Render("pages/conversation.page.html", conversation)
}

func (rs Resources) updateConversation(c *fuego.ContextWithBody[models.ConversationUpdate]) (*models.Conversation, error) {
	id := c.PathParam("id")

	body, err := c.Body()
	if err != nil {
		return &models.Conversation{}, err
	}

	new, err := rs.Service.UpdateConversation(id, body)
	if err != nil {
		return &models.Conversation{}, err
	}

	return new, nil
}

func (rs Resources) deleteConversation(c *fuego.ContextNoBody) (*models.Conversation, error) {
	id := c.PathParam("id")
	return rs.Service.DeleteConversation(id)
}

func parseConversationString(messagesString string) ([]models.ChatCompletionMessage, error) {
	var messages []models.ChatCompletionMessage
	lines := strings.Split(messagesString, "\n")

	for i, line := range lines {
		if line == "" {
			continue // Skip empty lines
		}

		var parts []string
		if strings.HasPrefix(line, "ASSISTANT: ") {
			parts = []string{"ASSISTANT", line[10:]} // Skip "ASSISTANT:" and split
		} else if strings.HasPrefix(line, "USER: ") {
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
		messages = append(messages, models.ChatCompletionMessage{
			Role:    role,
			Content: strings.TrimSpace(parts[1]),
		})
	}
	if len(messages) == 0 {
		return nil, errors.New("no messages found")
	}
	return messages, nil
}
