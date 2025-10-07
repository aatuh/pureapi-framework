package accesslog

import (
	"context"
	"log"
	"time"
)

// Entry captures request/response metadata for structured logging.
type Entry struct {
	Method       string
	Path         string
	Status       int
	Duration     time.Duration
	RequestID    string
	RemoteAddr   string
	UserAgent    string
	ResponseSize int
	Err          error
}

// AccessLogger handles structured access log entries.
type AccessLogger interface {
	Log(ctx context.Context, entry Entry)
}

// LoggerFunc lifts a function into an AccessLogger.
type LoggerFunc func(ctx context.Context, entry Entry)

// Log implements AccessLogger.
func (f LoggerFunc) Log(ctx context.Context, entry Entry) {
	f(ctx, entry)
}

// StdLogger writes entries using the provided log.Logger.
type StdLogger struct {
	logger *log.Logger
}

// NewStdLogger builds an AccessLogger backed by log.Logger.
func NewStdLogger(l *log.Logger) *StdLogger {
	if l == nil {
		l = log.Default()
	}
	return &StdLogger{logger: l}
}

// Log implements AccessLogger.
func (l *StdLogger) Log(_ context.Context, entry Entry) {
	l.logger.Printf("method=%s path=%s status=%d duration=%s bytes=%d request_id=%s error=%v", entry.Method, entry.Path, entry.Status, entry.Duration, entry.ResponseSize, entry.RequestID, entry.Err)
}
