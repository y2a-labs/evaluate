package models

type Prompt struct {
	BaseModel
	Content      string `json:"content"`
	AgentID      string `json:"agent_id"`
	BasePromptID string `json:"base_prompt_id"`
	Version      int    `json:"version"`
}

type PromptCreate struct {
	// TODO add ressources
	ID string `json:"id"`
}

type PromptUpdate struct {
	// TODO add ressources
	ID      string `json:"id"`
	Content string `json:"content"`
}
