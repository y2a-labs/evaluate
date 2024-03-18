package service

import (
	"script_validation/models"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

func TestAppendMessages(t *testing.T) {
	s := New("../test.db", "../.env")
	// Create the conversation
	conversation := &models.Conversation{
		BaseModel: models.BaseModel{ID: "hello"},
		LastMessageIndex: 1,
		Version: 1,
		Messages: []*models.Message{
			{
				BaseModel: models.BaseModel{ID: "hello3"},
				Role: "user",
				Content: "hello",
				MessageIndex: 0,
				ConversationVersion: 0,
			},
			{
				BaseModel: models.BaseModel{ID: "hello2"},
				Role: "assistant",
				Content: "hello",
				MessageIndex: 1,
				ConversationVersion: 1,
			},
		},
	}

	messages := []models.ChatCompletionMessage{
		{
			Role: "assistant",
			Content: "Hello world",
		},
		{
			Role: "user",
			Content: "Thanks!",
		},
	}
	_, err := s.AddMessagesToConversation(conversation, messages)
	assert.Nil(t, err, "No error")
	assert.Equal(t, 4, len(conversation.Messages), "has a total of two messages")
	assert.Equal(t, 2, conversation.Version, "Expect the conversation version to increment by 1")
	assert.Equal(t, 3, conversation.LastMessageIndex, "Expect the last message index to be 3")
	assert.Equal(t, conversation.Version, conversation.Messages[2].ConversationVersion, "Expect the new message convo version to match the conversation version")

}


func TestCreateConversation(t *testing.T) {
	s := New(":memory:", "../.env")
	// Create the conversation
	createConversation := &models.ConversationCreate{
		Name: "Hello",
		Description: "World",
		IsTest: true,
		Messages: []openai.ChatCompletionMessage{
			{
				Role: "assistant",
				Content: "Hello",
			},
			{
				Role: "user",
				Content: "hello",
			},
		},
	}
	conversation, err := s.CreateConversation(*createConversation)
	assert.Nil(t, err, "Expect there to be no error")
	assert.Equal(t, 0, conversation.Version, "Expect version to be 0")
	assert.Equal(t, true, conversation.IsTest)
	assert.Equal(t, "Hello", conversation.Name, "Expect name to be set")
    assert.Equal(t, len(createConversation.Messages), len(conversation.Messages), "Expect the number of messages to match")
    for i, msg := range conversation.Messages {
        assert.Equal(t, createConversation.Messages[i].Content, msg.Content, "Expect message content to match")
        assert.Equal(t, createConversation.Messages[i].Role, msg.Role, "Expect message roles to match")
    }
}