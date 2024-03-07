package apihandlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"script_validation/internal/llm"
	"script_validation/internal/nomicai"
	"script_validation/models"
	"script_validation/web/pages"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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

func MessageToOpenaiMessage(messages []models.Message) []openai.ChatCompletionMessage {
	msgs := make([]openai.ChatCompletionMessage, len(messages))

	for i, msg := range messages {
		msgs[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}
	return msgs
}

// processSingleMessage processes a single message and returns its evaluation.
func processSingleMessage(ctx context.Context, r EvaluationRoutine, llmClient *openai.Client, embeddingClient *nomicai.Client) (*models.Message, error) {

	// Limit the number of messages to be processed
	messages := r.Conversation.Messages[:r.MessageIndex]

	SetSystemPrompt(&messages, r.Prompt)

	openaiMessages := MessageToOpenaiMessage(messages)

	// Turn the message into a chat completion request
	request := openai.ChatCompletionRequest{
		Model:    r.LLM.ID,
		Stream:   true,
		Messages: openaiMessages,
	}

	// Measure how long it takes for the first token
	startTime := time.Now()
	// Create the chat completion stream
	stream, err := llmClient.CreateChatCompletionStream(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get LLM response: %w", err)
	}
	defer stream.Close()
	firstTokenLatencyMs := 0
	chunkCount := 0
	responseBuffer := ""

	// Reads back the streamed response
	for {
		resp, err := stream.Recv()

		// If the stream is done, break out of the loop
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("failed to get LLM response: %w", err)
		}
		chunkCount++
		responseBuffer += resp.Choices[0].Delta.Content

		if firstTokenLatencyMs == 0 {
			firstTokenLatencyMs = int(time.Since(startTime).Milliseconds())
		}
	}

	responseEmbedding, err := embeddingClient.EmbedText([]string{responseBuffer}, nomicai.Clustering)
	if err != nil {
		return nil, fmt.Errorf("failed to get text embedding: %w", err)
	}

	totalLatencyMs := int(time.Since(startTime).Milliseconds())
	
	message, err := models.NewEvaluationMessage(&models.Message{
		Role:          "assistant",
		Content:       responseBuffer,
		ConversationID: r.Conversation.ID,
		PromptID:      r.Prompt,
		Metadata: models.MessageMetadata{
			StartLatencyMs: firstTokenLatencyMs,
			EndLatencyMs:   totalLatencyMs,
			TokenCount:     chunkCount,
			Embedding:      responseEmbedding.Embeddings[0],
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create evaluation message: %w", err)
	}

	return message, nil
}

func SetSystemPrompt(messages *[]models.Message, prompt string) error {
	if len(*messages) == 0 {
		return fmt.Errorf("err: no messages in the conversation")
	}
	if (*messages)[0].Role != "system" {
		*messages = append([]models.Message{{Role: "system", Content: prompt}}, *messages...)
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

func (app *App) GetEvaluationRoutines(conversation *models.Conversation, req *PostEvaluationAPIRequest, providers map[string]*models.Provider, evalConversationID string) ([]EvaluationRoutine, error) {
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

	// Gets the list of all of the messages to run eval on
	for i := startIndex; i < endIndex; i++ {
		if conversation.Messages[i].Role != "assistant" {
			continue
		}

		message := conversation.Messages[i]
		previousIsUser := conversation.Messages[i-1].Role == "user"

		// Run Eval on all assistant messages
		if previousIsUser && len(req.Messages) == 0 {
			evaluationMessages = append(evaluationMessages, EvaluationMessage{
				Message: message,
				Prompt:  basePrompt,
			})
			continue
		}
	}

	// Get the LLMs
	llms, err := app.FindOrCreateLLMs(req.Models)
	if err != nil || len(llms) == 0 {
		return nil, err
	}

	routineCount := len(evaluationMessages) * len(llms)
	if routineCount == 0 {
		return nil, fmt.Errorf("err: no evaluation messages in the conversation")
	}

	var routines []EvaluationRoutine

	// Create a routine for each model
	for _, evalMessage := range evaluationMessages {
		for _, llm := range llms {
			llm.Provider = *providers[llm.ProviderID]
			llm.ProviderID = providers[llm.ProviderID].ID
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

func (app *App) CreateLLMEvaluation(req *PostEvaluationAPIRequest, ctx *fiber.Ctx) (*[]models.Message, error) {
	// Get the conversation
	conversation, err := app.GetConversation(req.ID)
	if err != nil {
		return nil, err
	}

	app.Db.Model(&conversation).UpdateColumns(models.Conversation{
		EvalEndIndex:   req.EndIndex,
		EvalStartIndex: req.StartIndex,
		EvalPrompt:     req.Prompt,
		EvalModels:     req.Models,
		EvalTestCount:  req.TestCount,
	})

	llm_clients := make(map[string]*openai.Client)
	providers := make(map[string]*models.Provider)

	// Get the providers and init the llm clients
	for i, model := range req.Models {
		providerId := strings.Split(model, "/")[0]
		req.Models[i] = strings.TrimPrefix(model, providerId+"/")
		// check if provider exists in provider map
		if _, ok := llm_clients[providerId]; !ok {
			// if not, add it to the map
			provider := &models.Provider{ID: providerId}
			tx := app.Db.First(&provider)
			if tx.Error != nil {
				return nil, tx.Error
			}
			fmt.Println(provider)
			apiKey, err := app.Decrypt(provider.EncryptedAPIKey)
			if err != nil || apiKey == "" {
				return nil, err
			}
			llm_clients[providerId] = llm.NewClient(provider.BaseUrl, apiKey)
			providers[providerId] = provider
			app.limiter.GetLimiter(*provider)
		}

	}

	embeddingProvider := &models.Provider{ID: "nomicai"}
	tx := app.Db.First(embeddingProvider)
	if tx.Error != nil {
		return nil, tx.Error
	}
	apiKey, err := app.Decrypt(embeddingProvider.EncryptedAPIKey)
	if err != nil {
		return nil, err
	}
	embeddingsClient := nomicai.NewClient(apiKey)

	// Create the eval conversation
	evalConversation := &models.Conversation{
		BaseModel: models.BaseModel{ID: uuid.New().String()},
	}
	tx = app.Db.Save(evalConversation)
	if tx.Error != nil {
		return nil, tx.Error
	}

	// Make a list of tests to be run
	evaluationRoutines, err := app.GetEvaluationRoutines(conversation, req, providers, evalConversation.ID)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	evalCount := len(evaluationRoutines) * req.TestCount
	errorChan := make(chan error, evalCount)
	evalMessageChan := make(chan *models.Message, evalCount)

	fmt.Println("Evaluating", len(evaluationRoutines), "messages with", req.TestCount, "tests each")
	// Run the routines
	for _, routine := range evaluationRoutines {
		wg.Add(req.TestCount)
		provider := providers[routine.LLM.ProviderID]
		llmClient := llm_clients[routine.LLM.ProviderID]

		limiter := app.limiter.GetLimiter(*provider) // Use the rate limiter
		for j := 0; j < req.TestCount; j++ {
			go func(r EvaluationRoutine) {
				defer wg.Done()
				// Wait for permission before msaking the request
				if err := limiter.Wait(ctx.Context()); err != nil {
					errorChan <- fmt.Errorf("rate limiter wait error: %w", err)
					return
				}

				evalMessage, err := processSingleMessage(ctx.Context(), r, llmClient, embeddingsClient)
				evalMessage.ConversationID = evalConversation.ID
				if err != nil {
					errorChan <- err
					return
				}

				evalMessageChan <- evalMessage
			}(routine)
		}
	}
	wg.Wait()
	close(errorChan)
	close(evalMessageChan)

	// Check for errors
	for err := range errorChan {
		if err != nil {
			return nil, err
		}
	}

	// Take all of the messages from the channel
	evalMessages := make([]*models.Message, evalCount)
	for i := 0; i < evalCount; i++ {
		evalMessages[i] = <-evalMessageChan
	}

	// Store the evaluation messages in the database
	tx = app.Db.Save(&evalMessages)
	if tx.Error != nil {
		return nil, tx.Error
	}

	/*
		// Create a map to store slices of evaluations by their MessageID
		evaluationsMap := make(map[string][]models.MessageEvaluation)
		for _, evaluation := range evaluations {
			// Append the evaluation to the slice for its MessageID
			evaluationsMap[evaluation.MessageID] = append(evaluationsMap[evaluation.MessageID], evaluation)
		}

		// Sort the evaluations for each message by average similarity
		for messageID, evaluations := range evaluationsMap {
			evaluationsMap[messageID] = sortEvaluationsByAverageSimilarity(evaluations)
		}

		// Update the conversation with the results
		for i, message := range conversation.Messages {
			if message.Role != "assistant" {
				continue
			}
			if evaluations, ok := evaluationsMap[message.ID]; ok {
				// Prepend the new evaluations to the existing ones
				conversation.Messages[i].MessageEvaluations = append(evaluations, message.MessageEvaluations...)
			}
		}
	*/

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
	ID         string   `path:"id" validate:"required"`
	Models     []string `json:"models" form:"models" validate:"required"`
	TestCount  int      `json:"test_count" form:"test_count" validate:"required"`
	StartIndex int      `json:"start_index" form:"start_index"`
	EndIndex   int      `json:"end_index" form:"end_index" validate:"required"`
	Prompt     string   `json:"prompt" form:"prompt" validate:"required"`
	Messages   []struct {
		ID     string `json:"id" form:"id"`
		Prompt string `json:"prompt" form:"prompt"`
	}
}

type PostEvaluationAPIParams struct {
	ID string `path:"id"`
}

func (app *App) PostEvaluationAPI(ctx *fiber.Ctx) error {
	req := &PostEvaluationAPIRequest{}

	err := app.ValidateStruct(ctx, req)
	if err != nil {
		return err
	}
	messages, err := app.CreateLLMEvaluation(req, ctx)
	if err != nil {
		return err
	}
	if ctx.Accepts("html") != "" {
		return Render(ctx, pages.Messages(*messages))
	}
	return ctx.JSON(messages)
}
