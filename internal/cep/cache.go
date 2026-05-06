package cep

import (
	"context"
	"encoding/json"
	"time"
)

const cacheTTL = 3 * 24 * time.Hour

func cacheKey(id string) string { return "cep:" + id }

// getCache returns a cached Address or nil on miss/error.
func (h *Handler) getCache(ctx context.Context, id string) *AddressResponse {
	b, err := h.cache.Get(ctx, cacheKey(id))
	if err != nil {
		return nil
	}
	var addr AddressResponse
	if err := json.Unmarshal([]byte(b), &addr); err != nil {
		return nil
	}
	return &addr
}

// setCache stores an Address in the cache, ignoring errors (best-effort).
func (h *Handler) setCache(ctx context.Context, addr *AddressResponse) {
	b, err := json.Marshal(addr)
	if err != nil {
		return
	}
	_ = h.cache.Set(ctx, cacheKey(addr.PostalCode), string(b), cacheTTL)
}
