package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"script_validation/components"
	"sort"
	"strconv"

	"github.com/a-h/templ"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humafiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/sashabaranov/go-openai"
)

// Options for the CLI.
type Options struct {
	Port int `help:"Port to listen on" short:"p" default:"8888"`
}

// GreetingInput represents the greeting operation request.
type GreetingInput struct {
	Name string `path:"name" maxLength:"30" example:"world" doc:"Name to greet"`
}

// GreetingOutput represents the greeting operation response.
type GreetingOutput struct {
	Body struct {
		Message string `json:"message" example:"Hello, world!" doc:"Greeting message"`
	}
}

func addRoutes(api huma.API) {
	// Register GET /greeting/{name}
	huma.Register(api, huma.Operation{
		OperationID: "post-script-chat",
		Summary:     "Post a script chat",
		Method:      http.MethodPost,
		Path:        "/script-chat",
	}, PostScriptChat)
}

type TestResultsWebPage struct {
	Test string
}

type Messages struct {
	Choice  openai.ChatCompletionMessage `json:"choice"`
	Results []results                    `json:"results"`
}

func GetResultsList(c *fiber.Ctx) error {
	filename := c.FormValue("filename")
	testCountStr := c.FormValue("test-count")
	testCount, err := strconv.Atoi(testCountStr)
	if err != nil {
		return err
	}

	form, err := c.MultipartForm()
	if err != nil {
		return err
	}

	models := form.Value["models"]
	if len(models) == 0 {
		return fmt.Errorf("no models provided")
	}

	payload, err := getPayloadFromYAML("./scripts/" + filename)
	if err != nil {
		return err
	}
	fmt.Println(models)

	resp, err := PostScriptChatValidation(context.Background(), payload, testCount, models)
	if err != nil {
		return err
	}

	messages := make([]Messages, len(resp.Body.Messages))

	for i, msg := range resp.Body.Messages {
		// Loads in the full list of messages
		messages[i] = Messages{Choice: msg}
	}

	// Adds the results to the list of messages
	for i, _ := range resp.Body.Messages {
		for _, query := range resp.Body.Results {
			if query.Query.MessageIndex == i {
				messages[i+1].Results = append(messages[i+1].Results, query.Results...)
			}
		}

		if i+1 < len(messages) {
			// Sorts the results by similarity in descending order
			sort.Slice(messages[i+1].Results, func(j, k int) bool {
				return messages[i+1].Results[j].Similarity > messages[i+1].Results[k].Similarity
			})
		}
	}

	fmt.Println("Recieved the results, now rendering the results page")
	return c.Render("components/TestResults", messages[1:])
}

func Render(c *fiber.Ctx, component templ.Component, options ...func(*templ.ComponentHandler)) error {
	componentHandler := templ.Handler(component)
	for _, o := range options {
		o(componentHandler)
	}
	return adaptor.HTTPHandler(componentHandler)(c)
}

func runDevServer() {
	cmd := exec.Command("/bin/sh", "-c", "./templ generate --watch --proxy=\"http://localhost:3504\" --cmd=\"go run . -server & ./tailwindcss -i ./public/input.css -o ./public/output.css --watch\"")
	err := cmd.Start()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	log.Printf("Waiting for command to finish...")
	err = cmd.Wait()
	log.Printf("Command finished with error: %v", err)
}

func main() {
	// Define CLI arguments
	startServer := flag.Bool("server", false, "Start the web server")
	migrateDB := flag.Bool("client", false, "Run database migrations")
	ChatValidation := flag.Bool("validation", false, "Validate the chat")
	devServer := flag.Bool("dev", false, "Start the dev server")

	// Parse CLI arguments
	flag.Parse()

	// Check what action to perform based on the CLI arguments
	if *startServer {
		app := fiber.New()
		app.Use(logger.New())
		app.Static("/", "./public")

		app.Get("/", func(c *fiber.Ctx) error {
			models := []string{
				"openchat/openchat-7b",
				"undi95/toppy-m-7b",
				"gryphe/mythomax-l2-13b",
				"nousresearch/nous-hermes-llama2-13b",
			}
			return Render(c, components.Home(models))
		})

		app.Post("/post-script", GetResultsList)

		// API Routes with validation
		api := humafiber.New(app, huma.DefaultConfig("My API", "1.0"))
		addRoutes(api)

		port := 3510
		for range 5 {
			portString := fmt.Sprintf(":%d", port)
			err := app.Listen(portString)
			if err != nil {
				fmt.Println("Port " + portString + " is already in use. Trying the next port.")
				port++
			} else {
				break
			}
		}
	} else if *migrateDB {
		// Run database migrations
		PostScriptClient()
	} else if *ChatValidation {
		// Run database migrations
		PostScriptValidationClient()
	} else if *devServer {
		runDevServer()
	} else {
		fmt.Println("No valid command provided. Use -start-server to start the server or -migrate-db to run database migrations.")
	}
}
