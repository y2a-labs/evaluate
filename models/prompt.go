package models

type CreatePromptRequest struct {
	ID string `path:"id"`
}

type FindOrCreatePromptRequest struct {
	Body FindOrCreatePromptRequestBody `json:"body"`
}

type FindOrCreatePromptRequestBody struct {
	Content string `json:"content" example:"Hello, world!"`
}

type FindOrCreatePromptResponse struct {
	Body FindOrCreatePromptResponseBody `json:"body"`
}

type FindOrCreatePromptResponseBody struct {
	Content string `json:"content"`
	BaseModel
}
