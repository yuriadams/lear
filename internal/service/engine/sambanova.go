package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const SambaNovaChatAPIURL = "https://api.sambanova.ai/v1/chat/completions"

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type AiEngine interface {
	StreamChat(prompt string) (io.ReadCloser, error)
}

type DefaultSambaNovaClient struct {
	apiURL    string
	authToken string
}

func NewSambaNovaClient() *DefaultSambaNovaClient {
	return &DefaultSambaNovaClient{
		apiURL:    SambaNovaChatAPIURL,
		authToken: os.Getenv("AI_API_TOKEN"),
	}
}

func (c *DefaultSambaNovaClient) StreamChat(prompt string) (io.ReadCloser, error) {
	chatRequest := ChatRequest{
		Model: "Meta-Llama-3.1-70B-Instruct",
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream: true,
	}

	reqBody, err := json.Marshal(chatRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create request body: %w", err)
	}

	req, err := http.NewRequest("POST", c.apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.authToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to analyze text, status %d: %s", resp.StatusCode, string(body))
	}

	return resp.Body, nil
}
