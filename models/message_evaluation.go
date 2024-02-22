package models

type MessageEvaluation struct {
	BaseModel
	AverageSimilarity        float64
	MessageID                string
	LLM                      LLM
	LLMID                    string
	PromptID                 string
	Prompt                   Prompt
	MessageEvaluationResults []MessageEvaluationResult
}

// Add this method to compute the average similarity
func (me *MessageEvaluation) ComputeAverageSimilarity() {
	total := 0.0
	for _, result := range me.MessageEvaluationResults {
		total += result.Similarity
	}
	me.AverageSimilarity = total / float64(len(me.MessageEvaluationResults))
}

type CreateLLMEvaluationRequest struct {
	ID   string                         `path:"id"`
	Body CreateLLMEvaluationRequestBody `json:"body"`
}

type CreateLLMEvaluationRequestBody struct {
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
