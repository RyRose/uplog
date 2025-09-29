package rawdata

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"

	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
	"github.com/RyRose/uplog/internal/templates"
)

func HandleGetSubworkoutView(roDB *sql.DB) http.HandlerFunc {
	return handleGetDataTableView(
		roDB,
		tableViewMetadata{
			headers: []string{"Subworkout", "Superworkout"},
			post:    "/view/data/subworkout",
		},
		(*workoutdb.Queries).RawSelectSubworkoutPage,
		func(limit, offset int64) workoutdb.RawSelectSubworkoutPageParams {
			return workoutdb.RawSelectSubworkoutPageParams{
				Limit:  limit,
				Offset: offset,
			}
		},
		func(ctx context.Context, roDB *sql.DB, items []workoutdb.Subworkout) ([]templates.DataTableRow, error) {
			q := workoutdb.New(roDB)
			workouts, err := q.ListAllIndividualWorkouts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list workouts: %w", err)
			}
			var rows []templates.DataTableRow
			for _, item := range items {
				rows = append(rows, templates.DataTableRow{
					PatchEndpoint:  urlPathJoin("/view/data/subworkout", item.Subworkout, item.Superworkout),
					DeleteEndpoint: urlPathJoin("/view/data/subworkout", item.Subworkout, item.Superworkout),
					Values: []templates.DataTableValue{
						{Name: "subworkout", Type: templates.Select, Value: item.Subworkout, SelectOptions: workouts},
						{Name: "superworkout", Type: templates.Select, Value: item.Superworkout, SelectOptions: workouts},
					},
				})
			}
			rows = append(rows, templates.DataTableRow{
				Values: []templates.DataTableValue{
					{Name: "subworkout", Type: templates.Select, SelectOptions: workouts},
					{Name: "superworkout", Type: templates.Select, SelectOptions: workouts},
				},
			})
			return rows, nil
		},
	)
}

func HandlePatchSubworkoutView(wDB *sql.DB) http.HandlerFunc {
	return handlePatchTableRowViewRequest(
		wDB,
		map[string]patcherReq{
			"subworkout": &patchReqParams[workoutdb.RawUpdateSubworkoutSubworkoutParams]{
				query: (*workoutdb.Queries).RawUpdateSubworkoutSubworkout,
				convert: func(r *http.Request, value string) (*workoutdb.RawUpdateSubworkoutSubworkoutParams, error) {
					return &workoutdb.RawUpdateSubworkoutSubworkoutParams{
						Out:          value,
						In:           r.PathValue("subworkout"),
						Superworkout: r.PathValue("superworkout"),
					}, nil
				},
			},
			"superworkout": &patchReqParams[workoutdb.RawUpdateSubworkoutSuperworkoutParams]{
				query: (*workoutdb.Queries).RawUpdateSubworkoutSuperworkout,
				convert: func(r *http.Request, value string) (*workoutdb.RawUpdateSubworkoutSuperworkoutParams, error) {
					return &workoutdb.RawUpdateSubworkoutSuperworkoutParams{
						Out:        value,
						In:         r.PathValue("superworkout"),
						Subworkout: r.PathValue("subworkout"),
					}, nil
				},
			},
		},
	)
}

func HandlePostSubworkoutView(roDB, wDB *sql.DB) http.HandlerFunc {
	return handlePostDataTableView(
		roDB,
		wDB,
		(*workoutdb.Queries).RawInsertSubworkout,
		func(_ context.Context, values url.Values) (*workoutdb.RawInsertSubworkoutParams, error) {
			return &workoutdb.RawInsertSubworkoutParams{
				Subworkout:   values.Get("subworkout"),
				Superworkout: values.Get("superworkout"),
			}, nil
		},
		func(ctx context.Context, q *workoutdb.Queries, item workoutdb.Subworkout) (*templates.DataTableRow, error) {
			workouts, err := q.ListAllIndividualWorkouts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list workouts: %w", err)
			}

			return &templates.DataTableRow{
				PatchEndpoint:  urlPathJoin("/view/data/subworkout", item.Subworkout, item.Superworkout),
				DeleteEndpoint: urlPathJoin("/view/data/subworkout", item.Subworkout, item.Superworkout),
				Values: []templates.DataTableValue{
					{Name: "subworkout", Type: templates.Select, Value: item.Subworkout, SelectOptions: workouts},
					{Name: "superworkout", Type: templates.Select, Value: item.Superworkout, SelectOptions: workouts},
				},
			}, nil
		},
	)
}
