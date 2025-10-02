package service

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/RyRose/uplog/internal/calendar"
	sloghttp "github.com/samber/slog-http"
	"github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"
)

func NewServer(ctx context.Context, db, roDB *sql.DB, calendarSrv *calendar.Service) http.Handler {
	mux := http.NewServeMux()
	AddRoutes(ctx, mux, db, roDB, calendarSrv)

	var handler http.Handler = mux

	// Debug logging middleware.
	if slog.Default().Enabled(ctx, slog.LevelDebug) {
		handler = sloghttp.Recovery(handler)
		handler = sloghttp.New(slog.Default())(handler)
	}

	// Prometheus metrics middleware.
	handler = std.Handler("", middleware.New(middleware.Config{
		Recorder: prometheus.NewRecorder(prometheus.Config{}),
	}), handler)

	return handler
}
