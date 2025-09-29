package rawdata

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/RyRose/uplog/internal/service/rawdata/base"
	"github.com/RyRose/uplog/internal/service/rawdata/util"
	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
	"github.com/RyRose/uplog/internal/templates"
)

func HandleGetProgressView(roDB *sql.DB) http.HandlerFunc {
	return base.HandleGetDataTableView(
		roDB,
		base.TableViewMetadata{
			Headers: []string{"ID", "Lift", "Date", "Weight", "Sets", "Reps", "SW"},
			Post:    "/view/data/progress",
		},
		(*workoutdb.Queries).RawSelectProgressPage,
		func(limit, offset int64) workoutdb.RawSelectProgressPageParams {
			return workoutdb.RawSelectProgressPageParams{
				Limit:  limit,
				Offset: offset,
			}
		},
		func(ctx context.Context, roDB *sql.DB, items []workoutdb.Progress) ([]templates.DataTableRow, error) {
			q := workoutdb.New(roDB)
			lifts, err := q.ListAllIndividualLifts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list lifts: %w", err)
			}
			sws, err := q.ListAllIndividualSideWeights(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list side weights: %w", err)
			}
			var rows []templates.DataTableRow
			for _, item := range items {
				sw, _ := item.SideWeight.(string)
				rows = append(rows, templates.DataTableRow{
					PatchEndpoint:  util.UrlPathJoin("/view/data/progress", fmt.Sprint(item.ID)),
					DeleteEndpoint: util.UrlPathJoin("/view/data/progress", fmt.Sprint(item.ID)),
					Values: []templates.DataTableValue{
						{Name: "id", Type: templates.Static, Value: fmt.Sprint(item.ID)},
						{Name: "lift", Type: templates.Select, Value: item.Lift, SelectOptions: lifts},
						{Name: "date", Type: templates.InputString, Value: item.Date},
						{Name: "weight", Type: templates.InputNumber, Value: fmt.Sprint(item.Weight)},
						{Name: "sets", Type: templates.InputNumber, Value: fmt.Sprint(item.Sets)},
						{Name: "reps", Type: templates.InputNumber, Value: fmt.Sprint(item.Reps)},
						{Name: "side_weight", Type: templates.Select, Value: sw, SelectOptions: append(sws, "")},
					},
				})
			}
			rows = append(rows, templates.DataTableRow{
				Values: []templates.DataTableValue{
					{Name: "id", Type: templates.Static},
					{Name: "lift", Type: templates.Select, SelectOptions: lifts},
					{Name: "date", Type: templates.InputString},
					{Name: "weight", Type: templates.InputNumber},
					{Name: "sets", Type: templates.InputNumber},
					{Name: "reps", Type: templates.InputNumber},
					{Name: "side_weight", Type: templates.Select, SelectOptions: append(sws, "")},
				},
			})
			return rows, nil
		},
	)
}

