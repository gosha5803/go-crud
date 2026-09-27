package aichat

import (
	"context"
	"fmt"
)

type ProviderKind string

const (
	ProviderGemini ProviderKind = "gemini"
	ProviderMock   ProviderKind = "mock"
)

type ProviderConfig struct {
	Kind   ProviderKind
	APIKey string
}

func NewAiProvider(ctx context.Context, cfg ProviderConfig) (IAiModelProvider, error) {
	switch cfg.Kind {
	case ProviderGemini:
		return NewGeminiProvider(ctx, cfg.APIKey)
	case ProviderMock:
		return NewMockProvider()
	default:
		return nil, fmt.Errorf("unknown provider: %s", cfg.Kind)
	}
}
