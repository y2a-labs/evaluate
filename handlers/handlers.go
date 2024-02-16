package handlers

import "github.com/sashabaranov/go-openai"

type Message struct {
	Choice  openai.ChatCompletionMessage `json:"choice"`
	Results []results                    `json:"results"`
}
