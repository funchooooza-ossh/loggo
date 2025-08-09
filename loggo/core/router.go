package core

import (
	"context"
	"sync"
)

// RouteProcessor связывает форматтер и writer, обрабатывает лог-события асинхронно.
type RouteProcessor struct {
	Formatter      FormatProcessor
	Writer         WriteProcessor
	LevelThreshold LogLevel

	queue  chan LogRecord
	closed bool
	mu     sync.RWMutex
}

// NewRouteProcessor создаёт маршрутизатор логов с указанным форматтером и writer'ом.
func NewRouteProcessor(formatter FormatProcessor, writer WriteProcessor, level LogLevel) *RouteProcessor {
	return &RouteProcessor{
		Formatter:      formatter,
		Writer:         writer,
		LevelThreshold: level,
		queue:          make(chan LogRecord, 1024),
	}
}

// ShouldLog проверяет, подходит ли уровень события для этого роута.
func (r *RouteProcessor) ShouldLog(level LogLevel) bool {
	return level >= r.LevelThreshold
}

// Enqueue отправляет событие в очередь логирования (если не закрыто).
func (r *RouteProcessor) Enqueue(record LogRecord) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return
	}

	select {
	case r.queue <- record:
	default:
		// очередь переполнена — можно добавить метрику/лог
	}
}

// Start запускает обработку очереди в отдельной горутине.
func (r *RouteProcessor) Start(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer r.drainQueue()

		for {
			select {
			case rec, ok := <-r.queue:
				if !ok {
					return
				}
				if data, err := r.Formatter.Format(rec); err == nil {
					_ = r.Writer.Write(data)
				}
			case <-ctx.Done():
				// просто ждём закрытия очереди, drain сделает остальное
				return
			}
		}
	}()
}

// drainQueue считывает остатки очереди и вызывает Flush().
func (r *RouteProcessor) drainQueue() {
	for rec := range r.queue {
		if data, err := r.Formatter.Format(rec); err == nil {
			_ = r.Writer.Write(data)
		}
	}

	if f, ok := r.Writer.(FlushableWriter); ok {
		_ = f.Flush()
	}
}

// Close завершает работу: закрывает очередь (если ещё нет).
func (r *RouteProcessor) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return
	}

	close(r.queue)
	r.closed = true
}
