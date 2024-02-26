package main

import (
	"log"
	database "script_validation"
	apihandlers "script_validation/api/handlers"
	"script_validation/models"
	valid "script_validation/validator"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func InitProviders() {
	providers := []models.Provider{
		{
			ID:       "groq",
			BaseUrl:  "https://api.groq.com/openai/v1",
			Requests: 10,
			Interval: 1,
			Unit:     "minute",
			EnvKey:   "GROQ_API_KEY",
		},
		{
			ID:       "openrouter",
			BaseUrl:  "https://openrouter.ai/api/v1",
			Requests: 250,
			Interval: 10,
			Unit:     "second",
			EnvKey:   "OPENROUTER_API_KEY",
		},
	}

	for _, provider := range providers {
		if err := database.DB.Where("base_url = ?", provider.BaseUrl).First(&models.Provider{}).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// The provider does not exist in the database, so create it
				database.DB.Create(&provider)
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
	database.DB.Save(&models)
}

func main() {
	app := fiber.New()
	godotenv.Load()
	app.Use(logger.New())
	database.ConnectDB("test.db")
	InitProviders()
	valid.InitValidator()

	// Routes
	app.Get("/conversations", apihandlers.GetConversationListAPI)
	app.Post("/conversations", apihandlers.CreateConversationAPI)

	app.Get("/conversations/:id", apihandlers.GetConversationAPI)
	app.Post("/conversations/:id/evaluate", apihandlers.PostEvaluationAPI)

	app.Static("/static", "./public")
	log.Fatal(app.Listen(":3000"))
}
