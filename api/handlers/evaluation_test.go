package apihandlers

import (
	database "script_validation"
	"script_validation/models"
	"testing"
)

func TestCreatePrompt(t *testing.T) {
	// Set up test database
	database.ConnectDB(":memory:")

	prompt, err := FindOrCreatePrompt("You are a helpful assistant")

	if err != nil {
		t.Fatalf("Was not able to create or get the prompt id: %v", err)
	}

	if prompt.ID == "" {
		t.Errorf("Expected id to be 'test', got '%s'", prompt.ID)
	}
}

func TestFindPrompt(t *testing.T) {
	// Set up test database
	database.ConnectDB(":memory:")

	prompt := models.Prompt{Content: "You are a helpful assistant"}

	database.DB.Create(&prompt)

	p, err := FindOrCreatePrompt("You are a helpful assistant")

	if err != nil {
		t.Fatalf("Was not able to create or get the prompt id: %v", err)
	}

	if p.ID != prompt.ID {
		t.Errorf("Expected id to be '%s', got '%s'", prompt.ID, p.ID)
	}
}
