# 🔪 Golog

[![Go Report Card](https://goreportcard.com/badge/github.com/tommynurwantoro/golog)](https://goreportcard.com/report/github.com/tommynurwantoro/golog)
[![GoDoc](https://godoc.org/github.com/tommynurwantoro/golog?status.svg)](https://godoc.org/github.com/tommynurwantoro/golog)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A high-performance, structured logging library for Go services. Golog provides standardized logging with automatic log rotation, context-aware tracing, and sensitive data masking.

## Features

- 🚀 **High Performance**: Asynchronous buffered I/O for optimal performance
- 📝 **Structured Logging**: JSON-formatted logs with context-aware fields
- 🔄 **Automatic Rotation**: Configurable log file rotation based on size and age
- 🔒 **Security**: Automatic masking of sensitive data (passwords, tokens, etc.)
- 📊 **TDR Logging**: Transaction Detail Request logging for API requests/responses
- 🎯 **Type-Safe Context**: Typed context keys for better code safety
- 🔌 **Flexible Usage**: Singleton pattern or direct logger instances
- 📦 **Production Ready**: Built on top of [zap](https://github.com/uber-go/zap) logger

## Installation

Install the library using Go modules:

```shell
go get -u github.com/tommynurwantoro/golog
```

## Quick Start

```golang
package main

import (
    "context"
    "github.com/tommynurwantoro/golog"
)

func main() {
    // Initialize logger
    config := golog.Config{
        App:          "myapp",
        AppVer:       "1.0.0",
        Env:          "development",
        FileLocation: "/var/log/myapp",
        FileMaxSize:  100, // megabytes
        FileMaxBackup: 5,
        FileMaxAge:   30, // days
        Stdout:       true,
    }
    
    logger := golog.Load(config)
    defer golog.Sync()
    
    // Create context with trace information
    ctx := context.Background()
    ctx = golog.WithTraceID(ctx, "trace-123")
    
    // Log messages
    logger.Info(ctx, "Application started")
    logger.Debug(ctx, "Debug information")
    logger.Warn(ctx, "Warning message")
}
```

## Configuration

The `Config` struct allows you to customize logging behavior. All fields are validated and defaults are applied automatically.

### Configuration Fields

| Field | Type | Required | Default | Description |
| --- | --- | --- | --- | --- |
| `App` | `string` | Yes | - | Application name (e.g., "myapp") |
| `AppVer` | `string` | Yes | - | Application version (e.g., "1.0.0") |
| `Env` | `string` | Yes | - | Environment: `"development"` or `"production"` |
| `FileLocation` | `string` | Yes | - | Directory where system logs will be saved |
| `FileTDRLocation` | `string` | No | `FileLocation + "/tdr.log"` | Path for TDR (Transaction Detail Request) logs |
| `FileMaxSize` | `int` | Yes | - | Maximum log file size in megabytes before rotation |
| `FileMaxBackup` | `int` | Yes | - | Maximum number of backup files to keep |
| `FileMaxAge` | `int` | Yes | - | Maximum age of backup files in days |
| `Stdout` | `bool` | No | `false` | Enable console output (useful for development) |
| `LogLevel` | `zapcore.Level` | No | `InfoLevel` | Minimum log level (Debug, Info, Warn, Error) |
| `VersionFilePath` | `string` | No | `"version.txt"` | Path to version file (overrides AppVer if exists) |

### Example Configuration

```golang
config := golog.Config{
    App:            "myapp",
    AppVer:         "1.0.0",
    Env:            "production",
    FileLocation:   "/var/log/myapp",
    FileTDRLocation: "/var/log/myapp/tdr.log", // Optional
    FileMaxSize:    500,                        // 500 MB
    FileMaxBackup:  10,                         // Keep 10 backups
    FileMaxAge:     90,                         // Keep for 90 days
    Stdout:         false,                      // Disable in production
    LogLevel:       zapcore.InfoLevel,         // Optional
    VersionFilePath: "version.txt",             // Optional
}
```

## Usage

### Singleton Pattern (Recommended)

The singleton pattern allows you to use the logger from anywhere in your application without passing it around:

```golang
package main

import (
    "context"
    "github.com/tommynurwantoro/golog"
    "go.uber.org/zap"
)

func main() {
    config := golog.Config{
        App:          "myapp",
        AppVer:       "1.0.0",
        Env:          "development",
        FileLocation: "/var/log/myapp",
        FileMaxSize:  100,
        FileMaxBackup: 5,
        FileMaxAge:   30,
        Stdout:       true,
    }
    
    // Initialize singleton logger
    golog.Load(config)
    defer golog.Sync() // Important: flush logs on exit
    
    // Use from anywhere
    ctx := context.Background()
    ctx = golog.WithTraceID(ctx, "trace-123")
    
    golog.Info(ctx, "Application started")
    golog.Debug(ctx, "Debug message", zap.String("key", "value"))
    golog.Warn(ctx, "Warning message")
}
```

### Direct Logger Instance

For more control or when you need multiple logger instances:

```golang
package main

import (
    "context"
    "github.com/tommynurwantoro/golog"
)

func main() {
    config := golog.Config{
        App:          "myapp",
        AppVer:       "1.0.0",
        Env:          "production",
        FileLocation: "/var/log/myapp",
        FileMaxSize:  500,
        FileMaxBackup: 10,
        FileMaxAge:   90,
        Stdout:       false,
    }
    
    logger := golog.NewLogger(config)
    defer logger.Sync() // Important: flush logs before exit
    
    ctx := context.Background()
    logger.Info(ctx, "Application started")
}
```

### Logging Methods

All logging methods accept a context and message, with optional fields:

```golang
ctx := context.Background()
ctx = golog.WithTraceID(ctx, "trace-123")

// Info level logging
golog.Info(ctx, "User logged in", zap.String("userID", "123"))

// Debug level logging
golog.Debug(ctx, "Processing request", zap.String("method", "GET"))

// Warning level logging
golog.Warn(ctx, "Rate limit approaching", zap.Int("requests", 95))

// Error level logging
err := errors.New("database connection failed")
golog.Error(ctx, "Failed to connect", err, zap.String("host", "db.example.com"))

// Fatal level logging (exits application)
// golog.Fatal(ctx, "Critical error", err)

// Panic level logging (panics)
// golog.Panic(ctx, "Unexpected error", err)
```

### Context-Aware Logging

Golog automatically extracts trace information from the context. This makes it easy to track requests across your application.

#### Using Typed Context Keys (Recommended)

Type-safe context keys prevent typos and provide better IDE support:

```golang
import (
    "context"
    "github.com/tommynurwantoro/golog"
)

func handleRequest(ctx context.Context) {
    // Add trace information to context
    ctx = golog.WithTraceID(ctx, "trace-123")
    ctx = golog.WithSrcIP(ctx, "192.168.1.1")
    ctx = golog.WithPort(ctx, "8080")
    ctx = golog.WithPath(ctx, "/api/users")
    
    // All logs will automatically include these fields
    golog.Info(ctx, "Request received")
    
    // Retrieve values if needed
    traceID, ok := golog.GetTraceID(ctx)
    if ok {
        // Use traceID
    }
}
```

#### Using String Keys (Backward Compatible)

For backward compatibility, you can still use string keys:

```golang
ctx = context.WithValue(ctx, "traceId", "trace-123")
ctx = context.WithValue(ctx, "srcIP", "192.168.1.1")
ctx = context.WithValue(ctx, "port", "8080")
ctx = context.WithValue(ctx, "path", "/api/users")
```

#### Available Context Keys

- `traceId` / `golog.WithTraceID()` - Request trace ID
- `srcIP` / `golog.WithSrcIP()` - Source IP address
- `port` / `golog.WithPort()` - Port number
- `path` / `golog.WithPath()` - Request path

All context values are automatically included in log entries.

### Transaction Detail Request (TDR) Logging

TDR logging captures complete request/response information for API calls, including headers, request/response bodies, status codes, and response times. Sensitive data is automatically masked.

```golang
import (
    "context"
    "time"
    "github.com/tommynurwantoro/golog"
)

func logAPIRequest(ctx context.Context, req, resp interface{}) {
    tdr := golog.LogModel{
        CorrelationID: "corr-456",
        Method:        "POST",
        Path:          "/api/users",
        StatusCode:   "200",
        HttpStatus:    200,
        Header:        requestHeaders,      // http.Header or fasthttp.RequestHeader
        Request:       req,                 // Request body (will be masked)
        Response:      resp,               // Response body (will be masked)
        ResponseTime:  150 * time.Millisecond,
        Error:         nil,                 // Error if any
        OtherData:     map[string]interface{}{"custom": "data"},
    }
    
    golog.TDR(ctx, tdr)
}
```

**Note**: TDR logs are written to a separate file (`FileTDRLocation`) for easier analysis and monitoring.

#### Sensitive Data Masking

Golog automatically masks sensitive fields in request/response bodies:

- `password`
- `license`, `license_code`
- `token`, `access_token`, `refresh_token`

Sensitive headers are also removed:
- `Authorization`
- `Signature`
- `Apikey`

Masked values are replaced with `*****` in logs.

## Log Output

### File Output

By default, logs are written to files in the `FileLocation` directory:
- System logs: `{FileLocation}/system.log`
- TDR logs: `{FileTDRLocation}` (defaults to `{FileLocation}/tdr.log`)

### Console Output

Enable console output for development by setting `Stdout: true`:

```golang
config := golog.Config{
    // ... other config
    Stdout: true, // Logs will appear in console
}
```

In production, set `Stdout: false` for better performance.

### Log Rotation

Log files are automatically rotated when they reach `FileMaxSize`:

- Old files are renamed with a timestamp suffix
- Only `FileMaxBackup` backup files are kept
- Files older than `FileMaxAge` days are automatically deleted

Example rotation:
```
system.log
system.log.2024-01-15
system.log.2024-01-14
```

### Log Format

Logs are written in JSON format for easy parsing:

```json
{
  "timestamp": "2024-01-15T10:30:45Z",
  "logLevel": "INFO",
  "message": "User logged in",
  "app": "myapp",
  "appVer": "1.0.0",
  "env": "production",
  "traceId": "trace-123",
  "srcIP": "192.168.1.1",
  "path": "/api/users",
  "userID": "123"
}
```

## Performance Considerations

### Best Practices

1. **Always call `Sync()` on shutdown**: Ensures all buffered logs are written
   ```golang
   defer golog.Sync() // For singleton
   defer logger.Sync() // For direct logger
   ```

2. **Use appropriate log levels**: Set `LogLevel` to filter out unnecessary logs in production
   ```golang
   config.LogLevel = zapcore.InfoLevel // Only log Info and above
   ```

3. **Disable stdout in production**: Console output adds overhead
   ```golang
   config.Stdout = false // Production
   config.Stdout = true  // Development
   ```

4. **Use context efficiently**: Pre-populate context at request entry point
   ```golang
   // At HTTP handler entry
   ctx = golog.WithTraceID(ctx, generateTraceID())
   ctx = golog.WithPath(ctx, r.URL.Path)
   ```

### Performance Characteristics

- **Asynchronous I/O**: Logs are buffered and written asynchronously
- **Zero-allocation context extraction**: Context fields are extracted efficiently
- **Optimized JSON operations**: Sensitive data masking is optimized for performance

## Advanced Usage

### Custom Log Levels

```golang
import "go.uber.org/zap/zapcore"

config := golog.Config{
    // ... other config
    LogLevel: zapcore.DebugLevel, // Enable debug logs
}
```

### Version File Override

If a version file exists, it will override `AppVer`:

```golang
config := golog.Config{
    AppVer:         "1.0.0",        // Default version
    VersionFilePath: "version.txt", // File containing "1.2.3"
    // AppVer will be "1.2.3" if file exists
}
```

### Testing

Reset the singleton logger between tests:

```golang
func TestSomething(t *testing.T) {
    golog.Reset() // Reset singleton
    
    config := golog.Config{/* test config */}
    logger := golog.Load(config)
    defer golog.Sync()
    
    // Your test code
}
```

## Troubleshooting

### Logs not appearing

- Ensure `Sync()` is called before application exit
- Check file permissions for `FileLocation` directory
- Verify `LogLevel` allows the log level you're using

### Performance issues

- Disable `Stdout` in production
- Increase `FileMaxSize` to reduce rotation frequency
- Use appropriate `LogLevel` to filter unnecessary logs

### Missing context fields

- Ensure context is passed through your call chain
- Use typed context keys (`golog.WithTraceID()`) for type safety
- Check that context values are set before logging

## Examples

See the [example_test.go](example_test.go) file for more comprehensive examples including:
- Basic logging
- Error handling
- TDR logging
- Context usage
- Configuration options

## Contributing

Contributions are welcome! Please follow the [Contribution Guidelines](CONTRIBUTING.md).

### Development Setup

```shell
git clone https://github.com/tommynurwantoro/golog.git
cd golog
go mod download
go test ./...
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built on top of [zap](https://github.com/uber-go/zap) - A blazing fast, structured, leveled logging library
- Uses [lumberjack](https://github.com/natefinch/lumberjack) for log rotation
