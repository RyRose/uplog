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

func HandleGetWorkoutView(roDB *sql.DB) http.HandlerFunc {
	return handleGetDataTableView(
		roDB,
		tableViewMetadata{
			headers: []string{"ID", "Template"},
			post:    "/view/data/workout",
		},
		(*workoutdb.Queries).RawSelectWorkoutPage,
		func(limit, offset int64) workoutdb.RawSelectWorkoutPageParams {
			return workoutdb.RawSelectWorkoutPageParams{
				Limit:  limit,
				Offset: offset,
			}
		},
		func(_ context.Context, _ *sql.DB, items []workoutdb.Workout) ([]templates.DataTableRow, error) {
			var rows []templates.DataTableRow
			for _, item := range items {
				rows = append(rows, templates.DataTableRow{
					PatchEndpoint:  util.UrlPathJoin("/view/data/workout", item.ID),
					DeleteEndpoint: util.UrlPathJoin("/view/data/workout", item.ID),
					Values: []templates.DataTableValue{
						{Name: "id", Type: templates.InputString, Value: item.ID},
						{Name: "template", Type: templates.TextArea, Value: item.Template},
					},
				})
			}
			rows = append(rows, templates.DataTableRow{
				Values: []templates.DataTableValue{
					{Name: "id", Type: templates.InputString},
					{Name: "template", Type: templates.TextArea},
				},
			})
			return rows, nil
		},
	)
}

func HandlePatchWorkoutView(wDB *sql.DB) http.HandlerFunc {
	return handlePatchTableRowViewID(
		wDB,
		map[string]patcherID{
			"id": &patchIDParams[workoutdb.RawUpdateWorkoutIdParams]{
				query: (*workoutdb.Queries).RawUpdateWorkoutId,
				convert: func(id, value string) (*workoutdb.RawUpdateWorkoutIdParams, error) {
					return &workoutdb.RawUpdateWorkoutIdParams{
						In:  id,
						Out: value,
					}, nil
				},
			},
			"template": &patchIDParams[workoutdb.RawUpdateWorkoutTemplateParams]{
				query: (*workoutdb.Queries).RawUpdateWorkoutTemplate,
				convert: func(id, value string) (*workoutdb.RawUpdateWorkoutTemplateParams, error) {
					return &workoutdb.RawUpdateWorkoutTemplateParams{
						ID:       id,
						Template: value,
					}, nil
				},
			},
		},
	)
}

func HandlePostWorkoutView(roDB, wDB *sql.DB) http.HandlerFunc {
	return handlePostDataTableView(
		roDB,
		wDB,
		(*workoutdb.Queries).RawInsertWorkout,
		func(_ context.Context, values url.Values) (*workoutdb.RawInsertWorkoutParams, error) {
			return &workoutdb.RawInsertWorkoutParams{
				ID:       values.Get("id"),
				Template: values.Get("template"),
			}, nil
		},
		func(_ context.Context, _ *workoutdb.Queries, item workoutdb.Workout) (*templates.DataTableRow, error) {
			return &templates.DataTableRow{
				PatchEndpoint:  util.UrlPathJoin("/view/data/workout", item.ID),
				DeleteEndpoint: util.UrlPathJoin("/view/data/workout", item.ID),
				Values: []templates.DataTableValue{
					{Name: "id", Type: templates.InputString, Value: item.ID},
					{Name: "template", Type: templates.TextArea, Value: item.Template},
				},
			}, nil
		},
	)
}
