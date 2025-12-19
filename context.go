package golog

import "context"

type contextKey string

const (
	// TraceIDKey is the context key for trace ID
	TraceIDKey contextKey = "traceId"
	// SrcIPKey is the context key for source IP
	SrcIPKey contextKey = "srcIP"
	// PortKey is the context key for port
	PortKey contextKey = "port"
	// PathKey is the context key for path
	PathKey contextKey = "path"
)

// WithTraceID adds trace ID to the context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// WithSrcIP adds source IP to the context
func WithSrcIP(ctx context.Context, srcIP string) context.Context {
	return context.WithValue(ctx, SrcIPKey, srcIP)
}

// WithPort adds port to the context
func WithPort(ctx context.Context, port string) context.Context {
	return context.WithValue(ctx, PortKey, port)
}

// WithPath adds path to the context
func WithPath(ctx context.Context, path string) context.Context {
	return context.WithValue(ctx, PathKey, path)
}

// GetTraceID retrieves trace ID from context
func GetTraceID(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(TraceIDKey).(string)
	return v, ok
}

// GetSrcIP retrieves source IP from context
func GetSrcIP(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(SrcIPKey).(string)
	return v, ok
}

// GetPort retrieves port from context
func GetPort(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(PortKey).(string)
	return v, ok
}

// GetPath retrieves path from context
func GetPath(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(PathKey).(string)
	return v, ok
}

