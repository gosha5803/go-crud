package aichat

import (
	"context"
	"iter"
	"net/http"

	"github.com/gin-gonic/gin"
)

type IAiChatService interface {
	AskModelStream(ctx context.Context, prompt string) iter.Seq2[*AiModelResponse, error]
}

type ChatController struct {
	service IAiChatService
	g       *gin.Engine
}

func NewAiChatController(service IAiChatService, g *gin.Engine) *ChatController {
	return &ChatController{service: service, g: g}
}

func (controller *ChatController) askModel(c *gin.Context) {
	var dto AiModelRequest

	if err := c.Bind(&dto); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(500, AiErrorDTO{Message: "streaming not supported"})
		return
	}

	stream := controller.service.AskModelStream(c.Request.Context(), dto.Prompt)

	var chunks []string // ← копилка

	for resp, err := range stream {
		if err != nil {
			c.SSEvent("error", gin.H{"error": err.Error()})
			flusher.Flush()
			return
		}

		chunks = append(chunks, resp.Text) // ← собираем

		c.SSEvent("message", gin.H{"text": resp.Text})
		flusher.Flush()
	}

	c.SSEvent("done", AiDoneDTO{Reason: "final"})
	flusher.Flush()

}

func (controller *ChatController) AssignRoutes() {
	controller.g.POST("/chat", controller.askModel)
}

// Придумать моковый DEEPSEEK PROVIDER и попробовать его внедрить, подумать над тем как лучше внедрять апи ключи. ConfigService?
//
