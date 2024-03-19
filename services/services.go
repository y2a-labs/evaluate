package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"script_validation/internal/limiter"
	"script_validation/models"
	"time"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Service struct {
	Db           *gorm.DB
	limiter      *limiter.RateLimiterManager
	llmProviders map[string]*llmProvider
}

type llmProvider struct {
	*models.Provider
	client *openai.Client
}

func (s *Service) GetLLMProviderNames() []string {
	names := make([]string, 0, len(s.llmProviders))
	for k := range s.llmProviders {
		names = append(names, k)
	}
	return names
}

func New(dbPath, envPath string) *Service {
	aesKey, err := loadOrCreateAESKey(envPath)
	if err != nil {
		panic(fmt.Errorf("error with aeskey," + err.Error()))
	}
	// Initialize database connection
	db := connectDB(dbPath)
	db.Debug()

	rateLimiter := limiter.NewRateLimiterManager()

	llmProviders := getOpenaiComatibleProviders(db, aesKey)

	setRateLimits(llmProviders, rateLimiter)

	return &Service{
		Db:           db,
		limiter:      rateLimiter,
		llmProviders: llmProviders,
	}
}

func setRateLimits(llmProviders map[string]*llmProvider, rateLimiter *limiter.RateLimiterManager) {
	for _, provider := range llmProviders {
		rateLimiter.GetLimiter(provider.Provider)
	}
	return
}

func getOpenaiComatibleProviders(db *gorm.DB, aesKey string) map[string]*llmProvider {
	llmProviders := make(map[string]*llmProvider)

	// Get the list of providers
	providers := []models.Provider{}
	tx := db.Where("type = ?", "llm").Find(&providers)
	if tx.Error != nil {
		log.Printf("error getting embedding providers: %v", tx.Error)
	}

	for _, provider := range providers {
		// Get the client
		decryptedKey, err := decrypt(provider.EncryptedAPIKey, aesKey)
		if err != nil {
			log.Printf("error decrypting api key for provider %s: %v", provider.ID, err)
			continue
		}
		client := openai.NewClient(decryptedKey, provider.BaseUrl)

		llmProviders[provider.ID] = &llmProvider{
			Provider: &provider,
			client:   client,
		}
	}
	return llmProviders
}

func connectDB(db_name string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(db_name), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now() // Use local timezone
		},
	})

	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(
		&models.Conversation{},
		&models.Message{},
		&models.LLM{},
		&models.MessageMetadata{},
		&models.Provider{},
	)

	// Creates the inital list of providers on the first run
	var count int64
	db.Model(&models.Provider{}).Count(&count)
	if count == 0 {
		// Seed the database
		seedDB(db)
	}

	return db
}

func seedDB(db *gorm.DB) error {
	providers := []models.Provider{
		{
			BaseModel: models.BaseModel{ID: "openai"},
			Type:      "llm",
			BaseUrl:   "https://api.openai.com/v1",
			Requests: 10,
			Interval: 1,
			Unit: "seconds",
		},
		{
			BaseModel: models.BaseModel{ID: "openrouter"},
			Type:      "llm",
			BaseUrl:   "https://openrouter.ai/api/v1",
			Requests: 250,
			Interval: 10,
			Unit: "seconds",
		},
		{
			BaseModel: models.BaseModel{ID: "local"},
			Type:      "llm",
			BaseUrl:   "http://localhost:8080/v1",
			Requests: 250,
			Interval: 10,
			Unit: "seconds",
		},
	}

	tx := db.Create(&providers)
	if tx.Error != nil {
		panic("failed to seed database")
	}
	return nil
}

// generateRandomBytes generates a random byte slice of n bytes.
func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

// generateAESKey generates a new AES-256 key.
func generateAESKey() (string, error) {
	key, err := generateRandomBytes(32) // 32 bytes for AES-256
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// loadOrCreateAESKey checks for an AES key in the .env file, or creates one if not present.
func loadOrCreateAESKey(envPath string) (string, error) {
	// Load environment variables from .env file
	err := godotenv.Load(envPath)
	if err != nil {
		// If the .env file doesn't exist, create it
		if os.IsNotExist(err) {
			file, err := os.Create(envPath)
			if err != nil {
				return "", fmt.Errorf("error creating .env file: %v", err)
			}
			file.Close()
		} else {
			return "", fmt.Errorf("error loading .env file: %v", err)
		}
	}

	aesKey := os.Getenv("AES_KEY")
	if aesKey == "" {
		// Generate a new AES key if not found
		newKey, err := generateAESKey()
		if err != nil {
			return "", fmt.Errorf("error generating aes key: %v", err)
		}

		envContent := fmt.Sprintf("AES_KEY=%s\n", newKey)

		file, err := os.OpenFile(".env", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return "", err
		}
		defer file.Close()

		_, err = io.WriteString(file, envContent)
		if err != nil {
			return "", err
		}

		return newKey, nil
	}

	return aesKey, nil
}

// Encrypt encrypts plaintext using AES-GCM with the AES key stored in the App struct.
func Encrypt(plainText, aesKey string) (string, error) {
	key, err := base64.StdEncoding.DecodeString(aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode aesKey: %v", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %v", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %v", err)
	}

	encrypted := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// Decrypt decrypts ciphertext using AES-GCM with the AES key stored in the App struct.
func decrypt(encryptedText, aesKey string) (string, error) {
	key, err := base64.StdEncoding.DecodeString(aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode aesKey: %v", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %v", err)
	}

	decodedMsg, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", fmt.Errorf("failed to decode encryptedText: %v", err)
	}

	if len(decodedMsg) < gcm.NonceSize() {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := decodedMsg[:gcm.NonceSize()], decodedMsg[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %v", err)
	}

	return string(plaintext), nil
}
