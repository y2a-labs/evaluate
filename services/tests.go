package service

import (
	"context"
	"fmt"
	"script_validation/internal/nomicai"
	"script_validation/models"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
)

type RunTestInput struct {
	Context      context.Context
	RunCount     int `json:"run_count"`
	Conversation *models.Conversation
	TestIndexes  []int
	LLMs         []*models.LLM
}

type ExecuteTestInput struct {
	Context        context.Context
	TestMessageID  string
	RunCount       int
	ConversationID string
}

func (s *Service) prepareTestData(input ExecuteTestInput) (*RunTestInput, error) {
	// Get the conversation with messages
	conversation, err := s.GetConversationWithMessages(input.ConversationID, -1)
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

	testIndexes, err := getTestIndexes(conversation.Messages, input.TestMessageID)
	if err != nil {
		return nil, err
	}

	result := &RunTestInput{
		Context:      input.Context,
		Conversation: conversation,
		LLMs:         llms,
		RunCount:     input.RunCount,
		TestIndexes:  testIndexes,
	}
	return result, nil
}

func (s *Service) ExecuteTestWorkflow(input ExecuteTestInput) ([]*models.Message, error) {
	if input.ConversationID == "" || input.RunCount < 1 {
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

func (s *Service) GetTestList() ([]*models.Conversation, error) {
	conversations := []*models.Conversation{}
	tx := s.Db.Where("is_test = ?", true).Limit(50).Order("created_at DESC").Find(&conversations)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return conversations, nil
}

func (s *Service) GetTest(conversationID string, selectedVersion int) (*models.Conversation, error) {

	conversation, err := s.GetConversationWithMessages(conversationID, selectedVersion)
	if err != nil {
		return nil, err
	}

	// For each message, preload Metadata and TestMessages (and their Metadata).
	for i, message := range conversation.Messages {
		if err := s.Db.Preload("Metadata").
			Preload("TestMessages", func(db *gorm.DB) *gorm.DB {
				return db.Where("conversation_version = ?", conversation.SelectedVersion).Preload("Metadata")
			}).
			Find(&message).Error; err != nil {
			return nil, err
		}
		conversation.Messages[i] = message
	}

	for _, message := range conversation.Messages {
		// If the message is not an assistant message, skip it
		if message.Role != "assistant" {
			continue
		}

		if message.Metadata == nil {
			return nil, fmt.Errorf("no metadata found for message")
		}

		// Calculate the score for every TestMessage
		uniqueTestMessages := make(map[string]*models.Message)
		for _, testMessage := range message.TestMessages {
			// Calculate the score
			embeddingProvider, ok := s.embeddingProviders["nomicai"]
			if !ok {
				return nil, fmt.Errorf("embedding provider not found")
			}

			// Make sure both messages have embeddings
			if testMessage.Metadata == nil || message.Metadata == nil {
				return nil, fmt.Errorf("missing metadata on message")
			}

			score, err := embeddingProvider.client.CosineSimilarity(testMessage.Metadata.Embedding, message.Metadata.Embedding)
			if err != nil {
				return nil, fmt.Errorf("failed to calculate cosine similarity: %w", err)
			}
			scoreStr := fmt.Sprintf("%.2f", score*100)
			testMessage.Score, _ = strconv.ParseFloat(scoreStr, 64)

			// If the test message content is not already in the map, add it
			if _, exists := uniqueTestMessages[testMessage.Content]; !exists {
				uniqueTestMessages[testMessage.Content] = testMessage
			}
		}

		// Replace the original TestMessages slice with the unique ones
		message.TestMessages = make([]*models.Message, 0, len(uniqueTestMessages))
		for _, testMessage := range uniqueTestMessages {
			message.TestMessages = append(message.TestMessages, testMessage)
		}

		// Sort TestMessages by score in descending order
		sort.Slice(message.TestMessages, func(i, j int) bool {
			return message.TestMessages[i].Score > message.TestMessages[j].Score
		})
	}

	return conversation, nil
}

func (s *Service) runTest(input *RunTestInput) (chan TestResult, int, error) {

	var wg sync.WaitGroup
	testCount := len(input.TestIndexes) * input.RunCount * len(input.LLMs)
	testResultChan := make(chan TestResult, testCount)

	for _, llm := range input.LLMs {
		embeddingProvider := s.embeddingProviders["nomicai"]
		llmProvider := s.llmProviders[llm.ProviderID]
		limiter := s.limiter.GetLimiter(llmProvider.Provider)
		for _, testIndex := range input.TestIndexes {
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
						testResultChan <- TestResult{Err: err}
						return
					}
					resultMessage.TestMessageID = input.Conversation.Messages[testIndex].ID
					resultMessage.ConversationID = input.Conversation.ID
					resultMessage.LLMID = llm.ID
					resultMessage.ConversationVersion = input.Conversation.SelectedVersion
					resultMessage.MessageIndex = input.Conversation.Messages[testIndex].MessageIndex

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
		Messages: openaiMessages,
		Stream:   false,
	}

	// Measure how long it takes for the first token
	startTime := time.Now()
	// Create the chat completion stream
	resp, err := llmClient.CreateChatCompletion(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get LLM response: %w", err)
	}
	if len(resp.Choices) == 0 {
		return &models.Message{
			Role:    "assistant",
			Content: "",
		}, nil
	}

	content := resp.Choices[0].Message.Content

	// Generate text embeddings
	responseEmbedding, err := embeddingClient.EmbedText([]string{content}, nomicai.Clustering)
	if err != nil {
		return nil, fmt.Errorf("failed to get text embedding: %w", err)
	}

	totalLatencyMs := int(time.Since(startTime).Milliseconds())

	message := &models.Message{
		Role:    "assistant",
		Content: content,
		Metadata: &models.MessageMetadata{
			BaseModel: models.BaseModel{
				ID: uuid.NewString(),
			},
			EndLatencyMs:     totalLatencyMs,
			OutputTokenCount: resp.Usage.CompletionTokens,
			InputTokenCount:  resp.Usage.PromptTokens,
			Embedding:        responseEmbedding.Embeddings[0],
		},
	}

	return message, nil
}

func getTestIndexes(messages []*models.Message, testMessageID string) ([]int, error) {
	evalIndexes := []int{}

	// If there is a single message to test
	if testMessageID != "" {
		for i, message := range messages {
			if message.ID == testMessageID {
				evalIndexes = append(evalIndexes, i)
				return evalIndexes, nil
			}
		}
		// if the ID can't be found
		return []int{}, fmt.Errorf("test message ID not found")
	}

	// Gets the list of all of the messages to run eval on
	for i, message := range messages {
		if message.Role != "assistant" || i == 0 {
			continue
		}
		evalIndexes = append(evalIndexes, i)
	}
	// Error if no messages found to evaluate
	if len(evalIndexes) == 0 {
		return []int{}, fmt.Errorf("no messaged found to evaluate")
	}
	return evalIndexes, nil
}

type TestManager interface {
	ExecuteTestWorkflow(input ExecuteTestInput) ([]*models.Message, error)
	GetTestList() ([]*models.Conversation, error)
}
