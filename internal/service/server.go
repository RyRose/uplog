package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/RyRose/uplog/internal/calendar"
	"github.com/RyRose/uplog/internal/templates"
	sloghttp "github.com/samber/slog-http"
	"github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"
)

type errorResponseWriter struct {
	http.ResponseWriter
	requestCtx context.Context
	statusCode int
}

func (w *errorResponseWriter) Write(b []byte) (int, error) {
	if w.statusCode < 400 {
		return w.ResponseWriter.Write(b)
	}

	w.ResponseWriter.WriteHeader(http.StatusUnprocessableEntity)
	var s strings.Builder
	n, err := fmt.Fprint(&s, w.statusCode, ": ")
	if err != nil {
		return 0, fmt.Errorf("failed to write status code: %w", err)
	}
	n, _ = s.Write(b)
	return n, templates.Alert(s.String()).Render(w.requestCtx, w.ResponseWriter)
}

func (w *errorResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode

	if statusCode < 400 {
		w.ResponseWriter.WriteHeader(statusCode)
	}
}

func alertMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(&errorResponseWriter{ResponseWriter: w, requestCtx: r.Context()}, r)
	})
}

func NewServer(ctx context.Context, db, roDB *sql.DB, calendarSrv *calendar.Service) http.Handler {
	mux := http.NewServeMux()
	addRoutes(ctx, mux, db, roDB, calendarSrv)

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
