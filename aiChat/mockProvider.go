package aichat

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"os"
	"time"
)

type Stream struct {
	Messages []Message `json:"messages"`
}

type Message struct {
	Message *Chunk `json:"message,omitempty"`
	Done    *Done  `json:"done,omitempty"`
}

type Chunk struct {
	Text string `json:"text"`
}

type Done struct {
	Reason string `json:"reason"`
}

type MockProvider struct {
	client   string
	filePath string
}

func NewMockProvider(filePath string) (*MockProvider, error) {
	return &MockProvider{filePath: filePath}, nil
}

func (p *MockProvider) StreamResponse(context.Context, AiModelRequest) iter.Seq2[*AiModelResponse, error] {
	// TODO можно использовать для тестов
	file, err := os.ReadFile(p.filePath)

	if err != nil {
		return func(yield func(*AiModelResponse, error) bool) {
			yield(&AiModelResponse{}, fmt.Errorf("mock provider: stream response: %w", ErrReadingJSONFile))
			return
		}
	}

	var s Stream

	if err := json.Unmarshal(file, &s); err != nil {
		return func(yield func(*AiModelResponse, error) bool) {
			yield(&AiModelResponse{}, fmt.Errorf("mock provider: stream response: %w", ErrParsingJSON))
			return
		}
	}

	return func(yield func(*AiModelResponse, error) bool) {
		for _, m := range s.Messages {

			if m.Message == nil {
				if m.Done != nil {
					yield(&AiModelResponse{IsFinal: true}, nil)
				}
				continue
			}

			time.Sleep(100 * time.Millisecond)

			resp := &AiModelResponse{Text: m.Message.Text}
			if !yield(resp, nil) {
				return
			}
		}
	}
}
