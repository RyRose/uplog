package rawdata

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
)

func HandleDeleteTableRowViewID(
	wDB *sql.DB,
	deleteQ func(*workoutdb.Queries, context.Context, string) error,
) http.HandlerFunc {
	queries := workoutdb.New(wDB)
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := r.PathValue("id")
		if err := deleteQ(queries, ctx, id); err != nil {
			http.Error(w, fmt.Sprintf("failed to delete row: %v", err), http.StatusInternalServerError)
			slog.ErrorContext(ctx, "failed to delete row", "error", err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
func HandleDeleteTableRowViewRequest[idType any](
	wDB *sql.DB,
	deleteQ func(*workoutdb.Queries, context.Context, idType) error,
	convert func(*http.Request) (*idType, error),
) http.HandlerFunc {
	queries := workoutdb.New(wDB)
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id, err := convert(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to convert id: %v", err), http.StatusBadRequest)
			slog.ErrorContext(ctx, "failed to convert id", "error", err)
			return
		}
		if err := deleteQ(queries, ctx, *id); err != nil {
			http.Error(w, fmt.Sprintf("failed to delete row: %v", err), http.StatusInternalServerError)
			slog.ErrorContext(ctx, "failed to delete row", "error", err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
