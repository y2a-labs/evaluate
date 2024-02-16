package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"script_validation/components"
	"script_validation/handlers"
	"sort"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humafiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func addRoutes(api huma.API) {
	// Register GET /greeting/{name}
	huma.Register(api, huma.Operation{
		OperationID: "post-script-chat",
		Summary:     "Post a script chat",
		Method:      http.MethodPost,
		Path:        "/script-chat",
	}, handlers.PostScriptChat)
}

type TestResultsWebPage struct {
	Test string
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

	fmt.Println(models)

	payload, err := getPayloadFromYAML("./scripts/" + filename)
	if err != nil {
		return err
	}
	fmt.Println(models)

	resp, err := handlers.PostScriptChatValidation(context.Background(), payload, testCount, models)
	if err != nil {
		return err
	}

	messages := make([]handlers.Message, len(resp.Body.Messages))

	for i, msg := range resp.Body.Messages {
		// Loads in the full list of messages
		messages[i] = handlers.Message{Choice: msg}
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
	return Render(c, components.TestResults(messages[1:]))
}

func Render(c *fiber.Ctx, component templ.Component, options ...func(*templ.ComponentHandler)) error {
	componentHandler := templ.Handler(component)
	for _, o := range options {
		o(componentHandler)
	}
	return adaptor.HTTPHandler(componentHandler)(c)
}

func getOpenPort(port int) (string, error) {
	for range 10 {
		portString := fmt.Sprintf("%d", port)
		conn, err := net.DialTimeout("tcp", net.JoinHostPort("localhost", portString), time.Second)

		// If there is an error, loop to the next port
		if err != nil {
			fmt.Println("Port is available:", portString)
			return portString, nil
		}

		// If the connection is successful, close the connection and return the port
		if conn != nil {
			defer conn.Close()
			port++
			continue

		}
	}
	return "", nil
}

func runDevServer() {
	port := 3000
	portString, err := getOpenPort(port) // Assuming getOpenPort returns an available port as a string and error
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	cmd1 := exec.Command("bash", "-c", `./templ generate`)
	cmd1.Stdout = os.Stdout
	cmd1.Stderr = os.Stderr
	err = cmd1.Run()
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("bash", "-c", fmt.Sprintf(`
		./templ generate --watch --proxy="http://localhost:%s" &
		go run . -server -port %s &
		./tailwindcss -i ./public/input.css -o ./public/output.css --watch &
		wait
		`, portString, portString))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()

	if err != nil {
		log.Fatal(err)
	}

	// Create a channel to receive OS signals
	c := make(chan os.Signal, 1)
	// Notify the channel for SIGINT signals
	signal.Notify(c, os.Interrupt)

	// Run a goroutine that will kill the command when an interrupt signal is received
	go func() {
		<-c
		if err := cmd.Process.Kill(); err != nil {
			log.Fatal("Failed to kill process: ", err)
		}
	}()

	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// Define CLI arguments
	startServer := flag.Bool("server", false, "Start the web server")
	migrateDB := flag.Bool("client", false, "Run database migrations")
	ChatValidation := flag.Bool("validation", false, "Validate the chat")
	devServer := flag.Bool("dev", false, "Start the dev server")
	serverPort := flag.String("port", "3000", "Port to start the server on")

	// Parse CLI arguments
	flag.Parse()

	// Check what action to perform based on the CLI arguments
	if *startServer {
		fmt.Println("Starting the server")
		app := fiber.New()
		app.Use(logger.New())
		app.Static("/", "./public")

		app.Get("/", func(c *fiber.Ctx) error {
			return Render(c, components.Home())
		})

		app.Post("/post-script", GetResultsList)

		// API Routes with validation
		api := humafiber.New(app, huma.DefaultConfig("My API", "1.0"))
		addRoutes(api)

		err := app.Listen(":" + *serverPort)
		if err != nil {
			fmt.Println("Error starting the server:", err)
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
