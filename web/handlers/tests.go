// Generated controller.go.tmpl
package web

import (
	"errors"
	"math"
	"script_validation/models"
	service "script_validation/services"
	"sort"

	"github.com/go-fuego/fuego"
)

func (rs Resources) RegisterTestRoutes(s *fuego.Server) {
	TestGroup := fuego.Group(s, "/tests")
	fuego.Post(TestGroup, "", rs.createTest)
	fuego.Delete(TestGroup, "/{id}", rs.deleteTest)
	fuego.Put(TestGroup, "/{id}", rs.updateTest)

	fuego.Put(TestGroup, "/{id}/appendmodel", rs.appendTestModel)
	fuego.Put(TestGroup, "/{id}/removemodel", rs.deleteTestModel)
	fuego.Post(TestGroup, "/{id}/messages", rs.addMessagesToTest)
	fuego.Get(TestGroup, "/{id}", rs.getTest)
	fuego.Post(TestGroup, "/{id}", rs.runTest)
}

type RunTestInput struct {
	RunCount      int    `form:"runCount"`
	TestMessageID string `form:"testMessageID"`
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
		ConversationID: conversationID,
		TestMessageID:  body.TestMessageID,
	})
	if err != nil {
		return nil, err
	}
	return c.Redirect(303, "/tests/"+conversationID)
}

type AddMessages struct {
	Content string `form:"content"`
	Role    string `form:"role"`
}

func (rs Resources) addMessagesToTest(c *fuego.ContextWithBody[AddMessages]) (fuego.HTML, error) {
	id := c.PathParam("id")
	body, err := c.Body()
	if err != nil {
		return "nil", err
	}
	inputMessages := []models.ChatCompletionMessage{
		{Role: body.Role, Content: body.Content},
	}
	conversation, err := rs.Service.GetConversation(id)
	if err != nil {
		return "", err
	}
	messages, err := rs.Service.AddMessagesToConversation(conversation, inputMessages)
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
	return c.Redirect(303, "/tests/"+conversation.ID)
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

func (rs Resources) deleteTestModel(c *fuego.ContextWithBody[ModelInput]) (fuego.HTML, error) {
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

	version := c.QueryParamInt("version", -1)

	conversation, err := rs.Service.GetTest(id, version)
	if err != nil {
		return "", err
	}

	providers := rs.Service.GetLLMProviderNames()

	if len(providers) == 0 {
		return "", errors.New("no providers found")
	}

	llms, err := rs.Service.GetLLMByProvider(providers[0])
	if err != nil {
		return "", err
	}

	versions := make([]int, conversation.Version+1)
	for i := range versions {
		versions[i] = i
	}

	// Get the score for each llm
	scoreSum := map[string]float64{}
	scoreCount := map[string]int{}

	// Gets the sum of the scores
	for _, msg := range conversation.Messages {
		if msg.Role != "assistant" {
			continue
		}
		for _, testMsg := range msg.TestMessages {
			scoreSum[testMsg.LLMID] += testMsg.Score
			scoreCount[testMsg.LLMID]++
		}
	}

	// Puts the new scores into the models slice
	for i, llm := range conversation.TestModels {
		if scoreCount[llm.Model] > 0 {
			averageScore := scoreSum[llm.Model] / float64(scoreCount[llm.Model])
			roundedScore := math.Round(averageScore*100) / 100
			conversation.TestModels[i].Score = roundedScore
		}
	}

	// Reorder the test models by
	// Sort the TestModels slice from highest to lowest score
	sort.Slice(conversation.TestModels, func(i, j int) bool {
		return conversation.TestModels[i].Score > conversation.TestModels[j].Score
	})

	return c.Render("pages/test.page.html", map[string]any{
		"test":         conversation,
		"llmProviders": providers,
		"versions":     versions,
		"llms":         llms,
	})
}

func (rs Resources) updateTest(c *fuego.ContextWithBody[models.ConversationUpdate]) (*models.Conversation, error) {
	id := c.PathParam("id")

	body, err := c.Body()
	if err != nil {
		return &models.Conversation{}, err
	}

	test, err := rs.Service.UpdateConversation(id, body)
	if err != nil {
		return &models.Conversation{}, err
	}

	return test, nil
}

func (rs Resources) deleteTest(c *fuego.ContextNoBody) (*models.Conversation, error) {
	id := c.PathParam("id")
	return rs.Service.DeleteConversation(id)
}
