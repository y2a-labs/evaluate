package apihandlers

/*
func TestGetConversation(t *testing.T) {
	db := database.ConnectDB(":memory:")
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
	db.Create(want)
	have, err := GetConversation(want.ID)
	assert.Nil(t, err)
	assert.Equal(t, have, want)
}
*/