package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"github.com/y2a-labs/evaluate/models"
	"strings"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

func (rs Resources) ProxyOpenaiEmbedding (c *fuego.ContextWithBody[openai.EmbeddingRequest]) (*openai.EmbeddingResponse, error) {
	body, err := c.Body()
	if err != nil {
		return nil, err
	}
	return rs.Service.ProxyOpenaiEmbedding(c.Context(), body)
}

func (rs Resources) ProxyOpenaiChatCompletion(c *fuego.ContextWithBody[openai.ChatCompletionRequest]) (any, error) {
	ctx, cancel := context.WithTimeout(c.Context(), 2*time.Minute)
	defer cancel()
	body, err := c.Body()
	providerId := c.Req.Header.Get("Provider-Id")
	if err != nil {
		return nil, err
	}

	var responseContent string

	if body.Stream {
		startTime := time.Now()
		stream, conversation, err := rs.Service.ProxyOpenaiStream(ctx, body, providerId)
		if err != nil {
			return nil, err
		}
		defer stream.Close()
		responseBuffer := strings.Builder{}
		firstTokenLatencyMs := 0
		for {
			resp, err := stream.Recv()
			// If the stream is done, break out of the loop
			if errors.Is(err, io.EOF) {
				_, writeErr := c.Res.Write([]byte("data:[DONE]"))
				if writeErr != nil {
					return nil, fmt.Errorf("failed to write response: %w", writeErr)
				}
				break
			}
			// If the stream is stopped early
			if err != nil {
				break
			}

			// Add the content of resp.Choices[0].Delta.Content to the response buffer
			if len(resp.Choices) > 0 {
				responseBuffer.WriteString(resp.Choices[0].Delta.Content)
				if firstTokenLatencyMs == 0 {
					firstTokenLatencyMs = int(time.Since(startTime).Milliseconds())
				}
			}

			respBytes, err := json.Marshal(resp)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response to JSON: %w", err)
			}

			var builder strings.Builder

			builder.WriteString("data: ")
			builder.Write(respBytes)
			builder.WriteString("\n\n")

			_, writeErr := c.Res.Write([]byte(builder.String()))
			if writeErr != nil {
				return nil, fmt.Errorf("failed to write response: %w", writeErr)
			}
		}

		// Get the accumulated content from the response buffer
		responseContent := responseBuffer.String()

		message := &models.Message{
			BaseModel:      models.BaseModel{ID: uuid.NewString()},
			Role:           "assistant",
			Content:        responseContent,
			ConversationID: conversation.ID,
			LLMID:          body.Model,
			MessageIndex:   len(body.Messages),
			Metadata: &models.MessageMetadata{
				BaseModel:      models.BaseModel{ID: uuid.NewString()},
				StartLatencyMs: firstTokenLatencyMs,
				EndLatencyMs:   int(time.Since(startTime).Milliseconds()),
			},
		}
		conversation.Messages = append(conversation.Messages, message)

		tx := rs.Service.Db.Save(conversation)
		if tx.Error != nil {
			return nil, tx.Error
		}

	} else {
		startTime := time.Now()
		response, conversation, err := rs.Service.ProxyOpenaiChat(c.Context(), body, providerId)
		if err != nil {
			return nil, err
		}
		responseContent = response.Choices[0].Message.Content
		message := &models.Message{
			BaseModel:      models.BaseModel{ID: uuid.NewString()},
			Role:           "assistant",
			Content:        responseContent,
			ConversationID: conversation.ID,
			LLMID:          body.Model,
			MessageIndex:   len(body.Messages),
			Metadata: &models.MessageMetadata{
				BaseModel:    models.BaseModel{ID: uuid.NewString()},
				EndLatencyMs: int(time.Since(startTime).Milliseconds()),
			},
		}
		conversation.Messages = append(conversation.Messages, message)
		tx := rs.Service.Db.Save(conversation)
		if tx.Error != nil {
			return nil, tx.Error
		}
		return response, nil
	}

	return responseContent, nil
}
