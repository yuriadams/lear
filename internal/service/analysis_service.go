package service

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/yuriadams/lear/internal/service/engine"
)

type IAnalysisService interface {
	StreamTextAnalysis(w http.ResponseWriter, r *http.Request, text string) error
}

type StreamedChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

type AnalysisService struct {
	AiEngine engine.AiEngine
}

func NewAnalysisService() *AnalysisService {
	return &AnalysisService{AiEngine: engine.NewSambaNovaClient()}
}

// StreamTextAnalysis handles streaming responses from the SambaNova API and sends them as SSE
func (a *AnalysisService) StreamTextAnalysis(w http.ResponseWriter, r *http.Request, text string) error {
	// Configure headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// shorten the text to don't raise a max token limit api error
	shortenedText := limitTextToTokens(text, 10000)

	prompt := fmt.Sprintf(`
	Given the following text:
	%s
	1. Identify the key characters.
	2. Detect the language.
	3. Perform sentiment analysis.
	4. Summarize the plot briefly.
	`, shortenedText)

	resp, err := a.AiEngine.StreamChat(prompt)

	if err != nil {
		return fmt.Errorf("failed to stream chat: %w", err)
	}

	if resp == nil {
		return fmt.Errorf("response stream is nil")
	}

	scanner := bufio.NewScanner(resp)
	for scanner.Scan() {
		line := scanner.Text()

		// Ignore keep-alive messages or empty lines
		if len(line) == 0 || strings.Contains(line, "data: [DONE]") {
			continue
		}

		// Remove the leading "data: " prefix
		line = strings.TrimPrefix(line, "data: ")

		var chunk StreamedChunk
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			return fmt.Errorf("failed to decode streamed chunk: %w", err)
		}

		// // Extract and stream the `delta.content`
		for _, choice := range chunk.Choices {
			content := choice.Delta.Content
			if content != "" {
				// Escape special characters to ensure valid JSON
				escapedContent := escapeJSONString(content)

				// Send the event to the client
				data := fmt.Sprintf(`event: CustomEvent
data: {"analysis": "%s"}

`, escapedContent)

				_, err := w.Write([]byte(data))
				if err != nil {
					return fmt.Errorf("failed to send event: %w", err)
				}
				w.(http.Flusher).Flush()
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stream: %w", err)
	}

	// Send close event after the stream ends
	closeMessage := `event: Close
data: Stream Ended

`
	_, err = w.Write([]byte(closeMessage))
	if err != nil {
		return fmt.Errorf("failed to send close event: %w", err)
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

func escapeJSONString(str string) string {
	escaped := strings.ReplaceAll(str, `"`, `\"`)
	escaped = strings.ReplaceAll(escaped, "\n", `\n`)
	escaped = strings.ReplaceAll(escaped, "\r", `\r`)
	escaped = strings.ReplaceAll(escaped, "\t", `\t`)
	return escaped
}
