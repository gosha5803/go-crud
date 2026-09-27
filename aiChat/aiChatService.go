package aichat

import (
	"context"
	"iter"
	"time"
)

type AiChatService struct {
	aiProvider   IAiModelProvider
	chunkSize    int // целевой размер одного чанка в рунах
	flushEvery   time.Duration
	chunkEmitter ChunkEmitter
}

type IAiModelProvider interface {
	// Name — для логов и выбора в фабрике.
	// Name() string

	// StreamResponse — единая точка входа для всех моделей.
	// Возвращает итератор: (чанк, nil) на каждом куске, (nil, err) при ошибке.
	StreamResponse(context.Context, AiModelRequest) iter.Seq2[*AiModelResponse, error]
}

func NewAiChatService(aiProvider IAiModelProvider, chunkSize int, flushEvery time.Duration) *AiChatService {
	if chunkSize <= 0 {
		chunkSize = 8
	}
	return &AiChatService{aiProvider: aiProvider, chunkSize: chunkSize, flushEvery: flushEvery}
}

// можно работать с каналом вместо итератора
// решить с моделью, может что-то платное купить
func (s *AiChatService) AskModelStream(ctx context.Context, prompt string) iter.Seq2[*AiModelResponse, error] {
	// TODO Понять как работает
	// TODO выводить бы пользователю остаток запросов
	return func(yield func(*AiModelResponse, error) bool) {

		// Имеет ли смысл вынести во внешний метод? Переписать/покрыть тестами
		// Может вынести буфер и время в состояние структур - класса?
		emitter := NewChunkEmitter(s.chunkSize, s.flushEvery, yield)

		for resp, err := range s.aiProvider.StreamResponse(ctx, AiModelRequest{Prompt: prompt}) {
			// если ошибка то надо очистить буфер и вернуть ошибку
			if err != nil {
				if !emitter.flushAll(true) {
					return
				}

				yield(nil, err)
				return
			}

			emitter.write(resp.Text)

			for emitter.isBufferReady() {
				// далее надо проверить готов ли буфер и пока он готов, флешить в SSE чанк и чистить буффер
				if !emitter.flush(false) {
					return
				}
				// также в случве ошибки флеша ретурн
			}

			// если isFinal, то flushAll вызываем, который сам елдит
			if resp.IsFinal {
				if !emitter.flushAll(true) {
					return
				}

				yield(&AiModelResponse{IsFinal: true}, nil)
				return
			}

		}

		// если isFinal не пришёл, но мы вывалились из цикла., то flushAll вызываем, который сам елдит
		emitter.flushAll(true)
	}
}
