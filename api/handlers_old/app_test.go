package apihandlers

import (
	"os"
	"script_validation/models"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncyryption(t *testing.T) {
	//Making sure the app works
	app, err := NewApp(":memory:", "../../.env")
	assert.Nil(t, err, "expected there to be no error initalizing app")
	assert.NotEmpty(t, app.aesKey, "Expect there to be an aes key")

	//Making sure encryption works
	dummyKey := "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"
	encryptedKey, err := app.Encrypt(dummyKey)
	assert.Nil(t, err, "expect no error when encryingt the key")
	assert.NotEmpty(t, encryptedKey, "Key to be encrypted")
	t.Log("API Key", dummyKey)
	t.Log("AES KEy", app.aesKey)
	t.Log("Encrypted Key:", encryptedKey)

	//Making sure decryption works
	decryptedKey, err := app.Decrypt(encryptedKey)
	assert.Nil(t, err, "expect no error when decrypting the key")
	assert.Equal(t, dummyKey, decryptedKey, "Expected the decrypted key to match the original key")
}

func Test2(t *testing.T) {
	app, err := NewApp("../../test.db", "../../.env")
	assert.Nil(t, err, "expected there to be no error initalizing app")
	assert.NotEmpty(t, app.aesKey, "Expect there to be an aes key")

	keys := []string{"GROQ_API_KEY", "OPENROUTER_API_KEY", "TOGETHER_API_KEY", "OPENAI_API_KEY", "NOMICAI_API_KEY"}
	t.Log("AES KEy", app.aesKey)
	for _, keyName := range keys {
		key := os.Getenv(keyName)
		encryptedKey, err := app.Encrypt(key)
		t.Log("API Key", key)
		t.Log("Encrypted Key:", encryptedKey)
		assert.Nil(t, err, "Problem encryping the key")
		id := strings.ToLower(strings.Split(keyName, "_")[0])
		t.Log("id", id)
		roq := &models.Provider{ID: id}
		tx := app.Db.First(roq)
		assert.Nil(t, tx.Error, "Expect no error in the tx")

		tx = app.Db.Model(roq).Update("encrypted_api_key", encryptedKey)
		assert.Nil(t, tx.Error, "Expect no error in the tx")

		//Making sure decryption works
		decryptedKey, err := app.Decrypt(encryptedKey)
		assert.Nil(t, err, "expect no error when decrypting the key5")
		assert.Equal(t, key, decryptedKey, "Expected the decrypted key to match the original key")
	}
}
