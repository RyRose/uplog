package mux

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/RyRose/uplog/internal/templates"
)

type Trace struct {
	Mux *http.ServeMux
}

func (t *Trace) Handle(pattern string, handler http.Handler) {
	t.Mux.Handle(pattern, handler)
	// TODO: Handle tracing here.
	// t.Mux.Handle(pattern, otelhttp.NewHandler(handler, pattern))
}

func (t *Trace) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	t.Handle(pattern, http.HandlerFunc(handler))
}

type Web struct {
	Mux *Trace
}

type errorResponseWriter struct {
	http.ResponseWriter
	ctx        context.Context
	statusCode int
}

func (w *errorResponseWriter) Write(b []byte) (int, error) {
	if w.statusCode < 400 {
		return w.ResponseWriter.Write(b)
	}

	w.ResponseWriter.WriteHeader(http.StatusUnprocessableEntity)
	var s strings.Builder
	_, err := fmt.Fprint(&s, w.statusCode, ": ")
	if err != nil {
		return 0, fmt.Errorf("failed to write status code: %w", err)
	}
	n, _ := s.Write(b)
	return n, templates.Alert(s.String()).Render(w.ctx, w.ResponseWriter)
}

func (w *errorResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode

	if statusCode < 400 {
		w.ResponseWriter.WriteHeader(statusCode)
	}
}

func (w *Web) Handle(pattern string, handler http.Handler) {
	w.Mux.Handle(pattern,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.ServeHTTP(&errorResponseWriter{ResponseWriter: w, ctx: r.Context()}, r)
		}))
}

func (w *Web) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	w.Handle(pattern, http.HandlerFunc(handler))
}
