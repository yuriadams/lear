package service

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const SambaNovaChatAPIURL = "https://api.sambanova.ai/v1/chat/completions"

type ChatMessage struct {
	Role    string `json:"role"`    // Role can be "user", "assistant", or "system"
	Content string `json:"content"` // Content of the message
}

type ChatRequest struct {
	Model    string        `json:"model"`    // Model name (e.g., "Meta-Llama-3.1-70B-Instruct")
	Messages []ChatMessage `json:"messages"` // List of conversation messages
	Stream   bool          `json:"stream"`   // Enable streaming responses
}

type StreamedChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

// StreamTextAnalysis handles streaming responses from the SambaNova API and sends them as SSE
func StreamTextAnalysis(w http.ResponseWriter, r *http.Request, text string) error {
	// Configure headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Create a prompt and shorten the text if necessary
	shortenedText := limitTextToTokens(text, 10000)

	prompt := fmt.Sprintf(`
	Given the following text:
	%s
	1. Identify the key characters.
	2. Detect the language.
	3. Perform sentiment analysis.
	4. Summarize the plot briefly.
	`, shortenedText)

	// Prepare the request payload
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

	// Encode the request payload
	reqBody, err := json.Marshal(chatRequest)
	if err != nil {
		return fmt.Errorf("failed to create request body: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", SambaNovaChatAPIURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("AI_API_TOKEN"))
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to analyze text, status %d: %s", resp.StatusCode, string(body))
	}

	// Process the streaming response line by line
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// Ignore keep-alive messages or empty lines
		if len(line) == 0 || strings.Contains(line, "data: [DONE]") {
			continue
		}

		// Remove the leading "data: " prefix
		line = strings.TrimPrefix(line, "data: ")

		// Decode the JSON chunk
		var chunk StreamedChunk
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			return fmt.Errorf("failed to decode streamed chunk: %w", err)
		}

		// Extract and stream the `delta.content`
		for _, choice := range chunk.Choices {
			content := choice.Delta.Content
			if content != "" {
				fmt.Fprintf(w, "%s", content)
				w.(http.Flusher).Flush() // Ensure immediate sending of the data
			}
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stream: %w", err)
	}

	return nil
}

// limitTextToTokens trims the text to a maximum number of words
func limitTextToTokens(text string, maxWords int) string {
	words := strings.Fields(text)
	if len(words) > maxWords {
		return strings.Join(words[:maxWords], " ")
	}
	return text
}
