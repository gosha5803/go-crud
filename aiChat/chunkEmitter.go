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
	// реслайс не освобождает базовый массива из памяти, это может быть утечкой.
	// Но лучше ли делать аллокацию?
	// при вызове flush происходит реслайс, но часть хвостового остатка удерживается в памяти.
	// но и cap уменьшается, так как указатель сдвигается вправо
	// запись осуществляется только при определённой длине, то есть в тоерии

	// в рамках одной жизни буффера, максимум, что он удерживает это весь свой хвост и до конца
	// но уже на следующей жизни он становится больше.
	// По сути назначение у нас то же
	// Бакинг аррей после каждой аллокации бесполезно растёт
	// Реалокация на каждое чтение. Так как чанк истачивает capacity почти в ноль.
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
