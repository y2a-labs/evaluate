package main

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"script_validation/llm"
	"strings"
	"sync"
	"time"

	"github.com/sashabaranov/go-openai"
)

type Task struct {
	Intro     string `json:"intro" yaml:"intro" validate:"required"`
	Goal      string `json:"goal" yaml:"goal" validate:"required"`
	Condition string `json:"condition" yaml:"condition" validate:"required"`
}

type Script struct {
	SystemPromptString string `json:"system_prompt_string" yaml:"system_prompt_string" validate:"required"`
	Character          string `json:"character" yaml:"character" validate:"required"`
	Task               []Task `json:"tasks" yaml:"tasks" validate:"required"`
}

type ScriptChatInput struct {
	Body ScriptChatInputBody
}

type ScriptChatInputBody struct {
	Script    Script                         `json:"script" yaml:"script" validate:"required"`
	Messages  []openai.ChatCompletionMessage `json:"messages" yaml:"messages"`
	TaskIndex int                            `json:"script_index" yaml:"script_index" validate:"required"`
	Model     string                         `json:"model" yaml:"model" validate:"required"`
}

type ScriptChatOutput struct {
	Body ScriptChatOutputBody
}

type ScriptChatOutputBody struct {
	Messages  []openai.ChatCompletionMessage `json:"messages"`
	TaskIndex int
}

type ScriptChatValidationOutput struct {
	Body BodyResults `json:"body"`
}

type BodyResults struct {
	Messages []openai.ChatCompletionMessage `json:"messages"`
	Results  []queryResult                  `json:"results"`
}

type results struct {
	Model       string  `json:"model"`
	LLMResponse string  `json:"llm_response"`
	Similarity  float64 `json:"similarity"`
}

type queryResult struct {
	Query   query     `json:"query"`
	Results []results `json:"results"`
}

type query struct {
	MessageIndex    int                            `json:"message_index"`
	Messages        []openai.ChatCompletionMessage `json:"messages"`
	UserResponse    string                         `json:"user_response"`
	CorrectResponse string                         `json:"correct_response"`
}

func parseSystemMessage(script *Script, task Task) (string, error) {
	systemPrompt := script.SystemPromptString

	replacements := map[string]string{
		"{goal}":      task.Goal,
		"{intro}":     task.Intro,
		"{condition}": task.Condition,
		"{character}": script.Character,
	}

	// Check for unknown placeholders
	re := regexp.MustCompile(`\{.*?\}`)
	matches := re.FindAllString(systemPrompt, -1)
	for _, match := range matches {
		if _, ok := replacements[match]; !ok {
			return "", fmt.Errorf("unknown variable %s in the system prompt", match)
		}
	}

	// Replace known placeholders
	for placeholder, value := range replacements {
		systemPrompt = strings.Replace(systemPrompt, placeholder, value, -1)
	}

	return systemPrompt, nil
}

func setSystemMessage(messages []openai.ChatCompletionMessage, systemMessage string) []openai.ChatCompletionMessage {
	// Checks if the first message is the system message
	if messages[0].Role == "system" {
		messages[0].Content = systemMessage
	} else {
		// Sets the system message as the first message
		messages = append([]openai.ChatCompletionMessage{{Role: "system", Content: systemMessage}}, messages...)
	}

	return messages
}

