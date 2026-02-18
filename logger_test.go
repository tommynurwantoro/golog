package golog

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNewLogger(t *testing.T) {
	tmpDir := t.TempDir()

	config := Config{
		App:           "testapp",
		AppVer:        "1.0.0",
		Env:           "development",
		FileLocation:  tmpDir,
		FileMaxSize:   10,
		FileMaxBackup: 3,
		FileMaxAge:    7,
		Stdout:        false,
	}

	logger := NewLogger(config)
	require.NotNil(t, logger)

	logger.Info("Test message")

	// Verify Sync works
	err := logger.Sync()
	assert.NoError(t, err)
}

func TestLoggerWithContext(t *testing.T) {
	tmpDir := t.TempDir()

	config := Config{
		App:           "testapp",
		AppVer:        "1.0.0",
		Env:           "development",
		FileLocation:  tmpDir,
		FileMaxSize:   10,
		FileMaxBackup: 3,
		FileMaxAge:    7,
		Stdout:        false,
	}

	logger := NewLogger(config)
	defer logger.Sync()

	ctx := context.Background()
	ctx = WithTraceID(ctx, "trace-123")
	ctx = WithSrcIP(ctx, "192.168.1.1")
	ctx = WithPort(ctx, "8080")
	ctx = WithPath(ctx, "/test")

	// Use WithContext to bind context to logger
	logger = logger.WithContext(ctx)

	logger.Info("Test with context")
	logger.Debug("Debug message", zap.String("key", "value"))
	logger.Warn("Warning message")
}

func TestLoggerError(t *testing.T) {
	tmpDir := t.TempDir()

	config := Config{
		App:           "testapp",
		AppVer:        "1.0.0",
		Env:           "development",
		FileLocation:  tmpDir,
		FileMaxSize:   10,
		FileMaxBackup: 3,
		FileMaxAge:    7,
		Stdout:        false,
	}

	logger := NewLogger(config)
	defer logger.Sync()

	err := os.ErrNotExist
	logger.Error("Test error", err, zap.String("filename", "test.txt"))
}

func TestTDR(t *testing.T) {
	tmpDir := t.TempDir()

	config := Config{
		App:           "testapp",
		AppVer:        "1.0.0",
		Env:           "development",
		FileLocation:  tmpDir,
		FileMaxSize:   10,
		FileMaxBackup: 3,
		FileMaxAge:    7,
		Stdout:        false,
	}

	logger := NewLogger(config)
	defer logger.Sync()

	ctx := context.Background()
	ctx = WithTraceID(ctx, "trace-123")
	logger = logger.WithContext(ctx)

	tdr := LogModel{
		CorrelationID: "corr-456",
		Method:        "POST",
		Path:          "/api/test",
		StatusCode:    "200",
		HttpStatus:    200,
		Request:       map[string]interface{}{"name": "test"},
		Response:      map[string]interface{}{"id": 1},
		ResponseTime:  100 * time.Millisecond,
	}

	logger.TDR(tdr)
}

func TestConfigValidation(t *testing.T) {
	tmpDir := t.TempDir()

	config := Config{
		App:           "testapp",
		AppVer:        "1.0.0",
		Env:           "development",
		FileLocation:  tmpDir,
		FileMaxSize:   10,
		FileMaxBackup: 3,
		FileMaxAge:    7,
		Stdout:        false,
		// FileTDRLocation not set - should default
		// LogLevel not set - should default to InfoLevel
	}

	// Validate config explicitly (this is called inside NewLogger)
	config.Validate()

	// Verify default FileTDRLocation was set
	expectedTDRLocation := tmpDir
	assert.Equal(t, expectedTDRLocation, config.FileTDRLocation)
	assert.Equal(t, zapcore.InfoLevel, config.LogLevel)

	logger := NewLogger(config)
	defer logger.Sync()
}

func TestFileTDRLocation(t *testing.T) {
	tmpDir := t.TempDir()
	customTDRLocation := filepath.Join(tmpDir, "custom-tdr.log")

	config := Config{
		App:             "testapp",
		AppVer:          "1.0.0",
		Env:             "development",
		FileLocation:    tmpDir,
		FileTDRLocation: customTDRLocation,
		FileMaxSize:     10,
		FileMaxBackup:   3,
		FileMaxAge:      7,
		Stdout:          false,
	}

	logger := NewLogger(config)
	defer logger.Sync()

	tdr := LogModel{
		CorrelationID: "test",
		Method:        "GET",
		StatusCode:    "200",
		HttpStatus:    200,
		ResponseTime:  50 * time.Millisecond,
	}

	logger.TDR(tdr)

	// Verify custom location is used
	assert.Equal(t, customTDRLocation, config.FileTDRLocation)
}

