package apihandlers

import (
	"context"
	"fmt"
	database "script_validation"
	"script_validation/internal/llm"
	"script_validation/internal/nomicai"
	"script_validation/models"
	"sync"
	"time"

	"github.com/sashabaranov/go-openai"
)

type LLMEvaluationInput struct {
	ConversationId string `path:"id"`
	Body           LLMEvaluationInputBody
}

type LLMEvaluationInputBody struct {
	Models    []string `json:"models" example:"openchat/openchat-7b"`
	TestCount int      `json:"test_count" example:"1"`
	Prompt    string   `json:"prompt" example:"You are a helpful assistant."`
}

type LLMEvaluationOutput struct {
	Body LLMEvaluationOuputBody
}

type LLMEvaluationOuputBody struct {
	Results []models.MessageEvaluation `json:"results"`
}

type LLMEvaluationOuputResults struct {
	Model       string  `json:"model"`
	LLMResponse string  `json:"llm_response"`
	Similarity  float64 `json:"similarity"`
	LatencyMs   int64   `json:"latency_ms"`
}

type ProcessMessage struct {
	EvaluationRoutine
	TestCount         int
	MessageEvaluation models.MessageEvaluation
	LLMClient         *openai.Client
	WaitGroup         *sync.WaitGroup
	ResultsCh         chan<- models.MessageEvaluation
	ErrCh             chan<- error
}

func ConvertToChatMessages(message []models.Message) []models.ChatMessage {
	chatMessages := make([]models.ChatMessage, len(message))
	for i, msg := range message {
		chatMessages[i] = msg.ChatMessage
	}
	return chatMessages
}

// processMessage handles the evaluation of a message and updates the evaluation record.
func processMessage(r EvaluationRoutine, llm_client *openai.Client, testCount int) (models.MessageEvaluation, error) {
	// Initialize the MessageEvaluation
	messageEvaluation := models.MessageEvaluation{
		LLMID:                    r.LLM.ID,
		LLM:                      r.LLM,
		PromptID:                 r.Prompt.ID,
		Prompt:                   r.Prompt,
		MessageID:                r.Conversation.Messages[r.MessageIndex].ID,
		MessageEvaluationResults: make([]models.MessageEvaluationResult, testCount),
	}

	// Attempt to create the MessageEvaluation record.
	if err := database.DB.Create(&messageEvaluation).Error; err != nil {
		return models.MessageEvaluation{}, fmt.Errorf("failed to create message evaluation: %w", err)
	}

	// Process each message.
	for i := 0; i < testCount; i++ {
		startTime := time.Now()
		messages := r.Conversation.Messages[:r.MessageIndex+1]
		SetSystemPrompt(&messages, r.Prompt.Content)

		llmResponse, err := llm.GetLLMResponse(llm_client, ConvertToChatMessages(messages), r.LLM.Name)
		if err != nil {
			return models.MessageEvaluation{}, fmt.Errorf("failed to get LLM response: %w", err)
		}

		assistantMessage := r.Conversation.Messages[r.MessageIndex]
		similarity, err := nomicai.GetTextSimilarity(llmResponse.Content, assistantMessage.Content)
		if err != nil {
			return models.MessageEvaluation{}, fmt.Errorf("failed to get text similarity: %w", err)
		}

		latency := time.Since(startTime).Milliseconds()
		messageEvaluation.MessageEvaluationResults[i] = models.MessageEvaluationResult{
			Content:    llmResponse.Content,
			LatencyMs:  int(latency),
			Similarity: similarity,
		}
	}

	// No transaction is used, so there is no need to commit.
	return messageEvaluation, nil
}

func SetSystemPrompt(messages *[]models.Message, prompt string) error {
	if len(*messages) == 0 {
		return fmt.Errorf("err: no messages in the conversation")
	}
	if (*messages)[0].Role != "system" {
		*messages = append([]models.Message{{ChatMessage: models.ChatMessage{Role: "system", Content: prompt}}}, *messages...)
	} else {
		(*messages)[0].Content = prompt
	}
	return nil
}

func FindOrCreateLLMs(model_names []string) ([]models.LLM, error) {
	llms := make([]models.LLM, len(model_names))
	for i, name := range model_names {
		llm := models.LLM{Name: name}
		r := database.DB.FirstOrCreate(&llm, "name = ?", name)
		if r.Error != nil {
			return nil, r.Error
		}
		llms[i] = llm
		fmt.Println("LLM: ", llm)
	}
	return llms, nil
}

type EvaluationRoutine struct {
	Conversation *models.Conversation
	MessageIndex uint
	LLM          models.LLM
	Prompt       models.Prompt
}

