package webhandlers

import (
	"context"
	"fmt"
	"io"
	"os"
	apihandlers "script_validation/api/handlers"
	"script_validation/views/components"
	"script_validation/web"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gopkg.in/yaml.v2"
)

func GetPayloadFromYAML(filename string) (*apihandlers.ScriptChatInput, error) {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read the file
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Unmarshal the YAML data into the Body field of a ScriptChatInput struct
	payload := &apihandlers.ScriptChatInput{}
	err = yaml.Unmarshal(data, &payload.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return payload, nil
}

func GetResultsList(c *fiber.Ctx) error {
	filename := c.FormValue("filename")
	testCountStr := c.FormValue("test-count")
	testCount, err := strconv.Atoi(testCountStr)
	if err != nil {
		return web.Render(c, components.ErrorMessage(err))
	}

	form, err := c.MultipartForm()
	if err != nil {
		return web.Render(c, components.ErrorMessage(err))
	}

	models := form.Value["models"]
	if len(models) == 0 {
		return web.Render(c, components.ErrorMessage(fmt.Errorf("no models provided")))
	}

	payload, err := GetPayloadFromYAML("./scripts/" + filename)
	if err != nil {
		return web.Render(c, components.ErrorMessage(err))
	}

	payload.Body.Models = models
	payload.Body.TestCount = testCount

	resp, err := apihandlers.PostScriptChatValidation(context.Background(), payload)
	if err != nil {
		return web.Render(c, components.ErrorMessage(err))
	}

	fmt.Println("Recieved the results, now rendering the results page")
	return web.Render(c, components.ResultMessage(resp.Body.Messages))
}
