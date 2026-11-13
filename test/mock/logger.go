package mock

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type TestLogger struct {
	mu       sync.Mutex
	Logger   *zap.Logger
	Observer *observer.ObservedLogs
}

func NewTestLogger() *TestLogger {
	core, logs := observer.New(zapcore.DebugLevel)
	return &TestLogger{
		Logger:   zap.New(core),
		Observer: logs,
	}
}

func (l *TestLogger) All() []observer.LoggedEntry {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.Observer.All()
}

func (l *TestLogger) TakeAll() []observer.LoggedEntry {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.Observer.TakeAll()
}
