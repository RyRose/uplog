package rawdata

import (
	"context"
	"database/sql"
	"net/http"
	"net/url"

	"github.com/RyRose/uplog/internal/service/rawdata/base"
	"github.com/RyRose/uplog/internal/service/rawdata/util"
	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
	"github.com/RyRose/uplog/internal/templates"
)

func HandleGetLiftGroupView(roDB *sql.DB) http.HandlerFunc {
	return base.HandleGetDataTableView(
		roDB,
		base.TableViewMetadata{
			Headers: []string{"ID"},
			Post:    "/view/data/lift_group",
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
	return base.HandlePatchTableRowViewID(
		wDB,
		map[string]base.PatcherID{
			"id": &base.PatchIDParams[workoutdb.RawUpdateLiftGroupIdParams]{
				Query: (*workoutdb.Queries).RawUpdateLiftGroupId,
				Convert: func(id, value string) (*workoutdb.RawUpdateLiftGroupIdParams, error) {
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
	return base.HandlePostDataTableView(
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
