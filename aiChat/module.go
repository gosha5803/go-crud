package aichat

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gosha5803/go-crud/initializers"
)

// TODo по сути все контрллеры должны имплементировать AssignRoutes
// Либо принимать извне конфиг?
type IAiChatController interface {
	AssignRoutes()
}

type AiChatModule struct {
	controller IAiChatController
}

func NewAiChatModule(ctx context.Context, engine *gin.Engine, aiConfig initializers.AiConfig) (*AiChatModule, error) {
	// TODO простота использования, просто тут поменяли два параметра в фабрике и переключились на другую модель
	aiClient, err := NewAiProvider(ctx, ProviderConfig{Kind: ProviderGemini, APIKey: aiConfig.GeminiApiKey})

	if err != nil {
		return nil, fmt.Errorf("new ai chat module: %w", err)
	}

	service := NewAiChatService(aiClient, 6, 25*time.Millisecond)
	controller := NewAiChatController(service, engine)

	return &AiChatModule{controller: controller}, nil

}

func (module *AiChatModule) Init() {
	module.controller.AssignRoutes()
}
