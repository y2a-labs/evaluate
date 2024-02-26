package apihandlers

import (
	database "script_validation"
	"script_validation/models"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestFindOrCreateLLMs(t *testing.T) {
	godotenv.Load()
	database.ConnectDB(":memory:")
	// Load the desired time zone``

	// Use the loaded location when creating time values
	llms := []models.LLM{
		{
			ID: "model1",
		},
	}
	database.DB.Create(&llms)

	results, err := FindOrCreateLLMs([]string{"model1", "model2"})
	assert.Nil(t, err, "Should not return an error")
	assert.Equal(t, 2, len(results), "Should return 2 results")
	assert.Equal(t, results[0], llms[0], "Should return the first model")
	assert.NotNil(t, results[1], "Should return the second model")
}