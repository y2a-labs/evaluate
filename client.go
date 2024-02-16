package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
	"gopkg.in/yaml.v2"
)

func getPayloadFromYAML(filename string) (*ScriptChatInput, error) {
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
	payload := &ScriptChatInput{}
	err = yaml.Unmarshal(data, &payload.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return payload, nil
}

func getUserMessage(messages []openai.ChatCompletionMessage) ([]openai.ChatCompletionMessage, error) {
	if len(messages) > 0 && messages[len(messages)-1].Role == "assistant" {
		fmt.Print("User: ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return []openai.ChatCompletionMessage{}, err
		}
		input = strings.TrimSpace(input) // Remove the newline at the end
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    "user",
			Content: input,
		})
	}
	return messages, nil
}

func PrintPreviousChat(messages []openai.ChatCompletionMessage) {
	for _, message := range messages {
		if message.Role == "assistant" {
			fmt.Println("AI: " + message.Content)
		} else if message.Role == "user" {
			fmt.Println("User: " + message.Content)
		}
	}
}

func PostScriptClient() {
	payload, err := getPayloadFromYAML("./scripts/air_script.yaml")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Prints the last message if its the assistant
	PrintPreviousChat(payload.Body.Messages)

	for {
		// Gets the users response
		payload.Body.Messages, err = getUserMessage(payload.Body.Messages)
		if err != nil {
			fmt.Println(err)
			return
		}

		resp, err := PostScriptChat(context.Background(), payload)
		if err != nil {
			fmt.Println(err)
			return
		}
		payload.Body.Messages = resp.Body.Messages
		// Print the message the user is going to respond to.
		lastIndex := len(resp.Body.Messages) - 1
		message := resp.Body.Messages[lastIndex]

		fmt.Println("AI: " + message.Content)

	}
}

func PostScriptValidationClient() {
	payload, err := getPayloadFromYAML("./scripts/air_script.yaml")
	if err != nil {
		fmt.Println(err)
		return
	}

	testCount := 1
	models := []string{"undi95/toppy-m-7b"}
	_, err = PostScriptChatValidation(context.Background(), payload, testCount, models)
	if err != nil {
		fmt.Println(err)
		return
	}
}
