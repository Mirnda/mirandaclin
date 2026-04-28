package cache

import (
	"context"
	"errors"
	"time"
)

// ErrCacheMiss é retornado pelo Noop e pelo Redis quando a chave não existe.
var ErrCacheMiss = errors.New("cache: chave não encontrada")

type noopCache struct{}

// NewNoop retorna um cache que não faz nada — usado como fallback quando Redis está indisponível.
func NewNoop() Cache { return &noopCache{} }

func (n *noopCache) Get(_ context.Context, _ string) (string, error) {
	return "", ErrCacheMiss
}
func (n *noopCache) Set(_ context.Context, _, _ string, _ time.Duration) error { return nil }
func (n *noopCache) Del(_ context.Context, _ ...string) error                  { return nil }
func (n *noopCache) Incr(_ context.Context, _ string) (int64, error)           { return 0, nil }
func (n *noopCache) Expire(_ context.Context, _ string, _ time.Duration) error { return nil }
