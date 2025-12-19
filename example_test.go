package golog

import (
	"context"
	"os"
	"time"

	"go.uber.org/zap"
)

func ExampleLoad() {
	// Initialize logger with configuration
	config := Config{
		App:          "myapp",
		AppVer:       "1.0.0",
		Env:          "development",
		FileLocation: "/tmp/logs",
		FileMaxSize:  100, // megabytes
		FileMaxBackup: 5,
		FileMaxAge:   30, // days
		Stdout:       true,
	}

	logger := Load(config)
	defer Sync()

	// Create context with trace information
	ctx := context.Background()
	ctx = WithTraceID(ctx, "trace-123")
	ctx = WithSrcIP(ctx, "192.168.1.1")
	ctx = WithPath(ctx, "/api/users")

	// Log messages
	logger.Info(ctx, "Application started")
	logger.Debug(ctx, "Debug information", zap.String("key", "value"))
	logger.Warn(ctx, "Warning message")
}

func ExampleNewLogger() {
	// Create logger instance directly (non-singleton)
	config := Config{
		App:          "myapp",
		AppVer:       "1.0.0",
		Env:          "production",
		FileLocation: "/var/log/myapp",
		FileMaxSize:  500,
		FileMaxBackup: 10,
		FileMaxAge:   90,
		Stdout:       false,
	}

	logger := NewLogger(config)
	defer logger.Sync()

	ctx := context.Background()
	logger.Info(ctx, "Direct logger usage")
}

func ExampleTDR() {
	config := Config{
		App:          "myapp",
		AppVer:       "1.0.0",
		Env:          "development",
		FileLocation: "/tmp/logs",
		FileMaxSize:  100,
		FileMaxBackup: 5,
		FileMaxAge:   30,
		Stdout:       true,
	}

	Load(config)
	defer Sync()

	ctx := context.Background()
	ctx = WithTraceID(ctx, "trace-123")

	// Log TDR (Transaction Detail Request)
	tdr := LogModel{
		CorrelationID: "corr-456",
		Method:        "POST",
		Path:          "/api/users",
		StatusCode:   "200",
		HttpStatus:    200,
		Request:       map[string]interface{}{"name": "John"},
		Response:      map[string]interface{}{"id": 1, "name": "John"},
		ResponseTime:  150 * time.Millisecond,
	}

	TDR(ctx, tdr)
}

func ExampleWithTraceID() {
	ctx := context.Background()

	// Using typed context keys (recommended)
	ctx = WithTraceID(ctx, "trace-123")
	ctx = WithSrcIP(ctx, "192.168.1.1")
	ctx = WithPort(ctx, "8080")
	ctx = WithPath(ctx, "/api/users")

	// Retrieve values
	traceID, ok := GetTraceID(ctx)
	if ok {
		_ = traceID
	}

	// Using string keys (backward compatible)
	ctx = context.WithValue(ctx, "traceId", "trace-456")
	ctx = context.WithValue(ctx, "srcIP", "10.0.0.1")
}

func ExampleLog_Error() {
	config := Config{
		App:          "myapp",
		AppVer:       "1.0.0",
		Env:          "development",
		FileLocation: "/tmp/logs",
		FileMaxSize:  100,
		FileMaxBackup: 5,
		FileMaxAge:   30,
		Stdout:       true,
	}

	logger := Load(config)
	defer Sync()

	ctx := context.Background()
	ctx = WithTraceID(ctx, "trace-123")

	// Log error
	err := os.ErrNotExist
	logger.Error(ctx, "Failed to open file", err, zap.String("filename", "config.json"))

	// Log fatal (exits application)
	// logger.Fatal(ctx, "Critical error", err)

	// Log panic (panics)
	// logger.Panic(ctx, "Unexpected error", err)
}

func ExampleNewLogger_configDefaults() {
	config := Config{
		App:          "myapp",
		AppVer:       "1.0.0",
		Env:          "development",
		FileLocation: "/tmp/logs",
		// FileTDRLocation will default to FileLocation + "/tdr.log"
		FileMaxSize:   100,
		FileMaxBackup: 5,
		FileMaxAge:    30,
		Stdout:        true,
		// LogLevel defaults to InfoLevel if not set
		// VersionFilePath defaults to "version.txt" if not set
	}

	// Config is automatically validated when creating logger
	logger := NewLogger(config)
	defer logger.Sync()
}

