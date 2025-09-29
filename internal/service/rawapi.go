package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
)

func handleRawPost[dataType, retType any](
	db *sql.DB,
	insert func(*workoutdb.Queries, context.Context, dataType) (retType, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		decoder := json.NewDecoder(r.Body)

		urlquery, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			slog.ErrorContext(ctx, "failed to parse url query", "error", err)
			http.Error(w, fmt.Sprintf("failed to parse url query: %v", err), http.StatusBadRequest)
		}

		var allData []dataType
		if urlquery.Has("bulk") {
			if err := decoder.Decode(&allData); err != nil {
				slog.ErrorContext(ctx, "failed to decode bulk request", "error", err)
				http.Error(w,
					fmt.Sprintf("failed to decode bulk request: %v", err), http.StatusBadRequest)
				return
			}
		} else {
			var data dataType
			if err := decoder.Decode(&data); err != nil {
				slog.ErrorContext(ctx, "failed to decode request", "error", err)
				http.Error(w,
					fmt.Sprintf("failed to decode request: %v", err), http.StatusBadRequest)
				return
			}
			allData = append(allData, data)
		}

		queries := workoutdb.New(db)
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			slog.ErrorContext(ctx, "failed to start transaction", "error", err)
			http.Error(w, fmt.Sprintf("failed to start transaction: %v", err), http.StatusBadRequest)
			return
		}
		defer tx.Rollback()
		queries = queries.WithTx(tx)

		for i, data := range allData {
			if _, err := insert(queries, ctx, data); err != nil {
				if urlquery.Has("continue") {
					slog.WarnContext(
						ctx, "failed to insert request",
						"index", i, "error", err, "data", data)
					continue
				}
				slog.ErrorContext(
					ctx, "failed to insert request", "index", i, "error", err, "data", data)
				http.Error(w,
					fmt.Sprintf("failed to insert request %d: %v", i, err), http.StatusBadRequest)
				return
			}
		}
		if err := tx.Commit(); err != nil {
			slog.ErrorContext(ctx, "failed to commit request", "error", err)
			http.Error(w, fmt.Sprintf("failed to commit request: %v", err), http.StatusBadRequest)
			return
		}
	}
}
