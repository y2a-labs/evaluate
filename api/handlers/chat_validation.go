package apihandlers

import (
	"context"
	"fmt"
	"math/rand"
	"script_validation/internal/llm"
	"script_validation/internal/nomicai"
	"sort"
	"sync"
	"time"

	"github.com/sashabaranov/go-openai"
)

func PostScriptChatValidation(ctx context.Context, input *ScriptChatInput) (*ScriptChatValidationOutput, error) {
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
	routineCount := userMessageCount * len(input.Body.Models) * input.Body.TestCount
	errCh := make(chan error, routineCount)
	resultsCh := make(chan Message, routineCount)

	// Range over the models and test count
	for _, model := range input.Body.Models {
		model := model
		for range input.Body.TestCount {

			for i, message := range input.Body.Messages {
				if message.Role != "user" {
					continue
				}

				wg.Add(1)
				go func(i int, userMessage openai.ChatCompletionMessage) {
					defer wg.Done()

					startTime := time.Now()
					// Add a random delay before sending the request
					delay := time.Duration(rand.Intn(1000)) * time.Millisecond
					time.Sleep(delay)

					// Get the LLMs Response
					llm_response, err := llm.GetLLMResponse(llm_client, input.Body.Messages[:i+1], model)
					if err != nil {
						errCh <- err
						return
					}

					similarity, err := nomicai.GetTextSimilarity(llm_response.Content, input.Body.Messages[i+1].Content)
					if err != nil {
						errCh <- err
						return
					}

					nextMessage := input.Body.Messages[i+1]

					// Capture the latency of the request
					endTime := time.Now()
					latency := endTime.Sub(startTime).Milliseconds()

					result := Message{
						Content:      nextMessage.Content,
						Role:         nextMessage.Role,
						MessageIndex: i + 1,
						TaskIndex:    input.Body.TaskIndex,
						LLMSimilarityList: map[string]LLMSimilarityList{
							model: {
								Model:        model,
								LLMResponses: []LLMSimilarity{{LLMResponse: llm_response.Content, Similarity: similarity, LatencyMs: latency}},
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

	results := make([]Message, len(input.Body.Messages))

	// Load the messages into the results
	for i, msg := range input.Body.Messages {
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

	return &ScriptChatValidationOutput{
		Body: BodyResults{
			Messages: results,
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
