package golog

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var SENSITIVE_HEADER = []string{
	"Authorization",
	"Signature",
	"Apikey",
}

var SENSITIVE_ATTR = map[string]bool{
	"password":      true,
	"license":       true,
	"license_code":  true,
	"token":         true,
	"access_token":  true,
	"refresh_token": true,
}

type Log struct {
	logger    *zap.Logger
	loggerTDR *zap.Logger
}

func NewLogger(conf Config) LoggerInterface {
	// Validate and set defaults
	conf.Validate()

	rotator := &lumberjack.Logger{
		Filename:   conf.FileLocation + "/system.log",
		MaxSize:    conf.FileMaxSize, // megabytes
		MaxBackups: conf.FileMaxBackup,
		MaxAge:     conf.FileMaxAge, // days
	}

	rotatorTDR := &lumberjack.Logger{
		Filename:   conf.FileTDRLocation + "/tdr.log",
		MaxSize:    conf.FileMaxSize, // megabytes
		MaxBackups: conf.FileMaxBackup,
		MaxAge:     conf.FileMaxAge, // days
	}

	encoderConfig := zap.NewDevelopmentEncoderConfig()

	if conf.Env == "production" {
		encoderConfig = zap.NewProductionEncoderConfig()
	}

	encoderConfig.TimeKey = "timestamp"
	encoderConfig.LevelKey = "logLevel"
	encoderConfig.MessageKey = "message"
	encoderConfig.StacktraceKey = "stacktrace"
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	logLevel := conf.LogLevel
	if logLevel == 0 {
		logLevel = zapcore.InfoLevel
	}

	core := zapcore.NewCore(
		jsonEncoder,
		zapcore.AddSync(rotator),
		zap.NewAtomicLevelAt(logLevel),
	)

	coreTDR := zapcore.NewCore(
		jsonEncoder,
		zapcore.AddSync(rotatorTDR),
		zap.NewAtomicLevelAt(logLevel),
	)

	if conf.Stdout {
		core = zapcore.NewTee(
			core,
			zapcore.NewCore(
				consoleEncoder,
				zapcore.AddSync(os.Stdout),
				zap.NewAtomicLevelAt(logLevel),
			),
		)

		coreTDR = zapcore.NewTee(
			coreTDR,
			zapcore.NewCore(
				consoleEncoder,
				zapcore.AddSync(os.Stdout),
				zap.NewAtomicLevelAt(logLevel),
			),
		)
	}

	appVer := conf.AppVer

	// Read version file if configured and exists
	if conf.VersionFilePath != "" {
		content, err := os.ReadFile(conf.VersionFilePath)
		if err == nil {
			c := string(content)
			appVer = strings.TrimSuffix(c, "\n")
		}
	}

	logger := zap.New(core, zap.AddStacktrace(zap.ErrorLevel), zap.AddCallerSkip(2)).With(
		zap.String("app", conf.App),
		zap.String("appVer", appVer),
		zap.String("env", conf.Env),
	)

	loggerTDR := zap.New(coreTDR, zap.AddCallerSkip(2)).With(
		zap.String("app", conf.App),
		zap.String("appVer", appVer),
		zap.String("env", conf.Env),
	)

	return &Log{
		logger:    logger,
		loggerTDR: loggerTDR,
	}
}

func (l *Log) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	ctxField := populateFieldFromContext(ctx)
	fields = append(fields, ctxField...)
	l.logger.Debug(msg, fields...)
}

func (l *Log) Info(ctx context.Context, msg string, fields ...zap.Field) {
	ctxField := populateFieldFromContext(ctx)
	fields = append(fields, ctxField...)
	l.logger.Info(msg, fields...)
}

func (l *Log) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	ctxField := populateFieldFromContext(ctx)
	fields = append(fields, ctxField...)
	l.logger.Warn(msg, fields...)
}

func (l *Log) Error(ctx context.Context, msg string, err error, fields ...zap.Field) {
	ctxField := populateFieldFromContext(ctx)
	fields = append(fields, ctxField...)
	fields = append(fields, zap.Any("error", toJSON(err)))
	l.logger.Error(msg, fields...)
}

func (l *Log) Fatal(ctx context.Context, msg string, err error, fields ...zap.Field) {
	ctxField := populateFieldFromContext(ctx)
	fields = append(fields, ctxField...)
	fields = append(fields, zap.Any("error", toJSON(err)))
	l.logger.Fatal(msg, fields...)
}

func (l *Log) Panic(ctx context.Context, msg string, err error, fields ...zap.Field) {
	ctxField := populateFieldFromContext(ctx)
	fields = append(fields, ctxField...)
	fields = append(fields, zap.Any("error", toJSON(err)))
	l.logger.Panic(msg, fields...)
}

