package golog

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

var (
	once      sync.Once
	singleton LoggerInterface
	mu        sync.RWMutex
)

// Load constructs and returns a singleton logger instance.
// The logger is initialized only once on the first call.
func Load(config Config) LoggerInterface {
	once.Do(func() {
		singleton = NewLogger(config)
	})

	return singleton
}

// Reset resets the singleton logger. This is primarily useful for testing.
// It should not be called in production code.
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	once = sync.Once{}
	singleton = nil
}

func WithContext(ctx context.Context) LoggerInterface {
	mu.RLock()
	defer mu.RUnlock()
	if singleton != nil {
		return singleton.WithContext(ctx)
	}
	return nil
}

// Debug logs a message at DebugLevel.
func Debug(msg string, fields ...zap.Field) {
	mu.RLock()
	defer mu.RUnlock()
	if singleton != nil {
		singleton.Debug(msg, fields...)
	}
}

// Info logs a message at InfoLevel.
func Info(msg string, fields ...zap.Field) {
	mu.RLock()
	defer mu.RUnlock()
	if singleton != nil {
		singleton.Info(msg, fields...)
	}
}

// Warn logs a message at WarnLevel.
func Warn(msg string, fields ...zap.Field) {
	mu.RLock()
	defer mu.RUnlock()
	if singleton != nil {
		singleton.Warn(msg, fields...)
	}
}

// Error logs a message at ErrorLevel.
func Error(msg string, err error, fields ...zap.Field) {
	mu.RLock()
	defer mu.RUnlock()
	if singleton != nil {
		singleton.Error(msg, err, fields...)
	}
}

// Fatal logs a message at FatalLevel.
//
// The logger then calls os.Exit(1), even if logging at FatalLevel is
// disabled.
func Fatal(msg string, err error, fields ...zap.Field) {
	mu.RLock()
	defer mu.RUnlock()
	if singleton != nil {
		singleton.Fatal(msg, err, fields...)
	}
}

// Panic logs a message at PanicLevel.
//
// The logger then panics, even if logging at PanicLevel is disabled.
func Panic(msg string, err error, fields ...zap.Field) {
	mu.RLock()
	defer mu.RUnlock()
	if singleton != nil {
		singleton.Panic(msg, err, fields...)
	}
}

// TDR (Transaction Detail Request) consist of request and response log
func TDR(model LogModel) {
	mu.RLock()
	defer mu.RUnlock()
	if singleton != nil {
		singleton.TDR(model)
	}
}

// Sync flushes any buffered log entries. Applications should take care to call
// Sync before exiting to ensure all log entries are written.
func Sync() error {
	mu.RLock()
	defer mu.RUnlock()
	if singleton == nil {
		return nil
	}
	return singleton.Sync()
}
