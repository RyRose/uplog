package rawdata

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"

	"github.com/RyRose/uplog/internal/service/rawdata/util"
	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
	"github.com/RyRose/uplog/internal/templates"
)

func HandleGetLiftView(roDB *sql.DB) http.HandlerFunc {
	return handleGetDataTableView(
		roDB,
		tableViewMetadata{
			headers: []string{"ID", "Link", "Side", "Notes", "Group"},
			post:    "/view/data/lift",
		},
		(*workoutdb.Queries).RawSelectLiftPage,
		func(limit, offset int64) workoutdb.RawSelectLiftPageParams {
			return workoutdb.RawSelectLiftPageParams{
				Limit:  limit,
				Offset: offset,
			}
		},
		func(ctx context.Context, roDB *sql.DB, lifts []workoutdb.Lift) ([]templates.DataTableRow, error) {
			q := workoutdb.New(roDB)
			sideWeightOpts, err := q.ListAllIndividualSideWeights(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list side weights: %w", err)
			}

			liftGroups, err := q.ListAllIndividualLiftGroups(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list side weights: %w", err)
			}
			var rows []templates.DataTableRow
			for _, lift := range lifts {
				rows = append(rows, templates.DataTableRow{
					PatchEndpoint:  util.UrlPathJoin("/view/data/lift", lift.ID),
					DeleteEndpoint: util.UrlPathJoin("/view/data/lift", lift.ID),
					Values: []templates.DataTableValue{
						{Name: "id", Value: lift.ID, Type: templates.InputString},
						{Name: "link", Value: lift.Link, Type: templates.InputString},
						{Name: "default_side_weight",
							Value: util.Zero(lift.DefaultSideWeight),
							Type:  templates.Select, SelectOptions: append(sideWeightOpts, "")},
						{Name: "notes", Value: util.Zero(lift.Notes), Type: templates.InputString},
						{Name: "lift_group",
							Value: util.Zero(lift.LiftGroup),
							Type:  templates.Select, SelectOptions: append(liftGroups, "")},
					},
				})
			}
			rows = append(rows, templates.DataTableRow{
				Values: []templates.DataTableValue{
					{Name: "id", Type: templates.InputString},
					{Name: "link", Type: templates.InputString},
					{Name: "default_side_weight",
						Type: templates.Select, SelectOptions: append(sideWeightOpts, "")},
					{Name: "notes", Type: templates.InputString},
					{Name: "lift_group",
						Type: templates.Select, SelectOptions: append(liftGroups, "")},
				},
			})
			return rows, nil
		},
	)
}

func HandlePatchLiftView(wDB *sql.DB) http.HandlerFunc {
	return handlePatchTableRowViewID(
		wDB,
		map[string]patcherID{
			"id": &patchIDParams[workoutdb.RawUpdateLiftIdParams]{
				query: (*workoutdb.Queries).RawUpdateLiftId,
				convert: func(id, value string) (*workoutdb.RawUpdateLiftIdParams, error) {
					return &workoutdb.RawUpdateLiftIdParams{
						In:  id,
						Out: value,
					}, nil
				},
			},
			"link": &patchIDParams[workoutdb.RawUpdateLiftLinkParams]{
				query: (*workoutdb.Queries).RawUpdateLiftLink,
				convert: func(id, value string) (*workoutdb.RawUpdateLiftLinkParams, error) {
					return &workoutdb.RawUpdateLiftLinkParams{
						ID:   id,
						Link: value,
					}, nil
				},
			},
			"default_side_weight": &patchIDParams[workoutdb.RawUpdateLiftDefaultSideWeightParams]{
				query: (*workoutdb.Queries).RawUpdateLiftDefaultSideWeight,
				convert: func(id, value string) (*workoutdb.RawUpdateLiftDefaultSideWeightParams, error) {
					return &workoutdb.RawUpdateLiftDefaultSideWeightParams{
						ID:                id,
						DefaultSideWeight: util.DeZero(value),
					}, nil
				},
			},
			"notes": &patchIDParams[workoutdb.RawUpdateLiftNotesParams]{
				query: (*workoutdb.Queries).RawUpdateLiftNotes,
				convert: func(id, value string) (*workoutdb.RawUpdateLiftNotesParams, error) {
					return &workoutdb.RawUpdateLiftNotesParams{
						ID:    id,
						Notes: util.DeZero(value),
					}, nil
				},
			},
			"lift_group": &patchIDParams[workoutdb.RawUpdateLiftLiftGroupParams]{
				query: (*workoutdb.Queries).RawUpdateLiftLiftGroup,
				convert: func(id, value string) (*workoutdb.RawUpdateLiftLiftGroupParams, error) {
					return &workoutdb.RawUpdateLiftLiftGroupParams{
						ID:        id,
						LiftGroup: util.DeZero(value),
					}, nil
				},
			},
		},
	)
}

func HandlePostLiftView(roDB, wDB *sql.DB) http.HandlerFunc {
	return handlePostDataTableView(
		roDB,
		wDB,
		(*workoutdb.Queries).RawInsertLift,
		func(_ context.Context, values url.Values) (*workoutdb.RawInsertLiftParams, error) {
			return &workoutdb.RawInsertLiftParams{
				ID:                values.Get("id"),
				Link:              values.Get("link"),
				DefaultSideWeight: util.DeZero(values.Get("default_side_weight")),
				Notes:             util.DeZero(values.Get("notes")),
				LiftGroup:         util.DeZero(values.Get("lift_group")),
			}, nil
		},
		func(ctx context.Context, q *workoutdb.Queries, lift workoutdb.Lift) (*templates.DataTableRow, error) {
			sideWeightOpts, err := q.ListAllIndividualSideWeights(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list side weights: %w", err)
			}
			liftGroups, err := q.ListAllIndividualLiftGroups(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list side weights: %w", err)
			}
			return &templates.DataTableRow{
				PatchEndpoint:  util.UrlPathJoin("/view/data/lift", lift.ID),
				DeleteEndpoint: util.UrlPathJoin("/view/data/lift", lift.ID),
				Values: []templates.DataTableValue{
					{Name: "id", Value: lift.ID, Type: templates.InputString},
					{Name: "link", Value: lift.Link, Type: templates.InputString},
					{Name: "default_side_weight",
						Value: util.Zero(lift.DefaultSideWeight),
						Type:  templates.Select, SelectOptions: append(sideWeightOpts, "")},
					{Name: "notes", Value: util.Zero(lift.Notes), Type: templates.InputString},
					{Name: "lift_group",
						Value: util.Zero(lift.LiftGroup),
						Type:  templates.Select, SelectOptions: append(liftGroups, "")},
				},
			}, nil
		},
	)
}
