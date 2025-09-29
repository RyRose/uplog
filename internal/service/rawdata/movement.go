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

func HandleGetMovementView(roDB *sql.DB) http.HandlerFunc {
	return base.HandleGetDataTableView(
		roDB,
		base.TableViewMetadata{
			Headers: []string{"ID", "Alias"},
			Post:    "/view/data/movement",
		},
		(*workoutdb.Queries).RawSelectMovementPage,
		func(limit, offset int64) workoutdb.RawSelectMovementPageParams {
			return workoutdb.RawSelectMovementPageParams{
				Limit:  limit,
				Offset: offset,
			}
		},
		func(_ context.Context, _ *sql.DB, movements []workoutdb.Movement) ([]templates.DataTableRow, error) {
			var rows []templates.DataTableRow
			for _, movement := range movements {
				rows = append(rows, templates.DataTableRow{
					PatchEndpoint:  util.UrlPathJoin("/view/data/movement", movement.ID),
					DeleteEndpoint: util.UrlPathJoin("/view/data/movement", movement.ID),
					Values: []templates.DataTableValue{
						{Name: "id", Value: movement.ID, Type: templates.InputString},
						{Name: "alias", Value: movement.Alias, Type: templates.InputString},
					},
				})
			}
			rows = append(rows, templates.DataTableRow{
				Values: []templates.DataTableValue{
					{Name: "id", Type: templates.InputString},
					{Name: "alias", Type: templates.InputString},
				},
			})
			return rows, nil
		},
	)
}

func HandlePatchMovementView(wDB *sql.DB) http.HandlerFunc {
	return base.HandlePatchTableRowViewID(
		wDB,
		map[string]base.PatcherID{
			"id": &base.PatchIDParams[workoutdb.RawUpdateMovementIdParams]{
				Query: (*workoutdb.Queries).RawUpdateMovementId,
				Convert: func(id, value string) (*workoutdb.RawUpdateMovementIdParams, error) {
					return &workoutdb.RawUpdateMovementIdParams{
						In:  id,
						Out: value,
					}, nil
				},
			},
			"alias": &base.PatchIDParams[workoutdb.RawUpdateMovementAliasParams]{
				Query: (*workoutdb.Queries).RawUpdateMovementAlias,
				Convert: func(id, value string) (*workoutdb.RawUpdateMovementAliasParams, error) {
					return &workoutdb.RawUpdateMovementAliasParams{
						ID:    id,
						Alias: value,
					}, nil
				},
			},
		},
	)
}

func HandlePostMovementView(roDB, wDB *sql.DB) http.HandlerFunc {
	return base.HandlePostDataTableView(
		roDB,
		wDB,
		(*workoutdb.Queries).RawInsertMovement,
		func(_ context.Context, values url.Values) (*workoutdb.RawInsertMovementParams, error) {
			return &workoutdb.RawInsertMovementParams{
				ID:    values.Get("id"),
				Alias: values.Get("alias"),
			}, nil
		},
		func(ctx context.Context, q *workoutdb.Queries, movement workoutdb.Movement) (*templates.DataTableRow, error) {
			return &templates.DataTableRow{
				PatchEndpoint:  util.UrlPathJoin("/view/data/movement", movement.ID),
				DeleteEndpoint: util.UrlPathJoin("/view/data/movement", movement.ID),
				Values: []templates.DataTableValue{
					{Name: "id", Value: movement.ID, Type: templates.InputString},
					{Name: "alias", Value: movement.Alias, Type: templates.InputString},
				},
			}, nil
		},
	)
}
