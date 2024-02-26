package apihandlers

import (
	database "script_validation"
	"script_validation/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConversation(t *testing.T) {
	database.ConnectDB(":memory:")
	want := &models.Conversation{
		Name: "test",
		Messages: []models.Message{
			{
				ChatMessage: models.ChatMessage{
					Role:    "user",
					Content: "hello",
				},
				MessageIndex: 0,
			},
		},
	}
	database.DB.Create(want)
	have, err := GetConversation(want.ID)
	assert.Nil(t, err)
	assert.Equal(t, have, want)
}