func (l *Log) TDR(ctx context.Context, log LogModel) {
	fields := populateFieldFromContext(ctx)

	fields = append(fields, zap.String("correlationId", log.CorrelationID))
	fields = append(fields, zap.Any("header", removeAuth(log.Header)))
	fields = append(fields, zap.Any("request", toJSON(maskField(log.Request))))
	fields = append(fields, zap.String("statusCode", log.StatusCode))
	fields = append(fields, zap.String("method", log.Method))
	fields = append(fields, zap.Uint64("httpStatus", log.HttpStatus))
	fields = append(fields, zap.Any("response", toJSON(maskField(log.Response))))
	fields = append(fields, zap.Int64("rt", log.ResponseTime.Milliseconds()))
	fields = append(fields, zap.Any("error", toJSON(log.Error)))
	fields = append(fields, zap.Any("otherData", toJSON(log.OtherData)))

	l.loggerTDR.Info(":", fields...)
}

// Sync flushes any buffered log entries. Applications should take care to call
// Sync before exiting to ensure all log entries are written.
func (l *Log) Sync() error {
	err1 := l.logger.Sync()
	err2 := l.loggerTDR.Sync()
	if err1 != nil {
		return err1
	}
	return err2
}

func toJSON(object interface{}) interface{} {
	if object == nil {
		return nil
	}

	// Early return for non-string types
	if w, ok := object.(string); ok {
		// Skip JSON parsing for empty strings
		if w == "" {
			return w
		}
		var jsonobj map[string]interface{}
		if err := json.Unmarshal([]byte(w), &jsonobj); err != nil {
			return w
		}
		return jsonobj
	}
	return object
}

func removeAuth(header interface{}) interface{} {
	// Fasthttp
	if mapHeader, ok := header.(fasthttp.RequestHeader); ok {
		for _, val := range SENSITIVE_HEADER {
			mapHeader.Del(val)
		}
		return string(mapHeader.Header())
	}

	// Http
	if mapHeader, ok := header.(http.Header); ok {
		for _, val := range SENSITIVE_HEADER {
			mapHeader.Del(val)
		}
	}

	return header
}

func maskField(body interface{}) interface{} {
	if body == nil {
		return nil
	}

	// Handle []byte input
	if bodyByte, ok := body.([]byte); ok {
		if len(bodyByte) == 0 {
			return bodyByte
		}
		var bodyMap map[string]interface{}
		if err := json.Unmarshal(bodyByte, &bodyMap); err != nil {
			return string(bodyByte)
		}
		return maskFieldMap(bodyMap)
	}

	// Handle map[string]interface{} directly (avoid re-marshaling)
	if bodyMap, ok := body.(map[string]interface{}); ok {
		return maskFieldMap(bodyMap)
	}

	return body
}

func maskFieldMap(bodyMap map[string]interface{}) map[string]interface{} {
	if bodyMap == nil {
		return nil
	}

	result := make(map[string]interface{}, len(bodyMap))
	for key, value := range bodyMap {
		switch v := value.(type) {
		case map[string]interface{}:
			// Recursively mask nested maps without marshaling
			result[key] = maskFieldMap(v)
		case []interface{}:
			// Handle arrays
			maskedArray := make([]interface{}, len(v))
			for i, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					maskedArray[i] = maskFieldMap(itemMap)
				} else {
					maskedArray[i] = item
				}
			}
			result[key] = maskedArray
		default:
			if isSensitiveField(key) {
				result[key] = strings.Repeat("*", 5)
			} else {
				result[key] = value
			}
		}
	}
	return result
}

func isSensitiveField(key string) bool {
	if _, ok := SENSITIVE_ATTR[strings.ToLower(key)]; ok {
		return true
	}
	return false
}

func populateFieldFromContext(ctx context.Context) []zap.Field {
	// Pre-allocate with estimated capacity (max 4 fields)
	fieldFromCtx := make([]zap.Field, 0, 4)

	// Support both typed keys and string keys for backward compatibility
	if v, ok := ctx.Value(TraceIDKey).(string); ok && v != "" {
		fieldFromCtx = append(fieldFromCtx, zap.String("traceId", v))
	} else if v, ok := ctx.Value("traceId").(string); ok && v != "" {
		fieldFromCtx = append(fieldFromCtx, zap.String("traceId", v))
	}

	if v, ok := ctx.Value(SrcIPKey).(string); ok && v != "" {
		fieldFromCtx = append(fieldFromCtx, zap.String("srcIP", v))
	} else if v, ok := ctx.Value("srcIP").(string); ok && v != "" {
		fieldFromCtx = append(fieldFromCtx, zap.String("srcIP", v))
	}

	if v, ok := ctx.Value(PortKey).(string); ok && v != "" {
		fieldFromCtx = append(fieldFromCtx, zap.String("port", v))
	} else if v, ok := ctx.Value("port").(string); ok && v != "" {
		fieldFromCtx = append(fieldFromCtx, zap.String("port", v))
	}

	if v, ok := ctx.Value(PathKey).(string); ok && v != "" {
		fieldFromCtx = append(fieldFromCtx, zap.String("path", v))
	} else if v, ok := ctx.Value("path").(string); ok && v != "" {
		fieldFromCtx = append(fieldFromCtx, zap.String("path", v))
	}

	return fieldFromCtx
}
