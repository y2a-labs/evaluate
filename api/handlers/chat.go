package apihandlers

import (
	"github.com/sashabaranov/go-openai"
)

type Task struct {
	Intro     string `json:"intro" yaml:"intro" validate:"required"`
	Goal      string `json:"goal" yaml:"goal" validate:"required"`
	Condition string `json:"condition" yaml:"condition" validate:"required"`
}

type Script struct {
	SystemPromptString string `json:"system_prompt_string" yaml:"system_prompt_string" validate:"required"`
	Character          string `json:"character" yaml:"character" validate:"required"`
	Task               []Task `json:"tasks" yaml:"tasks" validate:"required"`
}

type ScriptChatInput struct {
	Body ScriptChatInputBody
}

type ScriptChatInputBody struct {
	TestCount int                            `json:"test_count" yaml:"test_count"`
	Script    Script                         `json:"script" yaml:"script" validate:"required"`
	Messages  []openai.ChatCompletionMessage `json:"messages" yaml:"messages"`
	TaskIndex int                            `json:"script_index" yaml:"script_index" validate:"required"`
	Models    []string                       `json:"models" yaml:"models" validate:"required"`
}

type ScriptChatOutput struct {
	Body ScriptChatOutputBody
}

type ScriptChatOutputBody struct {
	Messages  []openai.ChatCompletionMessage `json:"messages"`
	TaskIndex int
}

type ScriptChatValidationOutput struct {
	Body BodyResults `json:"body"`
}

type BodyResults struct {
	Messages []Message `json:"messages"`
}

type Message struct {
	Content           string                       `json:"content"`
	Role              string                       `json:"role"`
	TaskIndex         int                          `json:"task_index"`
	MessageIndex      int                          `json:"message_index"`
	LLMSimilarityList map[string]LLMSimilarityList `json:"llm_results"`
}

type LLMSimilarityList struct {
	Model             string          `json:"model"`
	AverageSimilarity float64         `json:"average_similarity"`
	LLMResponses      []LLMSimilarity `json:"responses"`
}

type LLMSimilarity struct {
	LLMResponse string  `json:"llm_response"`
	Similarity  float64 `json:"similarity"`
	LatencyMs   int64   `json:"latency_ms"`
}