type EvaluationMessage struct {
	models.Message
	Prompt models.Prompt
}

func GetEvaluationRoutines(conversation *models.Conversation, req *models.CreateLLMEvaluationRequest) ([]EvaluationRoutine, error) {
	evaluationMessages := []EvaluationMessage{}

	basePrompt, err := FindOrCreatePrompt(req.Body.Prompt)

	if err != nil {
		return nil, err
	}

	for i := 1; i < len(conversation.Messages); i++ {
		if conversation.Messages[i].ChatMessage.Role != "assistant" {
			continue
		}

		message := conversation.Messages[i]
		previousIsUser := conversation.Messages[i-1].ChatMessage.Role == "user"

		// Run Eval on all assistant messages
		if previousIsUser && len(req.Body.Messages) == 0 {
			evaluationMessages = append(evaluationMessages, EvaluationMessage{
				Message: message,
				Prompt:  *basePrompt,
			})
			continue
		}
		// Run eval on the provided list of messages
		for _, reqMessage := range req.Body.Messages {
			if message.ID == reqMessage.ID {
				// Throw an error if its not an assistant message
				if message.ChatMessage.Role != "assistant" {
					return nil, fmt.Errorf("err: message with id %s is not an assistant message", message.ID)
				}
				// Throw an error if the previous message is not from the user
				if !previousIsUser {
					return nil, fmt.Errorf("err: message with id %s is not preceded by a user message", message.ID)
				}
				// Get the prompt
				prompt, err := FindOrCreatePrompt(req.Body.Prompt)

				if err != nil {
					return nil, err
				}
				evaluationMessages = append(evaluationMessages, EvaluationMessage{
					Message: message,
					Prompt:  *prompt,
				})
			}
		}
	}

	// Get the LLMs
	llms, err := FindOrCreateLLMs(req.Body.Models)
	if err != nil || len(llms) == 0 {
		return nil, err
	}

	routineCount := len(evaluationMessages) * len(llms)
	if routineCount == 0 {
		return nil, fmt.Errorf("err: no evaluation messages in the conversation")
	}

	routines := make([]EvaluationRoutine, 0, routineCount)

	// Range over the user messages
	for _, evalMessage := range evaluationMessages {
		for _, llm := range llms {
			routine := EvaluationRoutine{
				Conversation: conversation,
				MessageIndex: evalMessage.MessageIndex,
				LLM:          llm,
				Prompt:       evalMessage.Prompt,
			}

			routines = append(routines, routine)
		}
	}
	return routines, nil

}

func CreateLLMEvaluation(ctx context.Context, input *models.CreateLLMEvaluationRequest) (*[]models.Message, error) {
	// Get the conversation
	conversation, err := GetConversation(input.ID)
	if err != nil {
		return nil, err
	}

	// Create the LLM Client
	llm_client := llm.GetLLMClient(map[string]string{})

	// Make a list of tests to be run
	evaluationRoutines, err := GetEvaluationRoutines(conversation, input)
	if err != nil {
		return nil, err
	}

	evaluations := make([]models.MessageEvaluation, len(evaluationRoutines))
	errors := make([]error, len(evaluationRoutines))
	var wg sync.WaitGroup

	for i, routine := range evaluationRoutines {
		wg.Add(1)
		go func(index int, r EvaluationRoutine) {
			defer wg.Done()

			// Process the message and create message evaluation within the processMessage function
			evaluation, err := processMessage(r, llm_client, input.Body.TestCount)

			if err != nil {
				errors[index] = err
				return
			}

			evaluations[index] = evaluation
		}(i, routine)
	}

	wg.Wait()

	for _, err := range errors {
		if err != nil {
			return nil, err
		}
	}

	// Store the results in the database
	result := database.DB.Create(&evaluations)
	if result.Error != nil {
		return nil, result.Error
	}

	// Update the conversation with the results
	for i := range conversation.Messages {
		if conversation.Messages[i].ChatMessage.Role == "assistant" {
			for _, evaluation := range evaluations {
				if evaluation.MessageID == conversation.Messages[i].ID {
					conversation.Messages[i].MessageEvaluations = append(conversation.Messages[i].MessageEvaluations, evaluation)
					break
				}
			}
		}
	}

	return &conversation.Messages, nil
}

func CreateLLMEvaluationAPI(ctx context.Context, input *models.CreateLLMEvaluationRequest) (*models.CreateLLMEvaluationResponse, error) {
	messages, err := CreateLLMEvaluation(ctx, input)
	if err != nil {
		return nil, err
	}
	return &models.CreateLLMEvaluationResponse{
		Body: models.CreateLLMEvaluationResponseBody{
			Messages: messages,
		},
	}, nil
}
