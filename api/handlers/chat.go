package apihandlers

import (
	"context"
	"fmt"
	"regexp"
	"script_validation/internal/llm"
	"strings"

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
	TestCount int                            `json:"test_count" yaml:"test_count"`
	Script    Script                         `json:"script" yaml:"script" validate:"required"`
	Messages  []openai.ChatCompletionMessage `json:"messages" yaml:"messages"`
	TaskIndex int                            `json:"script_index" yaml:"script_index" validate:"required"`
	Models    []string                       `json:"models" yaml:"models" validate:"required"`
}

type ScriptChatOutput struct {
	Body ScriptChatOutputBody
}

type ScriptChatOutputBody struct {
	Messages  []openai.ChatCompletionMessage `json:"messages"`
	TaskIndex int
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

	message, err := llm.GetLLMResponse(llm_client, input.Body.Messages, input.Body.Models[0])
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

type ScriptChatValidationOutput struct {
	Body BodyResults `json:"body"`
}

type BodyResults struct {
	Messages []Message `json:"messages"`
}

type Message struct {
	Content           string                       `json:"content"`
	Role              string                       `json:"role"`
	TaskIndex         int                          `json:"task_index"`
	MessageIndex      int                          `json:"message_index"`
	LLMSimilarityList map[string]LLMSimilarityList `json:"llm_results"`
}

type LLMSimilarityList struct {
	Model             string          `json:"model"`
	AverageSimilarity float64         `json:"average_similarity"`
	LLMResponses      []LLMSimilarity `json:"responses"`
}

type LLMSimilarity struct {
	LLMResponse string  `json:"llm_response"`
	Similarity  float64 `json:"similarity"`
	LatencyMs   int64   `json:"latency_ms"`
}
