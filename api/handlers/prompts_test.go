package apihandlers

import (
	"fmt"
	"log"
	database "script_validation"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestFindOrCreatePrompt(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	database.ConnectDB("test.db")

	_, err = FindOrCreatePrompt("")
	assert.NotNil(t, err, "No Prompt should be created with empty content")

	t2, _ := FindOrCreatePrompt("test")
	fmt.Println("Prompt: ", t2.Content)
	assert.Equal(t, t2.Content, "test", "Prompt should have content")

	t3, _ := FindOrCreatePrompt("test")
	assert.Equal(t, t3.ID, t2.ID, "Should've found the previous prompt by ID")
}
