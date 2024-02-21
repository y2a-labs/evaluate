package main

import (
	"fmt"
	"log"
	database "script_validation"
	apiroutes "script_validation/api/routes"
	webroutes "script_validation/web/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	app := fiber.New()

	app.Use(logger.New())

	database.ConnectDB("test.db")

	app.Static("/static", "./public")
	apiroutes.SetupAPIRoutes(app)
	webroutes.SetupWebRoutes(app)

	fmt.Println("Server is running on port 8080")
	err = app.Listen(":3000")
	if err != nil {
		fmt.Println("Error starting server: ", err)
	}
}
