package base

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
)

type PatchIDParams[dataType any] struct {
	Query   func(*workoutdb.Queries, context.Context, dataType) error
	Convert func(string, string) (*dataType, error)
}

func (p *PatchIDParams[dataType]) Patch(
	ctx context.Context, queries *workoutdb.Queries, key, value string) error {
	data, err := p.Convert(key, value)
	if err != nil {
		return fmt.Errorf("failed to convert patch data: %w", err)
	}
	return p.Query(queries, ctx, *data)
}

type PatcherID interface {
	Patch(ctx context.Context, queries *workoutdb.Queries, key, value string) error
}

type PatchReqParams[dataType any] struct {
	Query   func(*workoutdb.Queries, context.Context, dataType) error
	Convert func(*http.Request, string) (*dataType, error)
}

func (p *PatchReqParams[dataType]) Patch(
	ctx context.Context, queries *workoutdb.Queries, r *http.Request, value string) error {
	data, err := p.Convert(r, value)
	if err != nil {
		return fmt.Errorf("failed to convert patch data: %w", err)
	}
	return p.Query(queries, ctx, *data)
}

type PatcherReq interface {
	Patch(context.Context, *workoutdb.Queries, *http.Request, string) error
}

func HandlePatchTableRowViewID(
	wDB *sql.DB, patchQ map[string]PatcherID) http.HandlerFunc {
	queries := workoutdb.New(wDB)
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := r.PathValue("id")
		if err := r.ParseForm(); err != nil {
			http.Error(w, fmt.Sprintf("failed to parse form: %v", err), http.StatusBadRequest)
			slog.ErrorContext(ctx, "failed to parse form", "error", err)
			return
		}
		for param, values := range r.Form {
			patcher, ok := patchQ[param]
			if !ok {
				slog.WarnContext(ctx, "unknown patch parameter", "param", param)
				continue
			}
			if len(values) != 1 {
				slog.WarnContext(ctx, "unexpected number of values", "param", param, "values", values)
				continue
			}
			if err := patcher.Patch(ctx, queries, id, values[0]); err != nil {
				http.Error(w, fmt.Sprintf("failed to patch row: %v", err), http.StatusInternalServerError)
				slog.ErrorContext(ctx, "failed to patch row", "error", err)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, "no patch data provided", http.StatusBadRequest)
		slog.ErrorContext(ctx, "no patch data provided", "id", id)
	}
}

func HandlePatchTableRowViewRequest(
	wDB *sql.DB, patchQ map[string]PatcherReq) http.HandlerFunc {
	queries := workoutdb.New(wDB)
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if err := r.ParseForm(); err != nil {
			http.Error(w, fmt.Sprintf("failed to parse form: %v", err), http.StatusBadRequest)
			slog.ErrorContext(ctx, "failed to parse form", "error", err)
			return
		}
		for param, values := range r.Form {
			patcher, ok := patchQ[param]
			if !ok {
				slog.WarnContext(ctx, "unknown patch parameter", "param", param)
				continue
			}
			if len(values) != 1 {
				slog.WarnContext(ctx, "unexpected number of values", "param", param, "values", values)
				continue
			}
			if err := patcher.Patch(ctx, queries, r, values[0]); err != nil {
				http.Error(w, fmt.Sprintf("failed to patch row: %v", err), http.StatusInternalServerError)
				slog.ErrorContext(ctx, "failed to patch row", "error", err)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, "no patch data provided", http.StatusBadRequest)
	}
}
