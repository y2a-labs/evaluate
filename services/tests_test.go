package service

import (
	"context"
	"script_validation/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEvalPrompts(t *testing.T) {
	messages := []*models.Message{
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

func TestGetTestWithResults(t *testing.T) {
	s := New(":memory:", "../.env")
	assert.NotNil(t, s)
	promptID := "prompt"
	prompt := &models.Prompt{BaseModel: models.BaseModel{ID: "prompt"}}
	s.Db.Create(prompt)

	conversation := &models.Conversation{
		BaseModel: models.BaseModel{ID: "conversation"},
	}
	s.Db.Create(conversation)

	messages := []*models.Message{
		&models.Message{
			BaseModel:     models.BaseModel{ID: "message"},
			MessageIndex: 1,
			ConversationID: "conversation",
			Content:       "Hello",
			Role: "assistant",
			TestMessageID: "",
		},
		&models.Message{
			BaseModel:     models.BaseModel{ID: "message1"},
			MessageIndex: 0,
			ConversationID: "conversation",
			Content:       "Hello",
			Role: "assistant",
			TestMessageID: "",
			Metadata: &models.MessageMetadata{
				BaseModel: models.BaseModel{ID: "metadata2"},
				Embedding: []float64{1, 2, 3},
			},
		},
		&models.Message{
			BaseModel:     models.BaseModel{ID: "testmessage"},
			Content:       "test",
			PromptID:      promptID,
			TestMessageID: "message",
			Metadata: &models.MessageMetadata{
				BaseModel: models.BaseModel{ID: "metadata"},
				StartLatencyMs: 100,
			},
		},
	}
	s.Db.Save(messages)

	test, err := s.GetTestResults("conversation", "prompt")
	assert.Nil(t, err, "expect no error")
	assert.Equal(t, 0, test.Messages[0].MessageIndex, "expect the first message to have an index of 0")
	assert.Equal(t, 2, len(test.Messages), "expect there to be two messages in the conversation")
	assert.Equal(t, 1, len(test.Messages[1].TestMessages), "expect there to be one test result on the second message")
	assert.Equal(t, 100, test.Messages[1].TestMessages[0].Metadata.StartLatencyMs, "expect the test to have metadata")
	assert.Equal(t, 3, len(test.Messages[0].Metadata.Embedding), "expect the first message to have an embedding")
}

func TestPrepareData(t *testing.T) {
	s := New("../test.db", "../.env")
	data, err := s.prepareTestData(ExecuteTestInput{
		Context:        context.Background(),
		PromptID:       "d2c0aeca-31f9-4c7d-a713-4e29f0119daf",
		RunCount:       5,
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
		Messages: []*models.Message{
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
		Context:      context.Background(),
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
