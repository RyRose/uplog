package rawdata

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
	"github.com/RyRose/uplog/internal/templates"
	"golang.org/x/exp/constraints"
)

func urlPathJoin(base string, parts ...string) string {
	p := []string{base}
	for _, part := range parts {
		p = append(p, url.PathEscape(part))
	}
	return path.Join(p...)
}

func minimum[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func maximum[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func zero[T any](v *T) T {
	if v == nil {
		var zero T
		return zero
	}
	return *v
}

func deZero[T comparable](v T) *T {
	var zero T
	if v == zero {
		return nil
	}
	return &v
}

func handleGetDataTableView[dataType any, limitParams any](
	roDB *sql.DB,
	metadata tableViewMetadata,
	selectQ func(*workoutdb.Queries, context.Context, limitParams) ([]dataType, error),
	convertQ func(int64, int64) limitParams,
	convert func(context.Context, *sql.DB, []dataType) ([]templates.DataTableRow, error),
) http.HandlerFunc {
	queries := workoutdb.New(roDB)
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var limit, offset int64
		limit, offset = 50, 0
		u, err := url.Parse(r.Header.Get("HX-Current-URL"))
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to parse current URL: %v", err), http.StatusBadRequest)
			slog.ErrorContext(ctx, "failed to parse current URL", "error", err, "url", r.Header.Get("HX-Current-URL"))
			return
		}
		rawOffset := u.Query().Get("offset")
		if rawOffset != "" {
			var err error
			offset, err = strconv.ParseInt(rawOffset, 10, 64)
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to parse offset: %v", err), http.StatusBadRequest)
				slog.ErrorContext(ctx, "failed to parse offset", "error", err)
				return
			}
		}
		rawValues, err := selectQ(queries, ctx, convertQ(limit, offset))
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to select data: %v", err), http.StatusInternalServerError)
			slog.ErrorContext(ctx, "failed to select data", "error", err)
			return
		}
		rows, err := convert(ctx, roDB, rawValues)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to convert data: %v", err), http.StatusInternalServerError)
			slog.ErrorContext(ctx, "failed to convert data", "error", err)
			return
		}
		var viewLimit int64
		viewLimit = minimum(limit, int64(len(rows)-1))
		tbl := templates.DataTable{
			Header: templates.DataTableHeader{
				Values: metadata.headers,
			},
			Rows: rows,
			Footer: templates.DataTableFooter{
				PostEndpoint: metadata.post,
				FormID:       "datatableform",
			},
			Start:       fmt.Sprint(maximum(offset, 0) + 1),
			StartOffset: fmt.Sprint(maximum(offset-limit, 0)),
			End:         fmt.Sprint(offset + viewLimit),
			LastPage:    viewLimit != limit,
		}
		templates.DataTableView(tbl).Render(ctx, w)
	}
}

type tableViewMetadata struct {
	headers []string
	post    string
}

func handlePostDataTableView[dataType, paramType any](
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

type patchIDParams[dataType any] struct {
	query   func(*workoutdb.Queries, context.Context, dataType) error
	convert func(string, string) (*dataType, error)
}

func (p *patchIDParams[dataType]) Patch(
	ctx context.Context, queries *workoutdb.Queries, key, value string) error {
	data, err := p.convert(key, value)
	if err != nil {
		return fmt.Errorf("failed to convert patch data: %w", err)
	}
	return p.query(queries, ctx, *data)
}

type patcherID interface {
	Patch(ctx context.Context, queries *workoutdb.Queries, key, value string) error
}

type patchReqParams[dataType any] struct {
	query   func(*workoutdb.Queries, context.Context, dataType) error
	convert func(*http.Request, string) (*dataType, error)
}

func (p *patchReqParams[dataType]) Patch(
	ctx context.Context, queries *workoutdb.Queries, r *http.Request, value string) error {
	data, err := p.convert(r, value)
	if err != nil {
		return fmt.Errorf("failed to convert patch data: %w", err)
	}
	return p.query(queries, ctx, *data)
}

type patcherReq interface {
	Patch(context.Context, *workoutdb.Queries, *http.Request, string) error
}

func handlePatchTableRowViewID(
	wDB *sql.DB, patchQ map[string]patcherID) http.HandlerFunc {
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

func handlePatchTableRowViewRequest(
	wDB *sql.DB, patchQ map[string]patcherReq) http.HandlerFunc {
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