func PostScriptChat(ctx context.Context, input *ScriptChatInput) (*ScriptChatOutput, error) {
	// Parse the system message
	task := input.Body.Script.Task[input.Body.TaskIndex]
	systemMessage, err := parseSystemMessage(&input.Body.Script, task)
	if err != nil {
		return &ScriptChatOutput{}, err
	}

	// Sets the messages and exits if there are no messages
	if len(input.Body.Messages) == 0 {
		messages := []openai.ChatCompletionMessage{
			{Role: "system", Content: systemMessage},
			{Role: "assistant", Content: task.Intro},
		}

		return &ScriptChatOutput{
			Body: ScriptChatOutputBody{
				Messages:  messages,
				TaskIndex: input.Body.TaskIndex,
			},
		}, nil
	} else {
		input.Body.Messages = setSystemMessage(input.Body.Messages, systemMessage)

		// Check to see if the last message is from the assistant
		if input.Body.Messages[len(input.Body.Messages)-1].Role == "assistant" {
			return &ScriptChatOutput{
				Body: ScriptChatOutputBody{
					Messages:  input.Body.Messages,
					TaskIndex: input.Body.TaskIndex,
				},
			}, nil
		}
	}

	// Create the LLM Client
	llm_client := llm.GetLLMClient(map[string]string{})

	message, err := llm.GetLLMResponse(llm_client, input.Body.Messages, input.Body.Model)
	if err != nil {
		return &ScriptChatOutput{}, err
	}

	input.Body.Messages = append(input.Body.Messages, message)

	result := &ScriptChatOutput{
		Body: ScriptChatOutputBody{
			Messages:  input.Body.Messages,
			TaskIndex: 0,
		},
	}

	return result, nil
}

func PostScriptChatValidation(ctx context.Context, input *ScriptChatInput, testCount int, models []string) (*ScriptChatValidationOutput, error) {
	// Parse the system message
	task := input.Body.Script.Task[input.Body.TaskIndex]
	systemMessage, err := parseSystemMessage(&input.Body.Script, task)
	if err != nil {
		return &ScriptChatValidationOutput{}, err
	}

	// Sets the messages and exits if there are no messages
	if len(input.Body.Messages) == 0 {
		return &ScriptChatValidationOutput{}, fmt.Errorf("no messages provided")
	} else {
		input.Body.Messages = setSystemMessage(input.Body.Messages, systemMessage)
	}

	// Create the LLM Client
	llm_client := llm.GetLLMClient(map[string]string{})

	// Get the number of messages where the role is user
	userMessageCount := 0
	for _, message := range input.Body.Messages {
		if message.Role == "user" {
			userMessageCount++
		}
	}

	var wg sync.WaitGroup
	routineCount := userMessageCount * len(models) * testCount
	errCh := make(chan error, routineCount)
	resultsCh := make(chan queryResult, routineCount)

	// Range over the models and test count
	for _, model := range models {
		model := model
		for range testCount {

			for i, message := range input.Body.Messages {
				if message.Role != "user" {
					continue
				}

				wg.Add(1)
				go func(i int, userMessage openai.ChatCompletionMessage) {
					defer wg.Done()

					// Add a random delay before sending the request
					delay := time.Duration(rand.Intn(1000)) * time.Millisecond
					time.Sleep(delay)

					// Get the LLMs Response
					llm_response, err := llm.GetLLMResponse(llm_client, input.Body.Messages[:i+1], model)
					if err != nil {
						errCh <- err
						return
					}

					similarity, err := llm.GetTextSimilarity(llm_response.Content, input.Body.Messages[i+1].Content)
					if err != nil {
						errCh <- err
						return
					}
					result := queryResult{
						Query: query{
							MessageIndex:    i,
							Messages:        input.Body.Messages,
							UserResponse:    userMessage.Content,
							CorrectResponse: input.Body.Messages[i+1].Content,
						},
						Results: []results{
							{
								Model:       model,
								LLMResponse: llm_response.Content,
								Similarity:  similarity,
							},
						},
					}
					resultsCh <- result

				}(i, message)
			}
		}
	}

	wg.Wait()
	close(errCh)
	close(resultsCh)

	for err := range errCh {
		if err != nil {
			return &ScriptChatValidationOutput{}, err
		}
	}

	results := make([]queryResult, 0)

	for result := range resultsCh {

		// Groups the results by the user message
		var found bool
		for i, r := range results {
			if r.Query.UserResponse == result.Query.UserResponse {
				results[i].Results = append(results[i].Results, result.Results...)
				found = true
				break
			}
		}
		if !found {
			results = append(results, result)
		}
	}
	return &ScriptChatValidationOutput{
		Body: BodyResults{
			Results:  results,
			Messages: input.Body.Messages,
		},
	}, nil
}
