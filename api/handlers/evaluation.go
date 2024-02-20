package apihandlers

import (
	"context"
	"fmt"
	database "script_validation"
	"script_validation/internal/llm"
	"script_validation/internal/nomicai"
	"script_validation/models"
	"sort"
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
	Messages []LLMEvaluationOuputMessages `json:"messages"`
}

type LLMEvaluationOuputMessages struct {
	models.Message
	Results []LLMEvaluationOuputResults `json:"results"`
}

type LLMEvaluationOuputResults struct {
	Model       string  `json:"model"`
	LLMResponse string  `json:"llm_response"`
	Similarity  float64 `json:"similarity"`
	LatencyMs   int64   `json:"latency_ms"`
}

type ProcessMessage struct {
	i            int
	UserMessage  models.ChatMessage
	Conversation models.Conversation
	LLM          models.LLM
	PromptID     string
	ChatMessages []models.ChatMessage
	LLMClient    *openai.Client
	ResultsCh    chan<- models.LLMEvaluation
	ErrCh        chan<- error
}

func processMessage(p ProcessMessage) {
	startTime := time.Now()

	// Get the LLMs Response
	llm_response, err := llm.GetLLMResponse(p.LLMClient, p.ChatMessages[:p.i+1], p.LLM.Name)
	if err != nil {
		p.ErrCh <- err
		return
	}

	// Get the similarity
	similarity, err := nomicai.GetTextSimilarity(llm_response.Content, p.Conversation.Messages[p.i+1].Content)
	if err != nil {
		p.ErrCh <- err
		return
	}

	nextMessage := p.Conversation.Messages[p.i+1]

	// Capture the latency of the request
	endTime := time.Now()
	latency := endTime.Sub(startTime).Milliseconds()
	fmt.Println("Latency: ", latency, similarity, nextMessage.ID)
	result := models.LLMEvaluation{
		Content:    llm_response.Content,
		LLM:        p.LLM,
		LLMID: 	p.LLM.ID,
		Message:    nextMessage,
		MessageID: nextMessage.ID,
		PromptID:  p.PromptID,
		LatencyMs:  uint(latency),
		Similarity: similarity,
	}

	p.ResultsCh <- result
}

func SetSystemPrompt(messages *[]models.ChatMessage, prompt string) error {
	if len(*messages) == 0 {
		return fmt.Errorf("err: no messages in the conversation")
	}
	if (*messages)[0].Role != "system" {
		// In the first index set the first message to the system message
		*messages = append([]models.ChatMessage{{Role: "system", Content: prompt}}, *messages...)
	} else {
		// If the first message is a system message, update the content
		(*messages)[0].Content = prompt
	}
	return nil
}

func FindOrCreatePrompt(text string) (*models.Prompt, error) {
	prompt := models.Prompt{Content: text}
	// Finds the prompt in the database
	result := database.DB.First(&prompt, "content = ?", text)

	// Creates it if it doesn't exist
	if result.Error != nil {
		fmt.Println("Not found, creating a new prompt")
		r := database.DB.Create(&prompt)
		if r.Error != nil {
			return nil, r.Error
		}
	}
	return &prompt, nil
}

func GetLLMs(model_names []string) ([]models.LLM, error) {
	llms := make([]models.LLM, len(model_names))
	r := database.DB.Find(&llms, "name IN ?", model_names)
	if r.Error != nil {
		return nil, r.Error
	}
	return llms, nil
}

