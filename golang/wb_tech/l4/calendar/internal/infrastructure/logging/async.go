package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"
)

type Record struct {
	Timestamp time.Time      `json:"timestamp"`
	Level     string         `json:"level"`
	Message   string         `json:"message"`
	Fields    map[string]any `json:"fields,omitempty"`
}

type AsyncLogger struct {
	writer io.Writer
	ch     chan Record

	mu     sync.RWMutex
	closed bool
	wg     sync.WaitGroup
}

func NewAsyncLogger(writer io.Writer, buffer int) *AsyncLogger {
	if writer == nil {
		writer = io.Discard
	}
	if buffer <= 0 {
		buffer = 128
	}

	logger := &AsyncLogger{
		writer: writer,
		ch:     make(chan Record, buffer),
	}

	logger.wg.Add(1)
	go logger.run()

	return logger
}

func (l *AsyncLogger) Info(message string, fields map[string]any) {
	l.enqueue("INFO", message, fields)
}

func (l *AsyncLogger) Error(message string, fields map[string]any) {
	l.enqueue("ERROR", message, fields)
}

func (l *AsyncLogger) Close() {
	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()
		return
	}
	l.closed = true
	close(l.ch)
	l.mu.Unlock()

	l.wg.Wait()
}

func (l *AsyncLogger) enqueue(level string, message string, fields map[string]any) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.closed {
		return
	}

	record := Record{
		Timestamp: time.Now().UTC(),
		Level:     level,
		Message:   message,
		Fields:    fields,
	}

	select {
	case l.ch <- record:
	default:
	}
}

func (l *AsyncLogger) run() {
	defer l.wg.Done()

	for record := range l.ch {
		payload, err := json.Marshal(record)
		if err != nil {
			l.writeMarshalFallback(err)
			continue
		}
		_, _ = l.writer.Write(append(payload, '\n'))
	}
}

func (l *AsyncLogger) writeMarshalFallback(err error) {
	fallback := map[string]any{
		"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		"level":     "ERROR",
		"message":   "marshal log record failed",
		"fields": map[string]string{
			"error": err.Error(),
		},
	}

	payload, marshalErr := json.Marshal(fallback)
	if marshalErr != nil {
		_, _ = fmt.Fprintf(l.writer, "{\"level\":\"ERROR\",\"message\":\"marshal log record failed\"}\n")
		return
	}

	_, _ = l.writer.Write(append(payload, '\n'))
}
