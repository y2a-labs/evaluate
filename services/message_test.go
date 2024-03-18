package service

import (
	"script_validation/models"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

func TestUpdateMessage(t *testing.T) {
	s := New("../test.db", "../.env")
	conversation, err := s.CreateConversation(models.ConversationCreate{
		Name: "hello",
		Messages: []openai.ChatCompletionMessage{
			{
				Role: "assistant",
				Content: "Hello?",
			},
		},
	})
	assert.NoError(t, err, "No error creating conversation")
	assert.Equal(t, 0, conversation.Version, "Expect version to be 0")

	message, err := s.UpdateMessage(conversation.Messages[0].ID, models.MessageUpdate{
		Content: "This is a new world!",
	})
	assert.NoError(t, err, "No error updating the message")
	assert.Equal(t, 1, message.ConversationVersion, "Expect message conversation version to be 1")
	assert.Equal(t, "This is a new world!", message.Content, "Expect content to update")
}
