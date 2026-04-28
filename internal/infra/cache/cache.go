package cache

import (
	"context"
	"time"
)

// Cache é a interface de acesso ao cache — usada por todos os domínios.
// Nunca importar redis diretamente fora deste pacote.
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
	Incr(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
}
