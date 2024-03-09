package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/sashabaranov/go-openai"
)

func (rs Resources) ProxyOpenai(c *fuego.ContextWithBody[openai.ChatCompletionRequest]) (any, error) {
	body, err := c.Body()
	agentId := c.Req.Header.Get("AGENT-ID")
	fmt.Println(agentId)
	if err != nil {
		return nil, err
	}

	var responseContent string

	if body.Stream {
		stream, err := rs.Service.ProxyOpenaiStream(c.Context(), body, agentId)
		if err != nil {
			return nil, err
		}
		responseBuffer := strings.Builder{}
		for {
			resp, err := stream.Recv()
			// If the stream is done, break out of the loop
			if errors.Is(err, io.EOF) {
				_, writeErr := c.Res.Write([]byte("data: [DONE]\n\n"))
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
			}

			respBytes, err := json.Marshal(resp)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response to JSON: %w", err)
			}

			data := "data: " + string(respBytes) + "\n\n"
			_, writeErr := c.Res.Write([]byte(data))
			if writeErr != nil {
				return nil, fmt.Errorf("failed to write response: %w", writeErr)
			}
		}

		// Get the accumulated content from the response buffer
		responseContent := responseBuffer.String()

		fmt.Println("resp: ", responseContent)
	} else {
		response, err := rs.Service.ProxyOpenaiChat(c.Context(), body, agentId)
		if err != nil {
			return nil, err
		}
		responseContent = response.Choices[0].Message.Content
	}
	fmt.Println(responseContent)

	return responseContent, nil
}
