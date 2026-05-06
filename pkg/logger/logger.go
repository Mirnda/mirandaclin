package logger

import "go.uber.org/zap"

// Field é um alias de zap.Field para que os importadores não dependam de zap diretamente.
type Field = zap.Field

// Logger é a interface desacoplada de logging. Nenhum pacote interno importa zap diretamente.
type Logger interface {
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Debug(msg string, fields ...Field)
	With(key, value string) Logger
	WithField(fields ...Field) Logger
	WithErr(err error) Logger
	Sync() error
}
