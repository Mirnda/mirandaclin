package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/mirandev/mirandaclin/internal/infra/cache"
	"gorm.io/gorm"
)

type Handler struct {
	db    *gorm.DB
	cache cache.Cache
}

func NewHandler(db *gorm.DB, c cache.Cache) *Handler {
	return &Handler{db: db, cache: c}
}

// @Summary     Liveness check
// @Tags        health
// @Produce     json
// @Success     200 {object} map[string]string
// @Router      /health [get]
func (h *Handler) Liveness(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// @Summary     Readiness check
// @Tags        health
// @Produce     json
// @Success     200 {object} map[string]any
// @Failure     503 {object} map[string]any
// @Router      /health/ready [get]
func (h *Handler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	checks := map[string]string{}
	status := http.StatusOK

	if sqlDB, err := h.db.DB(); err != nil || sqlDB.PingContext(ctx) != nil {
		checks["database"] = "unavailable"
		status = http.StatusServiceUnavailable
	} else {
		checks["database"] = "ok"
	}

	if err := h.cache.Set(ctx, "health:ping", "1", 5*time.Second); err != nil {
		checks["cache"] = "unavailable"
		status = http.StatusServiceUnavailable
	} else {
		checks["cache"] = "ok"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"status": map[int]string{200: "ok", 503: "degraded"}[status], "checks": checks})
}
