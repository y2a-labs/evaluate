package models

type LLMEvaluation struct {
	BaseModel
	Content    string
	LLM        LLM 
	LLMID      string
	Message    Message 
	MessageID  string
	Prompt     Prompt 
	PromptID   string
	LatencyMs  uint
	Similarity float64
}