func CreateLLMEvaluation(ctx context.Context, input *LLMEvaluationInput) (*LLMEvaluationOutput, error) {

	// Get the conversation
	conversation, err := GetConversation(ctx, &models.GetConversationInput{Id: input.ConversationId})
	if err != nil {
		return nil, err
	}

	// Get the prompt ID
	prompt, err := FindOrCreatePrompt(input.Body.Prompt)

	if err != nil || prompt.ID == ""{
		return nil, err
	}

	// Set the system prompt
	SetSystemPrompt(&conversation.Body.Messages, input.Body.Prompt)

	// Get the models
	llms, err := GetLLMs(input.Body.Models)
	if err != nil || len(llms) == 0 {
		return nil, err
	}
	fmt.Println("LLMs: ", llms)

	// Create the LLM Client
	llm_client := llm.GetLLMClient(map[string]string{})

	// Get the number of messages where the role is user
	userMessageCount := 0
	for _, message := range conversation.Body.Messages {
		if message.Role == "user" {
			userMessageCount++
		}
	}

	var wg sync.WaitGroup
	routineCount := userMessageCount * len(input.Body.Models) * input.Body.TestCount
	errCh := make(chan error, routineCount)
	resultsCh := make(chan models.LLMEvaluation, routineCount)

	// Range over the models and test count
	for _, llm := range llms {
		llm := llm
		for range input.Body.TestCount {

			for i, message := range conversation.Body.Messages {
				// Only generate a response on user messages
				if message.Role != "user" {
					continue
				}

				wg.Add(1)

				go processMessage(ProcessMessage{
					i:           i,
					UserMessage: message,
					Conversation: models.Conversation{
						Name:      conversation.Body.Name,
						BaseModel: conversation.Body.BaseModel,
					},
					LLM:        llm,
					PromptID:  prompt.ID,
					ChatMessages: conversation.Body.Messages,
					LLMClient:    llm_client,
					ResultsCh:    resultsCh,
					ErrCh:        errCh,
				})
			}
		}
	}

	wg.Wait()
	close(errCh)
	close(resultsCh)

	for err := range errCh {
		if err != nil {
			return &LLMEvaluationOutput{}, err
		}
	}

	evaluationResults := make([]models.LLMEvaluation, routineCount)

	// Load the messages into the results
	for result := range resultsCh {
		evaluationResults = append(evaluationResults, result)
	}

	fmt.Println(evaluationResults)

	/*
		results := make([]Message, len(conversation.Body.Messages))

		// Load the messages into the results
		for i, msg := range conversation.Body.Messages {
			results[i] = Message{
				Content:      msg.Content,
				Role:         msg.Role,
				MessageIndex: i,
			}
		}

		// Load the test results into the results
		for result := range resultsCh {
			i := result.MessageIndex
			noLLMList := results[i].LLMSimilarityList == nil || len(results[i].LLMSimilarityList) == 0
			if noLLMList {
				results[i].LLMSimilarityList = result.LLMSimilarityList
			} else {
				for model, llmSimilarity := range result.LLMSimilarityList {
					temp := results[i].LLMSimilarityList[model]
					temp.Model = model
					temp.LLMResponses = append(temp.LLMResponses, llmSimilarity.LLMResponses...)
					results[i].LLMSimilarityList[model] = temp
				}
			}
		}

		// Calculates the average similarity
		for _, result := range results {
			if result.Role != "assistant" {
				continue
			}
			result.SetAverageSimilarity()
			result.SortModelResults()
		}
	*/

	return &LLMEvaluationOutput{
		Body: LLMEvaluationOuputBody{
			Messages: []LLMEvaluationOuputMessages{},
		},
	}, nil
}

func (message *Message) SortModelResults() {
	for model, value := range message.LLMSimilarityList {
		// Sort the LLMResponses by similarity
		sort.Slice(value.LLMResponses, func(i, j int) bool {
			return value.LLMResponses[i].Similarity > value.LLMResponses[j].Similarity
		})
		message.LLMSimilarityList[model] = value
	}

}

func (message *Message) SetAverageSimilarity() {
	for model, value := range message.LLMSimilarityList {
		averageSimilarity := 0.0
		for _, llmSimilarity := range value.LLMResponses {
			averageSimilarity += llmSimilarity.Similarity
		}

		averageSimilarity = averageSimilarity / float64(len(value.LLMResponses))
		value.AverageSimilarity = averageSimilarity
		temp := message.LLMSimilarityList[model]
		temp.AverageSimilarity = value.AverageSimilarity
		message.LLMSimilarityList[model] = temp
	}
}
