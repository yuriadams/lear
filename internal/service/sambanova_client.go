package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
)

const SambaNovaAPIURL = "https://api.sambanova.ai/analyze"

type AnalysisRequest struct {
	Text string `json:"text"`
}

func RequestCompletion(text string) (io.ReadCloser, error) {
	requestBody, err := json.Marshal(AnalysisRequest{Text: text})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", SambaNovaAPIURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("AI_API_TOKEN"))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to analyze text with SambaNova")
	}

	return resp.Body, nil
}
