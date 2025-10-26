package service

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/RyRose/uplog/internal/config"
	sloghttp "github.com/samber/slog-http"
	"github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"
)

func NewServer(ctx context.Context, cfg *config.Data, state *State) http.Handler {
	mux := http.NewServeMux()
	AddRoutes(ctx, mux, cfg, state)

	var handler http.Handler = mux

	// Debug logging middleware.
	if slog.Default().Enabled(ctx, slog.LevelDebug) {
		handler = sloghttp.Recovery(handler)
		handler = sloghttp.New(slog.Default())(handler)
	}

	// Prometheus metrics middleware.
	// TODO: Replace with otelhttp.
	handler = std.Handler("", middleware.New(middleware.Config{
		Recorder: prometheus.NewRecorder(prometheus.Config{}),
	}), handler)

	return handler
}
