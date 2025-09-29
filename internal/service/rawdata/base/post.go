package base

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
	"github.com/RyRose/uplog/internal/templates"
)

func HandlePostDataTableView[dataType, paramType any](
	roDB *sql.DB,
	wDB *sql.DB,
	insertQ func(*workoutdb.Queries, context.Context, paramType) (dataType, error),
	convertParams func(context.Context, url.Values) (*paramType, error),
	toRow func(context.Context, *workoutdb.Queries, dataType) (*templates.DataTableRow, error),
) http.HandlerFunc {
	roQ := workoutdb.New(roDB)
	wQ := workoutdb.New(wDB)
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if err := r.ParseForm(); err != nil {
			http.Error(w, fmt.Sprintf("failed to parse form: %v", err), http.StatusBadRequest)
			slog.ErrorContext(ctx, "failed to parse form", "error", err)
			return
		}
		params, err := convertParams(ctx, r.Form)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to convert data: %v", err), http.StatusBadRequest)
			slog.ErrorContext(ctx, "failed to convert data", "error", err)
			return
		}
		data, err := insertQ(wQ, ctx, *params)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to insert data: %v", err), http.StatusInternalServerError)
			slog.ErrorContext(ctx, "failed to insert data", "error", err)
			return
		}
		row, err := toRow(ctx, roQ, data)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to convert data to row: %v", err),
				http.StatusInternalServerError)
			slog.ErrorContext(ctx, "failed to convert data to row", "error", err)
			return
		}
		if err := templates.DataTableRowView(*row).Render(ctx, w); err != nil {
			http.Error(w, fmt.Sprintf("failed to render row: %v", err), http.StatusInternalServerError)
			slog.ErrorContext(ctx, "failed to render row", "error", err)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}
