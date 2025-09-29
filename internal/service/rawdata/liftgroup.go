package rawdata

import (
	"context"
	"database/sql"
	"net/http"
	"net/url"

	"github.com/RyRose/uplog/internal/service/rawdata/util"
	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
	"github.com/RyRose/uplog/internal/templates"
)

func HandleGetLiftGroupView(roDB *sql.DB) http.HandlerFunc {
	return handleGetDataTableView(
		roDB,
		tableViewMetadata{
			headers: []string{"ID"},
			post:    "/view/data/lift_group",
		},
		(*workoutdb.Queries).RawSelectLiftGroupPage,
		func(limit, offset int64) workoutdb.RawSelectLiftGroupPageParams {
			return workoutdb.RawSelectLiftGroupPageParams{
				Limit:  limit,
				Offset: offset,
			}
		},
		func(_ context.Context, _ *sql.DB, liftGroups []string) ([]templates.DataTableRow, error) {
			var rows []templates.DataTableRow
			for _, liftGroup := range liftGroups {
				rows = append(rows, templates.DataTableRow{
					PatchEndpoint:  util.UrlPathJoin("/view/data/lift_group", liftGroup),
					DeleteEndpoint: util.UrlPathJoin("/view/data/lift_group", liftGroup),
					Values: []templates.DataTableValue{
						{Name: "id", Value: liftGroup, Type: templates.InputString},
					},
				})
			}
			rows = append(rows, templates.DataTableRow{
				Values: []templates.DataTableValue{
					{Name: "id", Type: templates.InputString},
				},
			})
			return rows, nil
		},
	)
}

func HandlePatchLiftGroupView(wDB *sql.DB) http.HandlerFunc {
	return handlePatchTableRowViewID(
		wDB,
		map[string]patcherID{
			"id": &patchIDParams[workoutdb.RawUpdateLiftGroupIdParams]{
				query: (*workoutdb.Queries).RawUpdateLiftGroupId,
				convert: func(id, value string) (*workoutdb.RawUpdateLiftGroupIdParams, error) {
					return &workoutdb.RawUpdateLiftGroupIdParams{
						In:  id,
						Out: value,
					}, nil
				},
			},
		},
	)
}

func HandlePostLiftGroupView(roDB, wDB *sql.DB) http.HandlerFunc {
	return handlePostDataTableView(
		roDB,
		wDB,
		(*workoutdb.Queries).RawInsertLiftGroup,
		func(_ context.Context, values url.Values) (*string, error) {
			id := values.Get("id")
			return &id, nil
		},
		func(ctx context.Context, q *workoutdb.Queries, liftGroup string) (*templates.DataTableRow, error) {
			return &templates.DataTableRow{
				PatchEndpoint:  util.UrlPathJoin("/view/data/lift_group", liftGroup),
				DeleteEndpoint: util.UrlPathJoin("/view/data/lift_group", liftGroup),
				Values: []templates.DataTableValue{
					{Name: "id", Value: liftGroup, Type: templates.InputString},
				},
			}, nil
		},
	)
}
