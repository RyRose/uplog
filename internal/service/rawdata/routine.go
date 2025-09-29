package rawdata

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"

	"github.com/RyRose/uplog/internal/service/rawdata/base"
	"github.com/RyRose/uplog/internal/service/rawdata/util"
	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
	"github.com/RyRose/uplog/internal/templates"
)

func HandleGetRoutineView(roDB *sql.DB) http.HandlerFunc {
	return base.HandleGetDataTableView(
		roDB,
		base.TableViewMetadata{
			Headers: []string{"ID", "Steps", "Lift"},
			Post:    "/view/data/routine",
		},
		(*workoutdb.Queries).RawSelectRoutinePage,
		func(limit, offset int64) workoutdb.RawSelectRoutinePageParams {
			return workoutdb.RawSelectRoutinePageParams{
				Limit:  limit,
				Offset: offset,
			}
		},
		func(ctx context.Context, roDB *sql.DB, items []workoutdb.Routine) ([]templates.DataTableRow, error) {
			q := workoutdb.New(roDB)
			lifts, err := q.ListAllIndividualLifts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list lifts: %w", err)
			}

			var rows []templates.DataTableRow
			for _, item := range items {
				rows = append(rows, templates.DataTableRow{
					PatchEndpoint:  util.UrlPathJoin("/view/data/routine", item.ID),
					DeleteEndpoint: util.UrlPathJoin("/view/data/routine", item.ID),
					Values: []templates.DataTableValue{
						{Name: "id", Value: item.ID, Type: templates.InputString},
						{Name: "steps", Value: item.Steps, Type: templates.InputString},
						{Name: "lift", Value: item.Lift, Type: templates.Select, SelectOptions: lifts},
					},
				})
			}
			rows = append(rows, templates.DataTableRow{
				Values: []templates.DataTableValue{
					{Name: "id", Type: templates.InputString},
					{Name: "link", Type: templates.InputString},
					{Name: "lift", Type: templates.Select, SelectOptions: lifts},
				},
			})
			return rows, nil
		},
	)
}

func HandlePatchRoutineView(wDB *sql.DB) http.HandlerFunc {
	return base.HandlePatchTableRowViewID(
		wDB,
		map[string]base.PatcherID{
			"id": &base.PatchIDParams[workoutdb.RawUpdateRoutineIdParams]{
				Query: (*workoutdb.Queries).RawUpdateRoutineId,
				Convert: func(id, value string) (*workoutdb.RawUpdateRoutineIdParams, error) {
					return &workoutdb.RawUpdateRoutineIdParams{
						In:  id,
						Out: value,
					}, nil
				},
			},
			"steps": &base.PatchIDParams[workoutdb.RawUpdateRoutineStepsParams]{
				Query: (*workoutdb.Queries).RawUpdateRoutineSteps,
				Convert: func(id, value string) (*workoutdb.RawUpdateRoutineStepsParams, error) {
					return &workoutdb.RawUpdateRoutineStepsParams{
						ID:    id,
						Steps: value,
					}, nil
				},
			},
			"lift": &base.PatchIDParams[workoutdb.RawUpdateRoutineLiftParams]{
				Query: (*workoutdb.Queries).RawUpdateRoutineLift,
				Convert: func(id, value string) (*workoutdb.RawUpdateRoutineLiftParams, error) {
					return &workoutdb.RawUpdateRoutineLiftParams{
						ID:   id,
						Lift: value,
					}, nil
				},
			},
		},
	)
}

func HandlePostRoutineView(roDB, wDB *sql.DB) http.HandlerFunc {
	return base.HandlePostDataTableView(
		roDB,
		wDB,
		(*workoutdb.Queries).RawInsertRoutine,
		func(_ context.Context, values url.Values) (*workoutdb.RawInsertRoutineParams, error) {
			return &workoutdb.RawInsertRoutineParams{
				ID:    values.Get("id"),
				Steps: values.Get("steps"),
				Lift:  values.Get("lift"),
			}, nil
		},
		func(ctx context.Context, q *workoutdb.Queries, item workoutdb.Routine) (*templates.DataTableRow, error) {
			lifts, err := q.ListAllIndividualLifts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list lifts: %w", err)
			}

			return &templates.DataTableRow{
				PatchEndpoint:  util.UrlPathJoin("/view/data/routine", item.ID),
				DeleteEndpoint: util.UrlPathJoin("/view/data/routine", item.ID),
				Values: []templates.DataTableValue{
					{Name: "id", Value: item.ID, Type: templates.InputString},
					{Name: "steps", Value: item.Steps, Type: templates.InputString},
					{Name: "lift", Value: item.Lift, Type: templates.Select, SelectOptions: lifts},
				},
			}, nil
		},
	)
}
