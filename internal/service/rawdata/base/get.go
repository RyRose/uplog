package base

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/RyRose/uplog/internal/service/rawdata/util"
	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
	"github.com/RyRose/uplog/internal/templates"
)

func HandleGetDataTableView[dataType any, limitParams any](
	roDB *sql.DB,
	metadata TableViewMetadata,
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
		viewLimit = util.Minimum(limit, int64(len(rows)-1))
		tbl := templates.DataTable{
			Header: templates.DataTableHeader{
				Values: metadata.Headers,
			},
			Rows: rows,
			Footer: templates.DataTableFooter{
				PostEndpoint: metadata.Post,
				FormID:       "datatableform",
			},
			Start:       fmt.Sprint(util.Maximum(offset, 0) + 1),
			StartOffset: fmt.Sprint(util.Maximum(offset-limit, 0)),
			End:         fmt.Sprint(offset + viewLimit),
			LastPage:    viewLimit != limit,
		}
		templates.DataTableView(tbl).Render(ctx, w)
	}
}
