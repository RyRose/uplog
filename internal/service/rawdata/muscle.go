package rawdata

import (
	"context"
	"database/sql"
	"net/http"
	"net/url"

	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
	"github.com/RyRose/uplog/internal/templates"
)

func HandleGetMuscleView(roDB *sql.DB) http.HandlerFunc {
	return handleGetDataTableView(
		roDB,
		tableViewMetadata{
			headers: []string{"ID", "Link", "Message"},
			post:    "/view/data/muscle",
		},
		(*workoutdb.Queries).RawSelectMusclePage,
		func(limit, offset int64) workoutdb.RawSelectMusclePageParams {
			return workoutdb.RawSelectMusclePageParams{
				Limit:  limit,
				Offset: offset,
			}
		},
		func(_ context.Context, _ *sql.DB, items []workoutdb.Muscle) ([]templates.DataTableRow, error) {
			var rows []templates.DataTableRow
			for _, item := range items {
				rows = append(rows, templates.DataTableRow{
					PatchEndpoint:  urlPathJoin("/view/data/muscle", item.ID),
					DeleteEndpoint: urlPathJoin("/view/data/muscle", item.ID),
					Values: []templates.DataTableValue{
						{Name: "id", Value: item.ID, Type: templates.InputString},
						{Name: "link", Value: item.Link, Type: templates.InputString},
						{Name: "message", Value: zero(item.Message), Type: templates.InputString},
					},
				})
			}
			rows = append(rows, templates.DataTableRow{
				Values: []templates.DataTableValue{
					{Name: "id", Type: templates.InputString},
					{Name: "link", Type: templates.InputString},
					{Name: "muscle", Type: templates.InputString},
				},
			})
			return rows, nil
		},
	)
}

func HandlePatchMuscleView(wDB *sql.DB) http.HandlerFunc {
	return handlePatchTableRowViewID(
		wDB,
		map[string]patcherID{
			"id": &patchIDParams[workoutdb.RawUpdateMuscleIdParams]{
				query: (*workoutdb.Queries).RawUpdateMuscleId,
				convert: func(id, value string) (*workoutdb.RawUpdateMuscleIdParams, error) {
					return &workoutdb.RawUpdateMuscleIdParams{
						In:  id,
						Out: value,
					}, nil
				},
			},
			"link": &patchIDParams[workoutdb.RawUpdateMuscleLinkParams]{
				query: (*workoutdb.Queries).RawUpdateMuscleLink,
				convert: func(id, value string) (*workoutdb.RawUpdateMuscleLinkParams, error) {
					return &workoutdb.RawUpdateMuscleLinkParams{
						ID:   id,
						Link: value,
					}, nil
				},
			},
			"message": &patchIDParams[workoutdb.RawUpdateMuscleMessageParams]{
				query: (*workoutdb.Queries).RawUpdateMuscleMessage,
				convert: func(id, value string) (*workoutdb.RawUpdateMuscleMessageParams, error) {
					return &workoutdb.RawUpdateMuscleMessageParams{
						ID:      id,
						Message: deZero(value),
					}, nil
				},
			},
		},
	)
}

func HandlePostMuscleView(roDB, wDB *sql.DB) http.HandlerFunc {
	return handlePostDataTableView(
		roDB,
		wDB,
		(*workoutdb.Queries).RawInsertMuscle,
		func(_ context.Context, values url.Values) (*workoutdb.RawInsertMuscleParams, error) {
			return &workoutdb.RawInsertMuscleParams{
				ID:      values.Get("id"),
				Link:    values.Get("link"),
				Message: deZero(values.Get("message")),
			}, nil
		},
		func(ctx context.Context, q *workoutdb.Queries, movement workoutdb.Muscle) (*templates.DataTableRow, error) {
			return &templates.DataTableRow{
				PatchEndpoint:  urlPathJoin("/view/data/muscle", movement.ID),
				DeleteEndpoint: urlPathJoin("/view/data/muscle", movement.ID),
				Values: []templates.DataTableValue{
					{Name: "id", Value: movement.ID, Type: templates.InputString},
					{Name: "link", Value: movement.Link, Type: templates.InputString},
					{Name: "message", Value: zero(movement.Message), Type: templates.InputString},
				},
			}, nil
		},
	)
}
