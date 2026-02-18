package golog

import (
	"context"

	"go.uber.org/zap"
)

type LoggerInterface interface {
	WithContext(ctx context.Context) LoggerInterface
	Debug(message string, fields ...zap.Field)
	Info(message string, fields ...zap.Field)
	Warn(message string, fields ...zap.Field)
	Error(message string, err error, fields ...zap.Field)
	Fatal(message string, err error, fields ...zap.Field)
	Panic(message string, err error, fields ...zap.Field)
	TDR(tdr LogModel)
	Sync() error
}
