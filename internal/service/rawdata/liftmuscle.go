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

func HandleGetLiftMuscleView(roDB *sql.DB) http.HandlerFunc {
	return handleGetDataTableView(
		roDB,
		tableViewMetadata{
			headers: []string{"Lift", "Muscle", "Movement"},
			post:    "/view/data/lift_muscle_mapping",
		},
		(*workoutdb.Queries).RawSelectLiftMusclePage,
		func(limit, offset int64) workoutdb.RawSelectLiftMusclePageParams {
			return workoutdb.RawSelectLiftMusclePageParams{
				Limit:  limit,
				Offset: offset,
			}
		},
		func(ctx context.Context, roDB *sql.DB, items []workoutdb.LiftMuscleMapping) ([]templates.DataTableRow, error) {
			q := workoutdb.New(roDB)
			lifts, err := q.ListAllIndividualLifts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list lifts: %w", err)
			}
			muscles, err := q.ListAllIndividualMuscles(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list muscles: %w", err)
			}
			movements, err := q.ListAllIndividualMovements(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list movements: %w", err)
			}
			var rows []templates.DataTableRow
			for _, item := range items {
				rows = append(rows, templates.DataTableRow{
					PatchEndpoint:  urlPathJoin("/view/data/lift_muscle_mapping", item.Lift, item.Muscle, item.Movement),
					DeleteEndpoint: urlPathJoin("/view/data/lift_muscle_mapping", item.Lift, item.Muscle, item.Movement),
					Values: []templates.DataTableValue{
						{Name: "lift", Type: templates.Select, Value: item.Lift, SelectOptions: lifts},
						{Name: "muscle", Type: templates.Select, Value: item.Muscle, SelectOptions: muscles},
						{Name: "movement", Type: templates.Select, Value: item.Movement, SelectOptions: movements},
					},
				})
			}
			rows = append(rows, templates.DataTableRow{
				Values: []templates.DataTableValue{
					{Name: "lift", Type: templates.Select, SelectOptions: lifts},
					{Name: "muscle", Type: templates.Select, SelectOptions: muscles},
					{Name: "movement", Type: templates.Select, SelectOptions: movements},
				},
			})
			return rows, nil
		},
	)
}

func HandlePatchLiftMuscleView(wDB *sql.DB) http.HandlerFunc {
	return handlePatchTableRowViewRequest(
		wDB,
		map[string]patcherReq{
			"lift": &patchReqParams[workoutdb.RawUpdateLiftMuscleMappingLiftParams]{
				query: (*workoutdb.Queries).RawUpdateLiftMuscleMappingLift,
				convert: func(r *http.Request, value string) (*workoutdb.RawUpdateLiftMuscleMappingLiftParams, error) {
					return &workoutdb.RawUpdateLiftMuscleMappingLiftParams{
						Out:      value,
						In:       r.PathValue("lift"),
						Muscle:   r.PathValue("muscle"),
						Movement: r.PathValue("movement"),
					}, nil
				},
			},
			"muscle": &patchReqParams[workoutdb.RawUpdateLiftMuscleMappingMuscleParams]{
				query: (*workoutdb.Queries).RawUpdateLiftMuscleMappingMuscle,
				convert: func(r *http.Request, value string) (*workoutdb.RawUpdateLiftMuscleMappingMuscleParams, error) {
					return &workoutdb.RawUpdateLiftMuscleMappingMuscleParams{
						Out:      value,
						In:       r.PathValue("muscle"),
						Lift:     r.PathValue("lift"),
						Movement: r.PathValue("movement"),
					}, nil
				},
			},
			"movement": &patchReqParams[workoutdb.RawUpdateLiftMuscleMappingMovementParams]{
				query: (*workoutdb.Queries).RawUpdateLiftMuscleMappingMovement,
				convert: func(r *http.Request, value string) (*workoutdb.RawUpdateLiftMuscleMappingMovementParams, error) {
					return &workoutdb.RawUpdateLiftMuscleMappingMovementParams{
						Out:    value,
						In:     r.PathValue("movement"),
						Lift:   r.PathValue("lift"),
						Muscle: r.PathValue("muscle"),
					}, nil
				},
			},
		},
	)
}

func HandlePostLiftMuscleView(roDB, wDB *sql.DB) http.HandlerFunc {
	return handlePostDataTableView(
		roDB,
		wDB,
		(*workoutdb.Queries).RawInsertLiftMuscle,
		func(_ context.Context, values url.Values) (*workoutdb.RawInsertLiftMuscleParams, error) {
			return &workoutdb.RawInsertLiftMuscleParams{
				Lift:     values.Get("lift"),
				Muscle:   values.Get("muscle"),
				Movement: values.Get("movement"),
			}, nil
		},
		func(ctx context.Context, q *workoutdb.Queries, item workoutdb.LiftMuscleMapping) (*templates.DataTableRow, error) {
			lifts, err := q.ListAllIndividualLifts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list lifts: %w", err)
			}
			muscles, err := q.ListAllIndividualMuscles(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list muscles: %w", err)
			}
			movements, err := q.ListAllIndividualMovements(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list movements: %w", err)
			}

			return &templates.DataTableRow{
				PatchEndpoint:  urlPathJoin("/view/data/lift_muscle_mapping", item.Lift, item.Muscle, item.Movement),
				DeleteEndpoint: urlPathJoin("/view/data/lift_muscle_mapping", item.Lift, item.Muscle, item.Movement),
				Values: []templates.DataTableValue{
					{Name: "lift", Type: templates.Select, Value: item.Lift, SelectOptions: lifts},
					{Name: "muscle", Type: templates.Select, Value: item.Muscle, SelectOptions: muscles},
					{Name: "movement", Type: templates.Select, Value: item.Movement, SelectOptions: movements},
				},
			}, nil
		},
	)
}
