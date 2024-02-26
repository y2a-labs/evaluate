package apihandlers

import (
	"fmt"
	"os"
	database "script_validation"
	"script_validation/internal/llm"
	"script_validation/internal/nomicai"
	"script_validation/limiter"
	"script_validation/models"
	valid "script_validation/validator"
	"script_validation/views/pages"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
	"golang.org/x/time/rate"
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
func processMessage(r EvaluationRoutine, llm_client *openai.Client, embeddingClient *nomicai.Client, limiter *rate.Limiter, ctx *fiber.Ctx) (*models.MessageEvaluation, error) {
	
    // Wait for permission before making the first request
    if err := limiter.Wait(ctx.Context()); err != nil {
        return nil, fmt.Errorf("rate limiter wait error: %w", err)
    }
	testCount := r.Conversation.EvalTestCount

	// Initialize the MessageEvaluation
	messageEvaluation := models.MessageEvaluation{
		BaseModel:                models.BaseModel{ID: uuid.New().String()},
		LLMID:                    r.LLM.ID,
		LLM:                      r.LLM,
		Prompt:                   r.Prompt,
		MessageID:                r.Conversation.Messages[r.MessageIndex].ID,
		MessageEvaluationResults: make([]models.MessageEvaluationResult, testCount),
	}

	// Process each message.
	for i := 0; i < testCount; i++ {
        if err := limiter.Wait(ctx.Context()); err != nil {
            return nil, fmt.Errorf("rate limiter wait error before request %d: %w", i, err)
        }
		fmt.Println("Processing message", i+1, "of", testCount, "for message", r.MessageIndex, "of", len(r.Conversation.Messages))
		startTime := time.Now()
		messages := r.Conversation.Messages[:r.MessageIndex]
		SetSystemPrompt(&messages, r.Prompt)

		llmResponse, err := llm.GetLLMResponse(llm_client, ConvertToChatMessages(messages), r.LLM.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get LLM response: %w", err)
		}

		assistantMessage := r.Conversation.Messages[r.MessageIndex]
		similarity, err := embeddingClient.GetTextSimilarity(llmResponse.Content, assistantMessage.Content)
		if err != nil {
			return nil, fmt.Errorf("failed to get text similarity: %w", err)
		}

		latency := time.Since(startTime).Milliseconds()
		messageEvaluation.MessageEvaluationResults[i] = models.MessageEvaluationResult{
			Content:    llmResponse.Content,
			LatencyMs:  int(latency),
			Similarity: similarity,
		}
	}
	messageEvaluation.ComputeAverageSimilarity()

	// No transaction is used, so there is no need to commit.
	return &messageEvaluation, nil
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

type EvaluationRoutine struct {
	Conversation *models.Conversation
	MessageIndex uint
	LLM          models.LLM
	Prompt       string
}

type EvaluationMessage struct {
	models.Message
	Prompt string
}

func GetEvaluationRoutines(conversation *models.Conversation, req *PostEvaluationAPIRequest) ([]EvaluationRoutine, error) {
	evaluationMessages := []EvaluationMessage{}

	basePrompt := req.Prompt

	// out of end index, and len messages, pick the lower value
	endIndex := len(conversation.Messages)
	if req.EndIndex < endIndex {
		endIndex = req.EndIndex + 1
	}

	startIndex := 1
	if req.StartIndex+1 > startIndex {
		startIndex = req.StartIndex
	}

	for i := startIndex; i < endIndex; i++ {
		if conversation.Messages[i].ChatMessage.Role != "assistant" {
			continue
		}

		message := conversation.Messages[i]
		previousIsUser := conversation.Messages[i-1].ChatMessage.Role == "user"

		// Run Eval on all assistant messages
		if previousIsUser && len(req.Messages) == 0 {
			evaluationMessages = append(evaluationMessages, EvaluationMessage{
				Message: message,
				Prompt:  basePrompt,
			})
			continue
		}
		// Run eval on the provided list of messages
		for _, reqMessage := range req.Messages {
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
				prompt := req.Prompt

				evaluationMessages = append(evaluationMessages, EvaluationMessage{
					Message: message,
					Prompt:  prompt,
				})
			}
		}
	}

	// Get the LLMs
	llms, err := FindOrCreateLLMs(req.Models)
	if err != nil || len(llms) == 0 {
		return nil, err
	}

	routineCount := len(evaluationMessages) * len(llms)
	if routineCount == 0 {
		return nil, fmt.Errorf("err: no evaluation messages in the conversation")
	}

	var routines []EvaluationRoutine

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

func CreateLLMEvaluation(req *PostEvaluationAPIRequest, ctx *fiber.Ctx) (*[]models.Message, error) {
	// Get the conversation
	conversation, err := GetConversation(req.ID)
	if err != nil {
		return nil, err
	}
	fmt.Println("Test Count", req.TestCount)

	database.DB.Model(&conversation).UpdateColumns(models.Conversation{
		EvalEndIndex:   req.EndIndex,
		EvalStartIndex: req.StartIndex,
		EvalPrompt:     req.Prompt,
		EvalModels:     req.Models,
		EvalTestCount:  req.TestCount,
	})

	llm_clients := make(map[string]*openai.Client)
	providers := make(map[string]*models.Provider)

	for i, model := range req.Models {
		providerId := strings.Split(model, "/")[0]
		req.Models[i] = strings.TrimPrefix(model, providerId+"/")
		// check if provider exists in provider map
		if _, ok := llm_clients[providerId]; !ok {
			// if not, add it to the map
			provider := &models.Provider{ID: providerId}
			tx := database.DB.First(&provider)
			if tx.Error != nil {
				return nil, tx.Error
			}
			llm_clients[providerId] = llm.GetLLMClient(provider)
			providers[providerId] = provider
			limiter.Manager.GetLimiter(*provider)
		}

	}

	embeddingsClient := nomicai.NewClient(os.Getenv("NOMICAI_API_KEY"))

	// Make a list of tests to be run
	evaluationRoutines, err := GetEvaluationRoutines(conversation, req)
	if err != nil {
		return nil, err
	}

	evaluations := make([]*models.MessageEvaluation, len(evaluationRoutines))
	errors := make([]error, len(evaluationRoutines))
	var wg sync.WaitGroup

	for i, routine := range evaluationRoutines {
		wg.Add(1)
		go func(index int, r EvaluationRoutine) {
			defer wg.Done()

			// Retrieve the provider and its rate limiter
			provider := providers[r.LLM.ProviderID]
			limiter := limiter.Manager.GetLimiter(*provider)
			// Now proceed to make the request
			llm_client := llm_clients[r.LLM.ProviderID]
			evaluation, err := processMessage(r, llm_client, embeddingsClient, limiter, ctx)

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
	result := database.DB.Save(&evaluations)
	if result.Error != nil {
		return nil, result.Error
	}

	// Create a map to store slices of evaluations by their MessageID
	evaluationsMap := make(map[string][]models.MessageEvaluation)
	for _, evaluation := range evaluations {
		// Append the evaluation to the slice for its MessageID
		evaluationsMap[evaluation.MessageID] = append(evaluationsMap[evaluation.MessageID], *evaluation)
	}

	// Sort the evaluations for each message by average similarity
	for messageID, evaluations := range evaluationsMap {
		evaluationsMap[messageID] = sortEvaluationsByAverageSimilarity(evaluations)
	}

	// Update the conversation with the results
	for i, message := range conversation.Messages {
		if message.ChatMessage.Role != "assistant" {
			continue
		}
		if evaluations, ok := evaluationsMap[message.ID]; ok {
			// Prepend the new evaluations to the existing ones
			conversation.Messages[i].MessageEvaluations = append(evaluations, message.MessageEvaluations...)
		}
	}

	return &conversation.Messages, nil
}

// sortEvaluationsByAverageSimilarity sorts a slice of evaluations by their average similarity in descending order
func sortEvaluationsByAverageSimilarity(evaluations []models.MessageEvaluation) []models.MessageEvaluation {
	sort.Slice(evaluations, func(i, j int) bool {
		return evaluations[i].AverageSimilarity > evaluations[j].AverageSimilarity
	})
	return evaluations
}

type PostEvaluationAPIRequest struct {
	ID         string   `path:"id"`
	Models     []string `json:"models" form:"models"`
	TestCount  int      `json:"test_count" form:"test_count"`
	StartIndex int      `json:"start_index" form:"start_index"`
	EndIndex   int      `json:"end_index" form:"end_index"`
	Prompt     string   `json:"prompt" form:"prompt"`
	Messages   []struct {
		ID     string `json:"id" form:"id"`
		Prompt string `json:"prompt" form:"prompt"`
	}
}

type PostEvaluationAPIParams struct {
	ID string `path:"id"`
}

func PostEvaluationAPI(ctx *fiber.Ctx) error {
	req := &PostEvaluationAPIRequest{}

	err := valid.ValidateStruct(ctx, req)
	if err != nil {
		return err
	}
	messages, err := CreateLLMEvaluation(req, ctx)
	if err != nil {
		return err
	}
	if ctx.Accepts("html") != "" {
		return Render(ctx, pages.Messages(*messages))
	}
	return ctx.JSON(messages)
}