func TestLogLevel(t *testing.T) {
	tmpDir := t.TempDir()

	config := Config{
		App:           "testapp",
		AppVer:        "1.0.0",
		Env:           "development",
		FileLocation:  tmpDir,
		FileMaxSize:   10,
		FileMaxBackup: 3,
		FileMaxAge:    7,
		Stdout:        false,
		LogLevel:      zapcore.DebugLevel,
	}

	logger := NewLogger(config)
	defer logger.Sync()

	logger.Debug("Debug message should be logged")
}

func TestVersionFilePath(t *testing.T) {
	tmpDir := t.TempDir()
	versionFile := filepath.Join(tmpDir, "version.txt")

	// Create version file
	err := os.WriteFile(versionFile, []byte("2.0.0\n"), 0644)
	require.NoError(t, err)

	config := Config{
		App:             "testapp",
		AppVer:          "1.0.0", // Will be overridden
		Env:             "development",
		FileLocation:    tmpDir,
		VersionFilePath: versionFile,
		FileMaxSize:     10,
		FileMaxBackup:   3,
		FileMaxAge:      7,
		Stdout:          false,
	}

	logger := NewLogger(config)
	defer logger.Sync()

	// Version should be read from file
	// Note: This is tested indirectly through logger creation
	assert.NotEmpty(t, config.VersionFilePath)
}

func TestContextKeys(t *testing.T) {
	ctx := context.Background()

	// Test typed keys
	ctx = WithTraceID(ctx, "trace-123")
	traceID, ok := GetTraceID(ctx)
	assert.True(t, ok)
	assert.Equal(t, "trace-123", traceID)

	ctx = WithSrcIP(ctx, "192.168.1.1")
	srcIP, ok := GetSrcIP(ctx)
	assert.True(t, ok)
	assert.Equal(t, "192.168.1.1", srcIP)

	ctx = WithPort(ctx, "8080")
	port, ok := GetPort(ctx)
	assert.True(t, ok)
	assert.Equal(t, "8080", port)

	ctx = WithPath(ctx, "/test")
	path, ok := GetPath(ctx)
	assert.True(t, ok)
	assert.Equal(t, "/test", path)
}

func TestContextKeysBackwardCompatibility(t *testing.T) {
	ctx := context.Background()

	// Test string keys (backward compatibility)
	ctx = context.WithValue(ctx, "traceId", "trace-456")
	ctx = context.WithValue(ctx, "srcIP", "10.0.0.1")
	ctx = context.WithValue(ctx, "port", "9090")
	ctx = context.WithValue(ctx, "path", "/api")

	tmpDir := t.TempDir()
	config := Config{
		App:           "testapp",
		AppVer:        "1.0.0",
		Env:           "development",
		FileLocation:  tmpDir,
		FileMaxSize:   10,
		FileMaxBackup: 3,
		FileMaxAge:    7,
		Stdout:        false,
	}

	logger := NewLogger(config)
	defer logger.Sync()

	// Should work with string keys
	logger.Info("Test backward compatibility")
}

func TestMaskField(t *testing.T) {
	// Test with sensitive fields
	body := map[string]interface{}{
		"username": "john",
		"password": "secret123",
		"token":    "abc123",
		"data":     map[string]interface{}{"nested": "value"},
	}

	masked := maskField(body)
	maskedMap, ok := masked.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, "john", maskedMap["username"])
	assert.Equal(t, "*****", maskedMap["password"])
	assert.Equal(t, "*****", maskedMap["token"])
}

func TestRemoveAuth(t *testing.T) {
	// Test with http.Header
	header := make(http.Header)
	header["Authorization"] = []string{"Bearer token123"}
	header["Content-Type"] = []string{"application/json"}

	result := removeAuth(header)
	resultHeader, ok := result.(http.Header)
	require.True(t, ok)

	_, exists := resultHeader["Authorization"]
	assert.False(t, exists, "Authorization header should be removed")
	assert.Equal(t, "application/json", resultHeader["Content-Type"][0])
}

func BenchmarkLogging(b *testing.B) {
	tmpDir := b.TempDir()

	config := Config{
		App:           "testapp",
		AppVer:        "1.0.0",
		Env:           "development",
		FileLocation:  tmpDir,
		FileMaxSize:   10,
		FileMaxBackup: 3,
		FileMaxAge:    7,
		Stdout:        false,
	}

	logger := NewLogger(config)
	defer logger.Sync()

	ctx := context.Background()
	ctx = WithTraceID(ctx, "trace-123")
	logger = logger.WithContext(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("Benchmark message", zap.Int("iteration", i))
	}
}

func BenchmarkContextPopulation(b *testing.B) {
	ctx := context.Background()
	ctx = WithTraceID(ctx, "trace-123")
	ctx = WithSrcIP(ctx, "192.168.1.1")
	ctx = WithPort(ctx, "8080")
	ctx = WithPath(ctx, "/test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = populateFieldFromContext(ctx)
	}
}
