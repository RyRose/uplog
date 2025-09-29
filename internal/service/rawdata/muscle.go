package rawdata

import (
	"context"
	"database/sql"
	"net/http"
	"net/url"

	"github.com/RyRose/uplog/internal/service/rawdata/base"
	"github.com/RyRose/uplog/internal/service/rawdata/util"
	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
	"github.com/RyRose/uplog/internal/templates"
)

func HandleGetMuscleView(roDB *sql.DB) http.HandlerFunc {
	return base.HandleGetDataTableView(
		roDB,
		base.TableViewMetadata{
			Headers: []string{"ID", "Link", "Message"},
			Post:    "/view/data/muscle",
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
					PatchEndpoint:  util.UrlPathJoin("/view/data/muscle", item.ID),
					DeleteEndpoint: util.UrlPathJoin("/view/data/muscle", item.ID),
					Values: []templates.DataTableValue{
						{Name: "id", Value: item.ID, Type: templates.InputString},
						{Name: "link", Value: item.Link, Type: templates.InputString},
						{Name: "message", Value: util.Zero(item.Message), Type: templates.InputString},
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
	return base.HandlePatchTableRowViewID(
		wDB,
		map[string]base.PatcherID{
			"id": &base.PatchIDParams[workoutdb.RawUpdateMuscleIdParams]{
				Query: (*workoutdb.Queries).RawUpdateMuscleId,
				Convert: func(id, value string) (*workoutdb.RawUpdateMuscleIdParams, error) {
					return &workoutdb.RawUpdateMuscleIdParams{
						In:  id,
						Out: value,
					}, nil
				},
			},
			"link": &base.PatchIDParams[workoutdb.RawUpdateMuscleLinkParams]{
				Query: (*workoutdb.Queries).RawUpdateMuscleLink,
				Convert: func(id, value string) (*workoutdb.RawUpdateMuscleLinkParams, error) {
					return &workoutdb.RawUpdateMuscleLinkParams{
						ID:   id,
						Link: value,
					}, nil
				},
			},
			"message": &base.PatchIDParams[workoutdb.RawUpdateMuscleMessageParams]{
				Query: (*workoutdb.Queries).RawUpdateMuscleMessage,
				Convert: func(id, value string) (*workoutdb.RawUpdateMuscleMessageParams, error) {
					return &workoutdb.RawUpdateMuscleMessageParams{
						ID:      id,
						Message: util.DeZero(value),
					}, nil
				},
			},
		},
	)
}

func HandlePostMuscleView(roDB, wDB *sql.DB) http.HandlerFunc {
	return base.HandlePostDataTableView(
		roDB,
		wDB,
		(*workoutdb.Queries).RawInsertMuscle,
		func(_ context.Context, values url.Values) (*workoutdb.RawInsertMuscleParams, error) {
			return &workoutdb.RawInsertMuscleParams{
				ID:      values.Get("id"),
				Link:    values.Get("link"),
				Message: util.DeZero(values.Get("message")),
			}, nil
		},
		func(ctx context.Context, q *workoutdb.Queries, movement workoutdb.Muscle) (*templates.DataTableRow, error) {
			return &templates.DataTableRow{
				PatchEndpoint:  util.UrlPathJoin("/view/data/muscle", movement.ID),
				DeleteEndpoint: util.UrlPathJoin("/view/data/muscle", movement.ID),
				Values: []templates.DataTableValue{
					{Name: "id", Value: movement.ID, Type: templates.InputString},
					{Name: "link", Value: movement.Link, Type: templates.InputString},
					{Name: "message", Value: util.Zero(movement.Message), Type: templates.InputString},
				},
			}, nil
		},
	)
}