func HandlePatchProgressView(wDB *sql.DB) http.HandlerFunc {
	return base.HandlePatchTableRowViewID(
		wDB,
		map[string]base.PatcherID{
			"lift": &base.PatchIDParams[workoutdb.RawUpdateProgressLiftParams]{
				Query: (*workoutdb.Queries).RawUpdateProgressLift,
				Convert: func(id, value string) (*workoutdb.RawUpdateProgressLiftParams, error) {
					idN, err := strconv.ParseInt(id, 10, 64)
					if err != nil {
						return nil, fmt.Errorf("failed to parse id: %w", err)
					}
					return &workoutdb.RawUpdateProgressLiftParams{
						ID:   idN,
						Lift: value,
					}, nil
				},
			},
			"date": &base.PatchIDParams[workoutdb.RawUpdateProgressDateParams]{
				Query: (*workoutdb.Queries).RawUpdateProgressDate,
				Convert: func(id, value string) (*workoutdb.RawUpdateProgressDateParams, error) {
					idN, err := strconv.ParseInt(id, 10, 64)
					if err != nil {
						return nil, fmt.Errorf("failed to parse id: %w", err)
					}
					return &workoutdb.RawUpdateProgressDateParams{
						ID:   idN,
						Date: value,
					}, nil
				},
			},
			"weight": &base.PatchIDParams[workoutdb.RawUpdateProgressWeightParams]{
				Query: (*workoutdb.Queries).RawUpdateProgressWeight,
				Convert: func(id, value string) (*workoutdb.RawUpdateProgressWeightParams, error) {
					idN, err := strconv.ParseInt(id, 10, 64)
					if err != nil {
						return nil, fmt.Errorf("failed to parse id: %w", err)
					}
					weight, err := strconv.ParseFloat(value, 64)
					if err != nil {
						return nil, fmt.Errorf("failed to parse weight: %w", err)
					}
					return &workoutdb.RawUpdateProgressWeightParams{
						ID:     idN,
						Weight: weight,
					}, nil
				},
			},
			"sets": &base.PatchIDParams[workoutdb.RawUpdateProgressSetsParams]{
				Query: (*workoutdb.Queries).RawUpdateProgressSets,
				Convert: func(id, value string) (*workoutdb.RawUpdateProgressSetsParams, error) {
					idN, err := strconv.ParseInt(id, 10, 64)
					if err != nil {
						return nil, fmt.Errorf("failed to parse id: %w", err)
					}
					sets, err := strconv.ParseInt(value, 10, 64)
					if err != nil {
						return nil, fmt.Errorf("failed to parse sets: %w", err)
					}
					return &workoutdb.RawUpdateProgressSetsParams{
						ID:   idN,
						Sets: sets,
					}, nil
				},
			},
			"reps": &base.PatchIDParams[workoutdb.RawUpdateProgressRepsParams]{
				Query: (*workoutdb.Queries).RawUpdateProgressReps,
				Convert: func(id, value string) (*workoutdb.RawUpdateProgressRepsParams, error) {
					idN, err := strconv.ParseInt(id, 10, 64)
					if err != nil {
						return nil, fmt.Errorf("failed to parse id: %w", err)
					}
					reps, err := strconv.ParseInt(value, 10, 64)
					if err != nil {
						return nil, fmt.Errorf("failed to parse reps: %w", err)
					}
					return &workoutdb.RawUpdateProgressRepsParams{
						ID:   idN,
						Reps: reps,
					}, nil
				},
			},
			"side_weight": &base.PatchIDParams[workoutdb.RawUpdateProgressSideWeightParams]{
				Query: (*workoutdb.Queries).RawUpdateProgressSideWeight,
				Convert: func(id, value string) (*workoutdb.RawUpdateProgressSideWeightParams, error) {
					idN, err := strconv.ParseInt(id, 10, 64)
					if err != nil {
						return nil, fmt.Errorf("failed to parse id: %w", err)
					}
					return &workoutdb.RawUpdateProgressSideWeightParams{
						ID:         idN,
						SideWeight: value,
					}, nil
				},
			},
		},
	)
}

func HandlePostProgressView(roDB, wDB *sql.DB) http.HandlerFunc {
	return base.HandlePostDataTableView(
		roDB,
		wDB,
		(*workoutdb.Queries).RawInsertProgress,
		func(_ context.Context, values url.Values) (*workoutdb.RawInsertProgressParams, error) {
			weight, err := strconv.ParseFloat(values.Get("weight"), 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse weight: %w", err)
			}
			sets, err := strconv.ParseInt(values.Get("sets"), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse sets: %w", err)
			}
			reps, err := strconv.ParseInt(values.Get("reps"), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse reps: %w", err)
			}
			return &workoutdb.RawInsertProgressParams{
				Date:       values.Get("date"),
				Lift:       values.Get("lift"),
				Weight:     weight,
				Sets:       sets,
				Reps:       reps,
				SideWeight: util.DeZero(values.Get("side_weight")),
			}, nil
		},
		func(ctx context.Context, q *workoutdb.Queries, item workoutdb.Progress) (*templates.DataTableRow, error) {
			lifts, err := q.ListAllIndividualLifts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list lifts: %w", err)
			}
			sws, err := q.ListAllIndividualSideWeights(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list side weights: %w", err)
			}

			sw, _ := item.SideWeight.(string)
			return &templates.DataTableRow{
				PatchEndpoint:  util.UrlPathJoin("/view/data/progress", fmt.Sprint(item.ID)),
				DeleteEndpoint: util.UrlPathJoin("/view/data/progress", fmt.Sprint(item.ID)),
				Values: []templates.DataTableValue{
					{Name: "id", Type: templates.Static, Value: fmt.Sprint(item.ID)},
					{Name: "lift", Type: templates.Select, Value: item.Lift, SelectOptions: lifts},
					{Name: "date", Type: templates.InputString, Value: item.Date},
					{Name: "weight", Type: templates.InputNumber, Value: fmt.Sprint(item.Weight)},
					{Name: "sets", Type: templates.InputNumber, Value: fmt.Sprint(item.Sets)},
					{Name: "reps", Type: templates.InputNumber, Value: fmt.Sprint(item.Reps)},
					{Name: "side_weight", Type: templates.Select, Value: sw, SelectOptions: append(sws, "")},
				},
			}, nil
		},
	)
}
