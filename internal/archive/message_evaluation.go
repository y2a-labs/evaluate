package archive

/*
import "fmt"

type MessageEvaluation struct {
	AverageSimilarity        float64
	MessageID                string
	LLM                      LLM
	LLMID                    string
	Prompt                   string
	MessageEvaluationResults []*MessageEvaluationResult `gorm:"constraint:OnDelete:CASCADE;"`
}

// Add this method to compute the average similarity
func (me *MessageEvaluation) ComputeAverageSimilarity() error {
	if len(me.MessageEvaluationResults) == 0 {
		fmt.Println("no results to compute average similarity")
		return fmt.Errorf("no results to compute average similarity")
	}
	total := 0.0
	for _, result := range me.MessageEvaluationResults {
		total += result.Similarity
	}
	me.AverageSimilarity = total / float64(len(me.MessageEvaluationResults))
	return nil
}

type CreateLLMEvaluationRequest struct {
	ID   string                         `path:"id"`
	Body CreateLLMEvaluationRequestBody `json:"body"`
}

type GroupedMessageEvalResults struct {
	AverageLatencyMs  int                        `json:"average_latency_ms"`
	Content           string                     `json:"content"`
	Count             int                        `json:"count"`
	AverageSimilarity float64                    `json:"average_similarity"`
	Results           []*MessageEvaluationResult `json:"results"`
}

func (eval *MessageEvaluation) GroupResults() map[string]*GroupedMessageEvalResults {
	grouped := make(map[string]*GroupedMessageEvalResults)

	for _, result := range eval.MessageEvaluationResults {
		if _, ok := grouped[result.Content]; !ok {
			grouped[result.Content] = &GroupedMessageEvalResults{
				Results: make([]*MessageEvaluationResult, 0),
			}
		}

		group := grouped[result.Content]
		group.Results = append(group.Results, result)
		group.Count++
		group.Content = result.Content
		group.AverageLatencyMs += result.LatencyMs   // Assuming MessageEvaluationResult has a LatencyMs field
		group.AverageSimilarity += result.Similarity // Assuming MessageEvaluationResult has a Similarity field
	}

	// Calculate averages
	for _, group := range grouped {
		group.AverageLatencyMs /= group.Count
		group.AverageSimilarity /= float64(group.Count)
		group.Count = group.Count / 2
	}

	return grouped
}

type CreateLLMEvaluationRequestBody struct {
	ID        string                       `json:"id" path:"id"`
	Models    []string                     `json:"models" default:"openchat/openchat-7b" required:"true"`
	TestCount int                          `json:"test_count" minimum:"1" maximum:"10" default:"1" required:"true"`
	Prompt    string                       `json:"prompt" example:"You are a helpful assistant."`
	Messages  []CreateLLMEvaluationMessage `json:"messages,omitempty"`
}

type CreateLLMEvaluationMessage struct {
	ID     string `json:"id"`
	Prompt string `json:"prompt"`
}

type CreateLLMEvaluationResponse struct {
	Body CreateLLMEvaluationResponseBody `json:"body"`
}

type CreateLLMEvaluationResponseBody struct {
	Messages *[]Message `json:"messages"`
}

type CreateLLMEvaluationResponseValidation struct {
	Body struct {
		ID string `json:"id"`
	}
}
*/