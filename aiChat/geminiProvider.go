package aichat

import (
	"context"
	"iter"

	"google.golang.org/genai"
)

type GeminiProvider struct {
	client *genai.Client
}

func NewGeminiProvider(ctx context.Context, apiKey string) (*GeminiProvider, error) {
	// Тут в провайдере создаётся синглтон клиент
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})

	if err != nil {
		return nil, err
	}

	return &GeminiProvider{client: client}, nil
}

func (p *GeminiProvider) StreamResponse(
	ctx context.Context,
	req AiModelRequest,
) iter.Seq2[*AiModelResponse, error] {
	// функция итератор - принимает возвращаемые на каждой итерации значения
	//  и возвращает bool внутрь цикла сигнализируя о том не был ли цикл прерван
	return func(yield func(*AiModelResponse, error) bool) {
		// genai возвращает свой iter.Seq2 — адаптируем его к доменному типу.
		// В этом адаптере не нао передавать модель заданную с клиента,
		// тут мы как минимум всегда завязаны на модель Gemini

		stream := p.client.Models.GenerateContentStream(
			ctx,
			// TODO сделать выбор моделей???
			// TODO сделать справочник моделей
			// Динамический выбор самих иишек не знаю
			// Сделать передачу предыдущих сообщений для контекста
			// Сделать хранение истории по id пользака как-то?
			"gemini-3.6-flash",
			genai.Text(req.Prompt),
			nil,
		)

		for resp, err := range stream {
			if err != nil {
				yield(nil, err)
				return
			}

			chunk := &AiModelResponse{
				Text:    resp.Text(), // хелпер из SDK
				IsFinal: false,       // можно выставить по finishReason
			}

			if !yield(chunk, nil) {
				return // потребитель сделал break/return — прекращаем
			}
		}
	}
}
