package main

import (
	"fmt"
	database "script_validation"
	apiroutes "script_validation/api/routes"
	webroutes "script_validation/web/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	app := fiber.New()

	app.Use(logger.New())

	database.ConnectDB()

	app.Static("/", "./public")
	apiroutes.SetupAPIRoutes(app)
	webroutes.SetupWebRoutes(app)

	fmt.Println("Server is running on port 8080")
	err := app.Listen(":3000")
	if err != nil {
		fmt.Println("Error starting server: ", err)
	}
}
