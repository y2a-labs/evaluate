package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"script_validation/internal/nomicai"
	"script_validation/models"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

type RunTestInput struct {
	Context      context.Context
	RunCount     int `json:"run_count"`
	Prompt       *models.Prompt
	Conversation *models.Conversation
	LLMs         []*models.LLM
}

type ExecuteTestInput struct {
	Context        context.Context
	RunCount       int
	PromptID       string
	ConversationID string
}

func (s *Service) prepareTestData(input ExecuteTestInput) (*RunTestInput, error) {
	// Get the conversation with messages
	conversation, err := s.GetConversationWithMessages(input.ConversationID)
	if err != nil {
		return nil, err
	}

	prompt, err := s.GetPrompt(input.PromptID)
	if err != nil {
		return nil, err
	}

	llmIDs := make([]string, len(conversation.TestModels))
	for i, item := range conversation.TestModels {
		llmIDs[i] = item.Model
	}

	llms, err := s.GetLLMByIds(llmIDs)
	if err != nil {
		return nil, err
	}

	result := &RunTestInput{
		Context:      input.Context,
		Conversation: conversation,
		Prompt:       prompt,
		LLMs:         llms,
		RunCount:     input.RunCount,
	}
	return result, nil
}

func (s *Service) ExecuteTestWorkflow(input ExecuteTestInput) ([]*models.Message, error) {
	if input.ConversationID == "" || input.PromptID == "" || input.RunCount < 1 {
		return nil, fmt.Errorf("error trying to validate workflow input")
	}
	// Get all of the data
	preparedInput, err := s.prepareTestData(input)
	if err != nil {
		return nil, err
	}

	// Generate the results
	resultsChan, totalResultCount, err := s.runTest(preparedInput)
	if err != nil {
		return nil, err
	}

	// Save the results in batches
	results, err := s.saveResultsInBatches(resultsChan, totalResultCount, input.RunCount)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (s *Service) saveResultsInBatches(resultsChan chan TestResult, totalResultCount, batchSize int) ([]*models.Message, error) {
	resultCount := 0
	batchIndex := 0 // This will track where in allResults the next batch starts.
	allResults := make([]*models.Message, totalResultCount)

	for result := range resultsChan {
		if result.Err != nil {
			return nil, result.Err
		}

		// Place the result directly into the allResults slice at the correct position.
		allResults[batchIndex+resultCount] = result.Message
		resultCount++

		if resultCount == batchSize {
			// Process the batch if needed, for example, saving to a database.
			tx := s.Db.Save(allResults[batchIndex : batchIndex+batchSize])
			if tx.Error != nil {
				return nil, tx.Error
			}
			batchIndex += batchSize
			resultCount = 0
		}
	}

	// Handle any remaining results if they don't fill a complete batch.
	if resultCount > 0 {
		tx := s.Db.Save(allResults[batchIndex : batchIndex+resultCount])
		if tx.Error != nil {
			return nil, tx.Error
		}
	}

	return allResults, nil
}

type TestResult struct {
	Message *models.Message
	Err     error
}

type TestResultsOutput struct {
	Overview []ModelTestResultsOverview
	Test     models.Conversation
}

type ModelTestResultsOverview struct {
	LLM                 models.LLM
	Score               float64
	FirstTokenLatencyMs int
	TotalResponseTimeMs int
}

func (s *Service) GetTestResults(conversationID, promptID string) (*models.Conversation, error) {
	// Assuming conversation struct is properly defined and includes Messages that can have Metadata and TestMessages.
	conversation := &models.Conversation{BaseModel: models.BaseModel{ID: conversationID}}

	// Loads the conversation
	tx := s.Db.First(conversation)

	if tx.Error != nil {
		return nil, tx.Error
	}

	// Loads the messages without a test_message_id, then preload all test messages and their metadata
	tx = s.Db.Model(&models.Message{}).
		Preload("Metadata").
		Preload("TestMessages.Metadata").
		Where("test_message_id = ?", "").
		Order("message_index").
		Find(&conversation.Messages)

	if tx.Error != nil {
		return nil, tx.Error
	}

	for _, message := range conversation.Messages {
		// If the message is not an assistant message, skip it
		if message.Role != "assistant" {
			continue
		}
		// Calculate the score for every TestMessage
		scoreSum := 0.0
		for _, testMessage := range message.TestMessages {
			// should be fixed
			score, err := s.embeddingProviders["nomicai"].client.CosineSimilarity(testMessage.Metadata.Embedding, message.Metadata.Embedding)
			if err != nil {
				return nil, fmt.Errorf("failed to calculate cosine similarity: %w", err)
			}
			scoreStr := fmt.Sprintf("%.2f", score*100)
			testMessage.Score, _ = strconv.ParseFloat(scoreStr, 64)
			scoreSum += testMessage.Score
		}
		// Calculate the average score for the message
		message.Score = scoreSum / float64(len(message.TestMessages))
		scoreStr := fmt.Sprintf("%.2f", message.Score)
		message.Score, _ = strconv.ParseFloat(scoreStr, 64)
	}

	if promptID == "" {
		promptID = conversation.PromptID
	}

	prompt := &models.Prompt{BaseModel: models.BaseModel{ID: promptID}}

	conversation.Prompt = *prompt

	return conversation, nil
}

func (s *Service) runTest(input *RunTestInput) (chan TestResult, int, error) {

	// Set the prompt as the first message in the conversation
	input.Conversation.Messages = append([]*models.Message{{Role: "system", Content: input.Prompt.Content}}, input.Conversation.Messages...)

	// Gets the list of end indexes to test on
	testIndexes, err := getTestIndexes(input.Conversation.Messages)
	if err != nil {
		return nil, 0, err
	}

	var wg sync.WaitGroup
	testCount := len(testIndexes) * input.RunCount * len(input.LLMs)
	testResultChan := make(chan TestResult, testCount)

	for _, llm := range input.LLMs {
		embeddingProvider := s.embeddingProviders["nomicai"]
		llmProvider := s.llmProviders[llm.ProviderID]
		limiter := s.limiter.GetLimiter(llmProvider.Provider)
		for _, testIndex := range testIndexes {
			wg.Add(input.RunCount)
			messages := input.Conversation.Messages[:testIndex]
			for range input.RunCount {
				go func() {
					defer wg.Done()
					// Wait for permission before msaking the request
					if err := limiter.Wait(input.Context); err != nil {
						testResultChan <- TestResult{Err: fmt.Errorf("rate limiter wait error: %w", err)}
						return
					}

					// Process the prompt
					resultMessage, err := processPrompt(input.Context, messages, llm.ID, llmProvider.client, embeddingProvider.client)
					if err != nil {
						testResultChan <- TestResult{Err: fmt.Errorf("problem processing the prompt: %w", err)}
						return
					}
					resultMessage.PromptID = input.Prompt.ID
					resultMessage.TestMessageID = input.Conversation.Messages[testIndex].ID
					resultMessage.ConversationID = input.Conversation.ID

					testResultChan <- TestResult{
						Message: resultMessage,
						Err:     nil,
					}
				}()
			}
		}
	}

	// Closes the routines
	go func() {
		wg.Wait()
		close(testResultChan)
	}()

	return testResultChan, testCount, nil
}

func processPrompt(ctx context.Context, messages []*models.Message, model string, llmClient *openai.Client, embeddingClient *nomicai.Client) (*models.Message, error) {

	// Turn the message into openai format
	openaiMessages := make([]openai.ChatCompletionMessage, len(messages))

	for i, msg := range messages {
		openaiMessages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Turn the message into a chat completion request
	request := openai.ChatCompletionRequest{
		Model:    model,
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

	if responseBuffer == "" {
		return nil, fmt.Errorf("no content in generated response")
	}

	// Generate text embeddings
	responseEmbedding, err := embeddingClient.EmbedText([]string{responseBuffer}, nomicai.Clustering)
	if err != nil {
		return nil, fmt.Errorf("failed to get text embedding: %w", err)
	}

	totalLatencyMs := int(time.Since(startTime).Milliseconds())

	message := &models.Message{
		Role:    "assistant",
		Content: responseBuffer,
		Metadata: &models.MessageMetadata{
			BaseModel: models.BaseModel{
				ID: uuid.NewString(),
			},
			StartLatencyMs:   firstTokenLatencyMs,
			EndLatencyMs:     totalLatencyMs,
			OutputTokenCount: chunkCount,
			Embedding:        responseEmbedding.Embeddings[0],
		},
	}

	return message, nil
}

func getTestIndexes(messages []*models.Message) ([]int, error) {
	evalIndexes := []int{}
	// Gets the list of all of the messages to run eval on
	for i, message := range messages {
		if message.Role != "assistant" || i == 0 {
			continue
		}
		previousIsUser := messages[i-1].Role == "user"

		// Run Eval on all assistant messages
		if previousIsUser {
			evalIndexes = append(evalIndexes, i)
		}
	}
	// Error if no messages found to evaluate
	if len(evalIndexes) == 0 {
		return []int{}, fmt.Errorf("no messaged found to evaluate")
	}
	return evalIndexes, nil
}

type TestManager interface {
	ExecuteTestWorkflow(input ExecuteTestInput) ([]*models.Message, error)
}
