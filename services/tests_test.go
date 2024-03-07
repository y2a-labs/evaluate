package service

import (
	"context"
	"script_validation/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEvalPrompts(t *testing.T) {
	messages := []models.Message{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there"},
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there"},
	}
	evalIndexes, err := getTestIndexes(messages)
	assert.Nil(t, err, "Expect no error")
	assert.Equal(t, 2, len(evalIndexes), "Expect two prompts")
	assert.Equal(t, []int{2, 4}, evalIndexes, "Expect the correct indexes")

}

func TestPrepareData(t *testing.T) {
	s := New("../test.db", "../.env")
	data, err := s.prepareTestData(ExecuteTestInput{
		Context: context.Background(),
		PromptID: "d2c0aeca-31f9-4c7d-a713-4e29f0119daf",
		RunCount: 5,
		ConversationID: "b21008a8-9286-4487-ac69-1f610664a3ba",
	})
	assert.Nil(t, err)
	assert.Greater(t, len(data.Conversation.Messages), 0)
	assert.Greater(t, len(data.LLMs), 0)
	assert.NotEmpty(t, data.Prompt.Content)
}

func TestRunTest(t *testing.T) {
	s := New("../test.db", "../.env")
	conversation := &models.Conversation{
		Messages: []models.Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi there"},
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi there"},
		},
	}
	prompt := &models.Prompt{
		Content: "You are a helpful assistant",
	}

	llms := []*models.LLM{
		&models.LLM{
			BaseModel: models.BaseModel{
				ID: "openchat/openchat-7b",
			},
			ProviderID: "openrouter",
		},
	}

	resultChan, totalResultCount, err := s.runTest(&RunTestInput{
		Context: context.Background(),
		Conversation: conversation,
		Prompt:       prompt,
		RunCount:     2,
		LLMs:         llms,
	})
	for result := range resultChan {
		assert.NotEqual(t, "", result.Message.Content)
		assert.NotNil(t, result.Message.Metadata.Embedding)
	}
	
	assert.Nil(t, err, "No errors when running the test")
	assert.Equal(t, 4, totalResultCount)
}