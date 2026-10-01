package aichat

import "testing"

func TestChunkEmitter_CapacityBehavior(t *testing.T) {
	var received []string

	// Проверить тест

	yield := func(resp *AiModelResponse, err error) bool {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		received = append(received, resp.Text)
		return true
	}

	// flushEvery=0 — чтобы тест не спал реальное время
	emitter := NewChunkEmitter(8, 0, yield)

	t.Logf("start:len=%d cap=%d", len(emitter.buffer), cap(emitter.buffer))

	chunks := []string{
		"привет как дела ",
		"у меня всё хорошо спасибо ",
		"а у тебя как дела дружище ",
		"тоже неплохо, работаю над тестами ",
	}

	reallocCount := 0
	prevCap := cap(emitter.buffer)

	for i, c := range chunks {
		emitter.write(c)
		t.Logf("write #%d:         len=%d cap=%d", i, len(emitter.buffer), cap(emitter.buffer))

		for emitter.isBufferReady() {
			beforeCap := cap(emitter.buffer)
			emitter.flush(false)
			afterCap := cap(emitter.buffer)
			t.Logf("  flush:          len=%d cap=%d->%d", len(emitter.buffer), beforeCap, afterCap)
		}

		// если cap после серии flush вырос относительно предыдущего замера — была переаллокация
		if cap(emitter.buffer) > prevCap {
			reallocCount++
		}
		prevCap = cap(emitter.buffer)
	}

	emitter.flushAll(true)
	t.Logf("end:              len=%d cap=%d", len(emitter.buffer), cap(emitter.buffer))
	t.Logf("reallocations observed: %d", reallocCount)
	t.Logf("chunks received: %d -> %v", len(received), received)
}
