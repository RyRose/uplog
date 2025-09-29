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

func HandleGetScheduleView(roDB *sql.DB) http.HandlerFunc {
	return handleGetDataTableView(
		roDB,
		tableViewMetadata{
			headers: []string{"Date", "Workout"},
			post:    "/view/data/schedule",
		},
		(*workoutdb.Queries).RawSelectSchedulePage,
		func(limit, offset int64) workoutdb.RawSelectSchedulePageParams {
			return workoutdb.RawSelectSchedulePageParams{
				Limit:  limit,
				Offset: offset,
			}
		},
		func(ctx context.Context, roDB *sql.DB, items []workoutdb.Schedule) ([]templates.DataTableRow, error) {
			q := workoutdb.New(roDB)
			workouts, err := q.ListAllIndividualWorkouts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list workouts: %w", err)
			}
			var rows []templates.DataTableRow
			for _, item := range items {
				workout, _ := item.Workout.(string)
				rows = append(rows, templates.DataTableRow{
					PatchEndpoint:  urlPathJoin("/view/data/schedule", item.Date),
					DeleteEndpoint: urlPathJoin("/view/data/schedule", item.Date),
					Values: []templates.DataTableValue{
						{Name: "date", Type: templates.InputString, Value: item.Date},
						{Name: "workout", Type: templates.Select, Value: workout, SelectOptions: workouts},
					},
				})
			}
			rows = append(rows, templates.DataTableRow{
				Values: []templates.DataTableValue{
					{Name: "date", Type: templates.InputString},
					{Name: "workout", Type: templates.Select, SelectOptions: workouts},
				},
			})
			return rows, nil
		},
	)
}

func HandlePatchScheduleView(wDB *sql.DB) http.HandlerFunc {
	return handlePatchTableRowViewID(
		wDB,
		map[string]patcherID{
			"date": &patchIDParams[workoutdb.RawUpdateScheduleDateParams]{
				query: (*workoutdb.Queries).RawUpdateScheduleDate,
				convert: func(date, value string) (*workoutdb.RawUpdateScheduleDateParams, error) {
					return &workoutdb.RawUpdateScheduleDateParams{
						In:  date,
						Out: value,
					}, nil
				},
			},
			"workout": &patchIDParams[workoutdb.RawUpdateScheduleWorkoutParams]{
				query: (*workoutdb.Queries).RawUpdateScheduleWorkout,
				convert: func(id, value string) (*workoutdb.RawUpdateScheduleWorkoutParams, error) {
					return &workoutdb.RawUpdateScheduleWorkoutParams{
						Date:    id,
						Workout: deZero(value),
					}, nil
				},
			},
		},
	)
}

func HandlePostScheduleView(roDB, wDB *sql.DB) http.HandlerFunc {
	return handlePostDataTableView(
		roDB,
		wDB,
		(*workoutdb.Queries).RawInsertSchedule,
		func(_ context.Context, values url.Values) (*workoutdb.RawInsertScheduleParams, error) {
			return &workoutdb.RawInsertScheduleParams{
				Date:    values.Get("date"),
				Workout: deZero(values.Get("workout")),
			}, nil
		},
		func(ctx context.Context, q *workoutdb.Queries, item workoutdb.Schedule) (*templates.DataTableRow, error) {
			workouts, err := q.ListAllIndividualWorkouts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list workouts: %w", err)
			}

			workout, _ := item.Workout.(string)
			return &templates.DataTableRow{
				PatchEndpoint:  urlPathJoin("/view/data/schedule", item.Date),
				DeleteEndpoint: urlPathJoin("/view/data/schedule", item.Date),
				Values: []templates.DataTableValue{
					{Name: "date", Type: templates.InputString, Value: item.Date},
					{Name: "workout", Type: templates.Select, Value: workout, SelectOptions: append(workouts, "")},
				},
			}, nil
		},
	)
}
