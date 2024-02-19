package models

type LLMEvaluation struct {
	Content   string `db:"content" json:"content"`
	ModelId   string   `db:"model_id" json:"model_id"`
	MessageId string   `db:"message_id" json:"message_id"`
	LatencyMs uint   `db:"latency_ms" json:"latency_ms"`
	PromptId  string   `db:"prompt_id" json:"prompt_id"`
}

type LLMEvaluationModel struct {
	BaseModel
	LLMEvaluation
}
