package models

type MessageEvaluation struct {
	BaseModel
	MessageID                string
	AverageSimilarity        float64
	LLM                      LLM
	LLMID                    string
	Prompt                   Prompt
	PromptID                 string
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
	ID   string `path:"id"`
	Body struct {
		Models    []string `json:"models" default:"openchat/openchat-7b"`
		TestCount int      `json:"test_count" minimum:"1" maximum:"10" default:"1"`
		Prompt    string   `json:"prompt" default:"You are a helpful assistant."`
	}
}

type CreateLLMEvaluationResponse struct {
	Body CreateLLMEvaluationResponseBody `json:"body"`
}

type CreateLLMEvaluationResponseBody struct {
	Results *[]MessageEvaluation `json:"results"`
}



type CreateLLMEvaluationResponseValidation struct {
	Body struct {
		ID                string  `json:"id"`
	}
}

/*
		MessageID         string  `json:"message_id"`
		AverageSimilarity float64 `json:"average_similarity"`
		LLM               struct {
			Name string `json:"name"`
		} `json:"llm"`
		Prompt struct {
			Content string `json:"content"`
		} `json:"prompt"`
		Results []struct {
			ID                  string  `json:"id"`
			Content             string  `json:"content"`
			LatencyMs           int     `json:"latency_ms"`
			Similarity          float64 `json:"similarity"`
			MessageEvaluationID string  `json:"message_evaluation_id"`
		} `json:"results"`
		*/