package service_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yuriadams/lear/internal/service"
)

type MockAiEngine struct {
	mock.Mock
}

func (m *MockAiEngine) StreamChat(prompt string) (io.ReadCloser, error) {
	args := m.Called(prompt)
	resp, _ := args.Get(0).(io.ReadCloser)
	return resp, args.Error(1)
}

func TestStreamTextAnalysis_Success(t *testing.T) {
	mockAiEngine := new(MockAiEngine)

	mockResponse := `
data: {"choices":[{"delta":{"content":"Character: John"}}]}
data: {"choices":[{"delta":{"content":"Language: English"}}]}
data: [DONE]
`
	mockAiEngine.On("StreamChat", mock.Anything).Run(func(args mock.Arguments) {
		fmt.Println("StreamChat called with prompt:", args.String(0))
	}).Return(io.NopCloser(bytes.NewBufferString(mockResponse)), nil)

	service := &service.AnalysisService{AiEngine: mockAiEngine}

	req, err := http.NewRequest(http.MethodGet, "/analyze", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	err = service.StreamTextAnalysis(rr, req, "This is a test text.")
	assert.NoError(t, err)

	expectedResponse := `event: CustomEvent
data: {"analysis": "Character: John"}

event: CustomEvent
data: {"analysis": "Language: English"}

event: Close
data: Stream Ended

`
	assert.Equal(t, expectedResponse, rr.Body.String())

	mockAiEngine.AssertExpectations(t)
}

func TestStreamTextAnalysis_FailedStreamChat(t *testing.T) {
	mockAiEngine := new(MockAiEngine)

	mockAiEngine.On("StreamChat", mock.Anything).Return(nil, errors.New("stream error"))

	service := &service.AnalysisService{AiEngine: mockAiEngine}

	req, err := http.NewRequest(http.MethodGet, "/analyze", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	err = service.StreamTextAnalysis(rr, req, "This is a test text.")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to stream chat")

	mockAiEngine.AssertExpectations(t)
}
