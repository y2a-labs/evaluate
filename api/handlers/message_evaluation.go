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

func processMessage(p ProcessMessage) {
	defer p.WaitGroup.Done()

	for i := range p.TestCount {
		startTime := time.Now()

		// Get the LLMs Response
		messages := p.Conversation.Messages[:p.EvalMessageIndex+1]

		SetSystemPrompt(&messages, p.Prompt.Content)

		llm_response, err := llm.GetLLMResponse(p.LLMClient, ConvertToChatMessages(messages), p.LLM.Name)
		if err != nil {
			p.ErrCh <- err
			return
		}

		// Get the similarity
		assistantMessage := p.Conversation.Messages[p.EvalMessageIndex]
		similarity, err := nomicai.GetTextSimilarity(llm_response.Content, assistantMessage.Content)
		if err != nil {
			p.ErrCh <- err
			return
		}

		// Capture the latency of the request
		endTime := time.Now()
		latency := endTime.Sub(startTime).Milliseconds()

		p.MessageEvaluation.MessageEvaluationResults[i] = models.MessageEvaluationResult{
			Content:    llm_response.Content,
			LatencyMs:  int(latency),
			Similarity: similarity,
		}

	}

	p.ResultsCh <- p.MessageEvaluation
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
	Conversation     *models.Conversation
	EvalMessageIndex uint
	LLM              models.LLM
	Prompt           models.Prompt
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
				Conversation:     conversation,
				EvalMessageIndex: evalMessage.MessageIndex,
				LLM:              llm,
				Prompt:           evalMessage.Prompt,
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

	var wg sync.WaitGroup
	routineCount := len(evaluationRoutines)
	errCh := make(chan error, routineCount)
	resultsCh := make(chan models.MessageEvaluation, routineCount)

	for _, routine := range evaluationRoutines {
		// Create the message evaluation
		message_evaluation := models.MessageEvaluation{
			LLMID:                    routine.LLM.ID,
			LLM:                      routine.LLM,
			PromptID:                 routine.Prompt.ID,
			Prompt:                   routine.Prompt,
			MessageID:                routine.Conversation.Messages[routine.EvalMessageIndex].ID,
			MessageEvaluationResults: make([]models.MessageEvaluationResult, input.Body.TestCount),
		}
		result := database.DB.Create(&message_evaluation)
		if result.Error != nil {
			return nil, result.Error
		}

		wg.Add(1)
		go processMessage(ProcessMessage{
			TestCount:         input.Body.TestCount,
			EvaluationRoutine: routine,
			MessageEvaluation: message_evaluation,
			LLMClient:         llm_client,
			ResultsCh:         resultsCh,
			ErrCh:             errCh,
			WaitGroup:         &wg,
		})
	}

	wg.Wait()
	close(errCh)
	close(resultsCh)

	for err := range errCh {
		if err != nil {
			return nil, err
		}
	}
	evaluations := make([]models.MessageEvaluation, routineCount)

	// Load the messages into the results
	i := 0
	for result := range resultsCh {
		result.ComputeAverageSimilarity()
		evaluations[i] = result
		i++
	}

	// Store the results in the database
	r := database.DB.Create(&evaluations)

	if r.Error != nil {
		return nil, r.Error
	}

	// Return an array of messages with the results
	for i, message := range conversation.Messages {
		if message.ChatMessage.Role != "assistant" {
			continue
		}
		// Loop through the evaluations and find the one that matches the message
		for _, evaluation := range evaluations {
			if evaluation.MessageID == message.ID {
				conversation.Messages[i].MessageEvaluations = append(conversation.Messages[i].MessageEvaluations, evaluation)
				// Remove the message evaluation from the array
				evaluations = append(evaluations[:i-1], evaluations[i:]...)
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
