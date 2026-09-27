package aichat

import "time"

type ChunkEmitter struct {
	chunkSize  int
	buffer     []rune
	lastYield  time.Time
	flushEvery time.Duration
	yield      func(*AiModelResponse, error) bool
}

func NewChunkEmitter(chunkSize int, flushEvery time.Duration, yield func(*AiModelResponse, error) bool) *ChunkEmitter {
	stableChunkSize := chunkSize
	if stableChunkSize < 0 {
		stableChunkSize = 8
	}

	return &ChunkEmitter{chunkSize: stableChunkSize, flushEvery: flushEvery, yield: yield}
}

func (emitter *ChunkEmitter) flush(force bool) bool {
	if len(emitter.buffer) == 0 {
		return true
	}

	if !emitter.lastYield.IsZero() && !force {
		if wait := emitter.flushEvery - time.Since(emitter.lastYield); wait > 0 {
			time.Sleep(wait)
		}
	}

	size := emitter.chunkSize

	if len(emitter.buffer) < size {
		size = len(emitter.buffer)
	}

	response := AiModelResponse{Text: string(emitter.buffer[:size])}
	emitter.buffer = emitter.buffer[size:]
	emitter.lastYield = time.Now()

	return emitter.yield(&response, nil)
}

func (emitter *ChunkEmitter) isBufferReady() bool {
	return len(emitter.buffer) >= emitter.chunkSize
}

func (emitter *ChunkEmitter) flushAll(force bool) bool {
	for len(emitter.buffer) > 0 {
		if !emitter.flush(force) {
			return false
		}
	}

	return true
}

func (emitter *ChunkEmitter) write(text string) {
	emitter.buffer = append(emitter.buffer, []rune(text)...)
}

// Бул возвращаем всё также, если цикл прервётся извне
// Доделать остальные методы
