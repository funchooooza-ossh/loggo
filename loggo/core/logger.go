package core

import (
	"context"
	"sync"
	"time"
)

// Logger управляет маршрутизацией логов и жизненным циклом воркеров.
type Logger struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	routes []*RouteProcessor
}

// NewLogger создаёт асинхронный логгер с переданными маршрутизаторами.
func NewLogger(routes ...*RouteProcessor) *Logger {
	ctx, cancel := context.WithCancel(context.Background())

	logger := &Logger{
		ctx:    ctx,
		cancel: cancel,
		routes: routes,
	}

	for _, r := range routes {
		r.Start(ctx, &logger.wg)
	}

	return logger
}

// Close корректно завершает все воркеры, дожидаясь полной обработки очередей и вызова Flush().
func (l *Logger) Close() {
	for _, r := range l.routes {
		r.Close()
	}
	l.cancel()
	l.wg.Wait()
}

// log формирует LogRecord и отправляет его в подходящие маршруты.
func (l *Logger) log(level LogLevel, msg string, fields map[string]interface{}) {

	record := LogRecord{
		Level:     level,
		Timestamp: time.Now(),
		Message:   msg,
		Fields:    fields,
	}

	for _, route := range l.routes {
		if route.ShouldLog(record.Level) {
			route.Enqueue(record)
		}
	}
}

// Trace отправляет TRACE-сообщение.
func (l *Logger) Trace(msg string, fields map[string]interface{}) {
	l.log(Trace, msg, fields)
}

// Debug отправляет DEBUG-сообщение.
func (l *Logger) Debug(msg string, fields map[string]interface{}) {
	l.log(Debug, msg, fields)
}

// Info отправляет INFO-сообщение.
func (l *Logger) Info(msg string, fields map[string]interface{}) {
	l.log(Info, msg, fields)
}

// Warn отправляет WARNING-сообщение.
func (l *Logger) Warn(msg string, fields map[string]interface{}) {
	l.log(Warning, msg, fields)
}

// Error отправляет ERROR-сообщение.
func (l *Logger) Error(msg string, fields map[string]interface{}) {
	l.log(Error, msg, fields)
}

// Exception отправляет сообщение об исключении.
func (l *Logger) Exception(msg string, fields map[string]interface{}) {
	l.log(Exception, msg, fields)
}
