package rawdata

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
	"github.com/RyRose/uplog/internal/templates"
)

func HandleGetScheduleListView(roDB *sql.DB) http.HandlerFunc {
	return handleGetDataTableView(
		roDB,
		tableViewMetadata{
			headers: []string{"ID", "Day", "Workout"},
			post:    "/view/data/schedule_list",
		},
		(*workoutdb.Queries).RawSelectScheduleListPage,
		func(limit, offset int64) workoutdb.RawSelectScheduleListPageParams {
			return workoutdb.RawSelectScheduleListPageParams{
				Limit:  limit,
				Offset: offset,
			}
		},
		func(ctx context.Context, roDB *sql.DB, items []workoutdb.ScheduleList) ([]templates.DataTableRow, error) {
			q := workoutdb.New(roDB)
			workouts, err := q.ListAllIndividualWorkouts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list workouts: %w", err)
			}

			var rows []templates.DataTableRow
			for _, item := range items {
				rows = append(rows, templates.DataTableRow{
					PatchEndpoint:  urlPathJoin("/view/data/schedule_list", item.ID),
					DeleteEndpoint: urlPathJoin("/view/data/schedule_list", item.ID),
					Values: []templates.DataTableValue{
						{Name: "id", Value: item.ID, Type: templates.InputString},
						{Name: "day", Value: fmt.Sprint(item.Day), Type: templates.InputNumber},
						{Name: "workout", Value: item.Workout, Type: templates.Select, SelectOptions: workouts},
					},
				})
			}
			rows = append(rows, templates.DataTableRow{
				Values: []templates.DataTableValue{
					{Name: "id", Type: templates.InputString},
					{Name: "day", Type: templates.InputNumber},
					{Name: "workout", Type: templates.Select, SelectOptions: workouts},
				},
			})
			return rows, nil
		},
	)
}

func HandlePatchScheduleListView(wDB *sql.DB) http.HandlerFunc {
	return handlePatchTableRowViewID(
		wDB,
		map[string]patcherID{
			"id": &patchIDParams[workoutdb.RawUpdateScheduleListIdParams]{
				query: (*workoutdb.Queries).RawUpdateScheduleListId,
				convert: func(id, value string) (*workoutdb.RawUpdateScheduleListIdParams, error) {
					return &workoutdb.RawUpdateScheduleListIdParams{
						In:  id,
						Out: value,
					}, nil
				},
			},
			"day": &patchIDParams[workoutdb.RawUpdateScheduleListDayParams]{
				query: (*workoutdb.Queries).RawUpdateScheduleListDay,
				convert: func(id, value string) (*workoutdb.RawUpdateScheduleListDayParams, error) {
					v, err := strconv.ParseInt(value, 10, 64)
					if err != nil {
						return nil, fmt.Errorf("failed to parse day: %w", err)
					}
					return &workoutdb.RawUpdateScheduleListDayParams{
						ID:  id,
						Day: v,
					}, nil
				},
			},
			"workout": &patchIDParams[workoutdb.RawUpdateScheduleListWorkoutParams]{
				query: (*workoutdb.Queries).RawUpdateScheduleListWorkout,
				convert: func(id, value string) (*workoutdb.RawUpdateScheduleListWorkoutParams, error) {
					return &workoutdb.RawUpdateScheduleListWorkoutParams{
						ID:      id,
						Workout: value,
					}, nil
				},
			},
		},
	)
}

func HandlePostScheduleListView(roDB, wDB *sql.DB) http.HandlerFunc {
	return handlePostDataTableView(
		roDB,
		wDB,
		(*workoutdb.Queries).RawInsertScheduleList,
		func(_ context.Context, values url.Values) (*workoutdb.RawInsertScheduleListParams, error) {
			day, err := strconv.ParseInt(values.Get("day"), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse day: %w", err)
			}
			return &workoutdb.RawInsertScheduleListParams{
				ID:      values.Get("id"),
				Day:     day,
				Workout: values.Get("workout"),
			}, nil
		},
		func(ctx context.Context, q *workoutdb.Queries, item workoutdb.ScheduleList) (*templates.DataTableRow, error) {
			workouts, err := q.ListAllIndividualWorkouts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list workouts: %w", err)
			}

			return &templates.DataTableRow{
				PatchEndpoint:  urlPathJoin("/view/data/schedule_list", item.ID),
				DeleteEndpoint: urlPathJoin("/view/data/schedule_list", item.ID),
				Values: []templates.DataTableValue{
					{Name: "id", Value: item.ID, Type: templates.InputString},
					{Name: "day", Value: fmt.Sprint(item.Day), Type: templates.InputString},
					{Name: "workout", Value: item.Workout, Type: templates.Select, SelectOptions: workouts},
				},
			}, nil
		},
	)
}
