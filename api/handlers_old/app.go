package apihandlers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"script_validation/limiter"
	"script_validation/models"
	"script_validation/web/pages"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
)

type App struct {
	Db        *gorm.DB
	aesKey    string
	limiter   *limiter.RateLimiterManager
	validator *validator.Validate
	Router    *fiber.App
	Pages     *pages.Page
}

// In your apihandlers package, adjust the NewApp function to encapsulate all initializations.
func NewApp(dbPath string, envPath string) (*App, error) {
	// Load or create AES key
	aesKey, err := loadOrCreateAESKey(envPath) // This should be adjusted to be a method of App or a standalone function as needed
	if err != nil {
		return nil, fmt.Errorf("error with aeskey," + err.Error())
	}

	// Initialize database connection
	db := connectDB(dbPath) // Ensure this is correctly implemented to handle errors
	db.Debug()

	router := fiber.New()
	router.Use(fiberlogger.New())

	// Initialize your app struct with all components
	app := &App{
		Db:        db,
		aesKey:    aesKey,
		limiter:   limiter.NewRateLimiterManager(),
		validator: validator.New(),
		Router:    router,
		Pages:     &pages.Page{},
	}

	// Additional initializations (e.g., loading providers) can be done here
	app.InitProviders()

	return app, nil
}

func (app *App) EnableDevMode() {
	app.Router.Use(func(c *fiber.Ctx) error {
		c.Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
		c.Set("Pragma", "no-cache")                                   // HTTP 1.0.
		c.Set("Expires", "0")                                         // Proxies.
		return c.Next()
	})
}

func (app *App) Start() {
	for port := 3000; port <= 3100; port++ {
		err := app.Router.Listen(fmt.Sprintf(":%d", port))
		if err != nil {
			if strings.Contains(err.Error(), "address already in use") {
				log.Printf("Port %d already in use, trying next one...", port)
				continue
			} else {
				log.Fatalf("Failed to start server: %v", err)
			}
		} else {
			log.Printf("Server started on port %d", port)
			break
		}
	}
}

func (app *App) ValidateStruct(ctx *fiber.Ctx, i interface{}) error {
	err := ctx.ParamsParser(i)
	if err != nil {
		return err
	}

	err = ctx.BodyParser(i)
	if err != nil {
		return err
	}

	//SetDefaultValues(i)

	err = app.validator.Struct(i)

	if err != nil {
		return err
	}

	return nil
}

func (app *App) InitProviders() {
	providers := []models.Provider{
		{
			ID:       "groq",
			BaseUrl:  "https://api.groq.com/openai/v1",
			Requests: 10,
			Interval: 1,
			Unit:     "minute",
		},
		{
			ID:       "openrouter",
			BaseUrl:  "https://openrouter.ai/api/v1",
			Requests: 250,
			Interval: 10,
			Unit:     "second",
		},
	}

	for _, provider := range providers {
		if err := app.Db.Where("base_url = ?", provider.BaseUrl).First(&models.Provider{}).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// The provider does not exist in the database, so create it
				app.Db.Create(&provider)
			} else {
				// An error occurred while trying to fetch the provider
				log.Printf("Error checking provider: %v", err)
			}
		}
	}
	models := []models.LLM{
		{ID: "llama2-70b-4096", ProviderID: "groq"},
		{ID: "mixtral-8x7b-32768", ProviderID: "groq"},
		{ID: "mistralai/mixtral-8x7b-instruct", ProviderID: "openrouter"},
		{ID: "openchat/openchat-7b", ProviderID: "openrouter"},
		{ID: "undi95/toppy-m-7b", ProviderID: "openrouter"},
	}
	app.Db.Save(&models)
}

func (app *App) Render(c *fiber.Ctx, component templ.Component, options ...func(*templ.ComponentHandler)) error {
	componentHandler := templ.Handler(component)
	for _, o := range options {
		o(componentHandler)
	}
	return adaptor.HTTPHandler(componentHandler)(c)
}

func connectDB(db_name string) *gorm.DB {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level
			IgnoreRecordNotFoundError: false,         // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,          // Disable color
		},
	)
	db, err := gorm.Open(sqlite.Open(db_name), &gorm.Config{
		NowFunc: func() time.Time {
			ti, _ := time.LoadLocation("UTC")
			return time.Now().In(ti)
		},
		Logger: newLogger,
	})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(
		&models.Conversation{},
		&models.Message{},
		&models.MessageEvaluation{},
		&models.MessageEvaluationResult{},
		&models.LLM{},
		&models.MessageMetadata{},
		&models.Provider{},
		&models.Agent{},
	)
	db.Set("gorm:time_zone", "UTC")
	return db
}

func (app *App) RenderTempl(pageFunc func(*fiber.Ctx) templ.Component) fiber.Handler {
	return func(c *fiber.Ctx) error {
		page := pageFunc(c)
		return app.Render(c, page)
	}
}

func (app *App) RenderReflectTempl(pageFunc reflect.Value, page *pages.Page) fiber.Handler {
	return func(c *fiber.Ctx) error {
		results := pageFunc.Call([]reflect.Value{reflect.ValueOf(page), reflect.ValueOf(c)})
		page := results[0].Interface().(templ.Component)
		return app.Render(c, pages.ComponentLayout(page))
	}
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
		return "", fmt.Errorf("error loading .env file")
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
func (app *App) Encrypt(plainText string) (string, error) {
	key, err := base64.StdEncoding.DecodeString(app.aesKey)
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
func (app *App) Decrypt(encryptedText string) (string, error) {
	key, err := base64.StdEncoding.DecodeString(app.aesKey)
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

func connectDB(db_name string) *gorm.DB {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level
			IgnoreRecordNotFoundError: false,         // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,          // Disable color
		},
	)
	db, err := gorm.Open(sqlite.Open(db_name), &gorm.Config{
		NowFunc: func() time.Time {
			ti, _ := time.LoadLocation("UTC")
			return time.Now().In(ti)
		},
		Logger: newLogger,
	})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(
		&models.Conversation{},
		&models.Message{},
		&models.MessageEvaluation{},
		&models.MessageEvaluationResult{},
		&models.LLM{},
		&models.MessageMetadata{},
		&models.Provider{},
		&models.Agent{},
	)
	db.Set("gorm:time_zone", "UTC")
	return db
}
