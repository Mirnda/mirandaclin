package logger

import (
	"context"

	"go.uber.org/zap"
)

type ctxKey struct{}

var loggerKey = ctxKey{}

type zapLogger struct {
	z *zap.Logger
}

// New cria um Logger concreto baseado em zap.
// Em development usa console colorido com nível debug; demais envs usam JSON com nível info.
func New(env string) Logger {
	var z *zap.Logger
	if env == "development" {
		z, _ = zap.NewDevelopment(
			zap.AddCallerSkip(1),
			// stacktraces as multi-line text break Loki log grouping; caller field is enough for dev
			zap.AddStacktrace(zap.ErrorLevel),
		)
	} else {
		z, _ = zap.NewProduction(zap.AddCallerSkip(1))
	}
	return &zapLogger{z: z}
}

func FromContext(ctx context.Context) Logger {
	if l, ok := ctx.Value(loggerKey).(Logger); ok {
		return l
	}
	return &zapLogger{z: zap.NewNop()}
}

func WithContext(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

func (l *zapLogger) Info(msg string, fields ...Field)  { l.z.Info(msg, fields...) }
func (l *zapLogger) Warn(msg string, fields ...Field)  { l.z.Warn(msg, fields...) }
func (l *zapLogger) Error(msg string, fields ...Field) { l.z.Error(msg, fields...) }
func (l *zapLogger) Debug(msg string, fields ...Field) { l.z.Debug(msg, fields...) }
func (l *zapLogger) Sync() error                       { return l.z.Sync() }

func (l *zapLogger) With(fields ...Field) Logger {
	return &zapLogger{z: l.z.With(fields...)}
}
