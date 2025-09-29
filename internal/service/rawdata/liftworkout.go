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

func HandleGetLiftWorkoutView(roDB *sql.DB) http.HandlerFunc {
	return handleGetDataTableView(
		roDB,
		tableViewMetadata{
			headers: []string{"Lift", "Workout"},
			post:    "/view/data/lift_workout_mapping",
		},
		(*workoutdb.Queries).RawSelectLiftWorkoutPage,
		func(limit, offset int64) workoutdb.RawSelectLiftWorkoutPageParams {
			return workoutdb.RawSelectLiftWorkoutPageParams{
				Limit:  limit,
				Offset: offset,
			}
		},
		func(ctx context.Context, roDB *sql.DB, items []workoutdb.LiftWorkoutMapping) ([]templates.DataTableRow, error) {
			q := workoutdb.New(roDB)
			lifts, err := q.ListAllIndividualLifts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list lifts: %w", err)
			}
			workouts, err := q.ListAllIndividualWorkouts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list workouts: %w", err)
			}
			var rows []templates.DataTableRow
			for _, item := range items {
				rows = append(rows, templates.DataTableRow{
					PatchEndpoint:  urlPathJoin("/view/data/lift_workout_mapping", item.Lift, item.Workout),
					DeleteEndpoint: urlPathJoin("/view/data/lift_workout_mapping", item.Lift, item.Workout),
					Values: []templates.DataTableValue{
						{Name: "lift", Type: templates.Select, Value: item.Lift, SelectOptions: lifts},
						{Name: "workout", Type: templates.Select, Value: item.Workout, SelectOptions: workouts},
					},
				})
			}
			rows = append(rows, templates.DataTableRow{
				Values: []templates.DataTableValue{
					{Name: "lift", Type: templates.Select, SelectOptions: lifts},
					{Name: "workout", Type: templates.Select, SelectOptions: workouts},
				},
			})
			return rows, nil
		},
	)
}

func HandlePatchLiftWorkoutView(wDB *sql.DB) http.HandlerFunc {
	return handlePatchTableRowViewRequest(
		wDB,
		map[string]patcherReq{
			"lift": &patchReqParams[workoutdb.RawUpdateLiftWorkoutMappingLiftParams]{
				query: (*workoutdb.Queries).RawUpdateLiftWorkoutMappingLift,
				convert: func(r *http.Request, value string) (*workoutdb.RawUpdateLiftWorkoutMappingLiftParams, error) {
					return &workoutdb.RawUpdateLiftWorkoutMappingLiftParams{
						Out:     value,
						In:      r.PathValue("lift"),
						Workout: r.PathValue("workout"),
					}, nil
				},
			},
			"workout": &patchReqParams[workoutdb.RawUpdateLiftWorkoutMappingWorkoutParams]{
				query: (*workoutdb.Queries).RawUpdateLiftWorkoutMappingWorkout,
				convert: func(r *http.Request, value string) (*workoutdb.RawUpdateLiftWorkoutMappingWorkoutParams, error) {
					return &workoutdb.RawUpdateLiftWorkoutMappingWorkoutParams{
						Out:  value,
						In:   r.PathValue("workout"),
						Lift: r.PathValue("lift"),
					}, nil
				},
			},
		},
	)
}

func HandlePostLiftWorkoutView(roDB, wDB *sql.DB) http.HandlerFunc {
	return handlePostDataTableView(
		roDB,
		wDB,
		(*workoutdb.Queries).RawInsertLiftWorkout,
		func(_ context.Context, values url.Values) (*workoutdb.RawInsertLiftWorkoutParams, error) {
			return &workoutdb.RawInsertLiftWorkoutParams{
				Lift:    values.Get("lift"),
				Workout: values.Get("workout"),
			}, nil
		},
		func(ctx context.Context, q *workoutdb.Queries, item workoutdb.LiftWorkoutMapping) (*templates.DataTableRow, error) {
			lifts, err := q.ListAllIndividualLifts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list lifts: %w", err)
			}
			workouts, err := q.ListAllIndividualWorkouts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list workouts: %w", err)
			}

			return &templates.DataTableRow{
				PatchEndpoint:  urlPathJoin("/view/data/lift_workout_mapping", item.Lift, item.Workout),
				DeleteEndpoint: urlPathJoin("/view/data/lift_workout_mapping", item.Lift, item.Workout),
				Values: []templates.DataTableValue{
					{Name: "lift", Type: templates.Select, Value: item.Lift, SelectOptions: lifts},
					{Name: "workout", Type: templates.Select, Value: item.Workout, SelectOptions: workouts},
				},
			}, nil
		},
	)
}
