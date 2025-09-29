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

func HandleGetMovementView(roDB *sql.DB) http.HandlerFunc {
	return handleGetDataTableView(
		roDB,
		tableViewMetadata{
			headers: []string{"ID", "Alias"},
			post:    "/view/data/movement",
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
	return handlePatchTableRowViewID(
		wDB,
		map[string]patcherID{
			"id": &patchIDParams[workoutdb.RawUpdateMovementIdParams]{
				query: (*workoutdb.Queries).RawUpdateMovementId,
				convert: func(id, value string) (*workoutdb.RawUpdateMovementIdParams, error) {
					return &workoutdb.RawUpdateMovementIdParams{
						In:  id,
						Out: value,
					}, nil
				},
			},
			"alias": &patchIDParams[workoutdb.RawUpdateMovementAliasParams]{
				query: (*workoutdb.Queries).RawUpdateMovementAlias,
				convert: func(id, value string) (*workoutdb.RawUpdateMovementAliasParams, error) {
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
	return handlePostDataTableView(
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
