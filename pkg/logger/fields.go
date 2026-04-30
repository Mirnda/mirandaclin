package logger

import "go.uber.org/zap"

// Re-exporta os construtores de Field para que nenhum pacote interno precise importar zap diretamente.

func String(key, val string) Field        { return zap.String(key, val) }
func Err(err error) Field                 { return zap.Error(err) }
func Int(key string, val int) Field       { return zap.Int(key, val) }
func Int64(key string, val int64) Field   { return zap.Int64(key, val) }
func Bool(key string, val bool) Field     { return zap.Bool(key, val) }
