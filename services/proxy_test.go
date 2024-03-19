package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetModel(t *testing.T) {
	s := New("../test.db", "../.env")
	
	// Model and provider are specified
	modelName := "openrouter/openchat/openchat-7b"
	modelID, providerID, err := s.GetModel(modelName)
	assert.Nil(t, err, "expect to not have an error")
	assert.Equal(t, "openchat/openchat-7b", modelID, "expect to have the correct model ID")
	assert.Equal(t, "openrouter", providerID, "expect to have the correct provider ID")

	// Model only is specified
	modelName = "openchat/openchat-7b"
	modelID, providerID, err = s.GetModel(modelName)
	assert.Nil(t, err)
	assert.Equal(t, "openchat/openchat-7b", modelID)
	assert.Equal(t, "openrouter", providerID)

	// When its not a valid model name
	modelName = "openchat-7b"
	modelID, providerID, err = s.GetModel(modelName)
	assert.Error(t, err)
	assert.Equal(t, "", modelID)
	assert.Equal(t, "", providerID)
}
