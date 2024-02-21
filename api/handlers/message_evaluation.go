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

		llm_response, err := llm.GetLLMResponse(p.LLMClient, ConvertToChatMessages(messages), p.LLM.Name)
		if err != nil {
			p.ErrCh <- err
			return
		}

		// Get the similarity
		assistantMessage := p.Conversation.Messages[p.EvalMessageIndex+1]
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

func GetEvaluationRoutines(conversation *models.Conversation, llms []models.LLM, prompt *models.Prompt) ([]EvaluationRoutine, error) {
	// Get how many messages are from the user
	useMessages := []models.Message{}
	for _, message := range conversation.Messages {
		if message.Role == "user" {
			useMessages = append(useMessages, message)
		}
	}

	routineCount := len(useMessages) * len(llms)
	if routineCount == 0 {
		return nil, fmt.Errorf("err: no user messages in the conversation")
	}

	routines := make([]EvaluationRoutine, 0, routineCount)

	// Range over the user messages
	for _, userMessage := range useMessages {
		for _, llm := range llms {
			routine := EvaluationRoutine{
				Conversation:     conversation,
				EvalMessageIndex: userMessage.MessageIndex + 1,
				LLM:              llm,
				Prompt:           *prompt,
			}

			routines = append(routines, routine)
		}
	}
	return routines, nil

}

func CreateLLMEvaluationAPI(ctx context.Context, input *models.CreateLLMEvaluationRequest) (*models.CreateLLMEvaluationResponse, error) {
	// Get the conversation

	conversation, err := GetConversation(ctx, input.ID)

	if err != nil {
		return nil, err
	}

	// Get the prompt
	prompt, err := FindOrCreatePrompt(input.Body.Prompt)

	if err != nil || prompt.ID == "" {
		return nil, err
	}

	// Set the system prompt
	SetSystemPrompt(&conversation.Messages, input.Body.Prompt)

	// Get the LLMs
	llms, err := FindOrCreateLLMs(input.Body.Models)
	if err != nil || len(llms) == 0 {
		return nil, err
	}

	// Create the LLM Client
	llm_client := llm.GetLLMClient(map[string]string{})

	// Make a list of tests to be run
	evaluationRoutines, err := GetEvaluationRoutines(conversation, llms, prompt)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	routineCount := len(evaluationRoutines)
	fmt.Println("Routine Count: ", routineCount)
	errCh := make(chan error, routineCount)
	resultsCh := make(chan models.MessageEvaluation, routineCount)

	for _, routine := range evaluationRoutines {
		// Create the message evaluation
		message_evaluation := models.MessageEvaluation{
			LLMID:                    routine.LLM.ID,
			PromptID:                 routine.Prompt.ID,
			MessageID:                routine.Conversation.Messages[routine.EvalMessageIndex+1].ID,
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
			return &models.CreateLLMEvaluationResponse{}, err
		}
	}
	fmt.Println("Routine Count: ", routineCount)
	evaluations := make([]models.MessageEvaluation, routineCount)

	// Load the messages into the results
	i := 0
	for result := range resultsCh {
		result.ComputeAverageSimilarity()
		evaluations[i] = result
		i++

		// Store the results in the database
		r := database.DB.Create(&result)
		if r.Error != nil {
			return nil, r.Error
		}
	}

	fmt.Println("ran this far", i)

	return &models.CreateLLMEvaluationResponse{
		Body: models.CreateLLMEvaluationResponseBody{
			Results: &evaluations,
		},
	}, err
}
