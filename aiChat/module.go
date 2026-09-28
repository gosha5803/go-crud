package aichat

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// TODo по сути все контрллеры должны имплементировать AssignRoutes
// Либо принимать извне конфиг?
type IAiChatController interface {
	askModel(c *gin.Context)
	AssignRoutes()
}

type AiChatModule struct {
	controller IAiChatController
}

func NewAiChatModule(ctx context.Context, engine *gin.Engine) (*AiChatModule, error) {
	// TODO тут нужен конфиг сервис
	apiKey := os.Getenv("GEMINI_API_KEY")

	// TODO простота использования, просто тут поменяли два параметра в фабрике и переключились на другую модель
	aiClient, err := NewAiProvider(ctx, ProviderConfig{Kind: ProviderGemini, APIKey: apiKey})

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
