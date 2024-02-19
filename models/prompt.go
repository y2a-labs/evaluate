package models

type Prompt struct {
	Content string `db:"content" json:"content"`
}

type PromptModel struct {
	BaseModel
	Prompt
}
