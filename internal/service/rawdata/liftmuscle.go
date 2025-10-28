package rawdata

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"

	"github.com/RyRose/uplog/internal/config"
	"github.com/RyRose/uplog/internal/service/rawdata/base"
	"github.com/RyRose/uplog/internal/service/rawdata/util"
	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
	"github.com/RyRose/uplog/internal/templates"
)

// HandleGetLiftMuscleView godoc
//
//	@Summary		Get lift muscle mapping data table view
//	@Description	Renders a paginated table view of lift-muscle-movement mappings
//	@Tags			rawdata
//	@Produce		html
//	@Param			offset	query		integer	false	"Pagination offset"
//	@Success		200		{string}	string	"HTML content"
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/view/data/lift_muscle_mapping [get]
func HandleGetLiftMuscleView(_ *config.Data, state *config.State) http.HandlerFunc {
	return base.HandleGetDataTableView(
		state.RDB,
		base.TableViewMetadata{
			Headers: []string{"Lift", "Muscle", "Movement"},
			Post:    "/view/data/lift_muscle_mapping",
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
					PatchEndpoint:  util.UrlPathJoin("/view/data/lift_muscle_mapping", item.Lift, item.Muscle, item.Movement),
					DeleteEndpoint: util.UrlPathJoin("/view/data/lift_muscle_mapping", item.Lift, item.Muscle, item.Movement),
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

// HandlePatchLiftMuscleView godoc
//
//	@Summary		Update lift muscle mapping data
//	@Description	Updates specific fields of a lift-muscle-movement mapping
//	@Tags			rawdata
//	@Accept			x-www-form-urlencoded
//	@Param			lift		path		string	true	"Lift ID"
//	@Param			muscle		path		string	true	"Muscle ID"
//	@Param			movement	path		string	true	"Movement ID"
//	@Param			lift		formData	string	false	"New lift ID"
//	@Param			muscle		formData	string	false	"New muscle ID"
//	@Param			movement	formData	string	false	"New movement ID"
//	@Success		200			{string}	string	"OK"
//	@Failure		400			{string}	string	"Bad request"
//	@Failure		500			{string}	string	"Internal server error"
//	@Router			/view/data/lift_muscle_mapping/{lift}/{muscle}/{movement} [patch]
func HandlePatchLiftMuscleView(_ *config.Data, state *config.State) http.HandlerFunc {
	return base.HandlePatchTableRowViewRequest(
		state.WDB,
		map[string]base.PatcherReq{
			"lift": &base.PatchReqParams[workoutdb.RawUpdateLiftMuscleMappingLiftParams]{
				Query: (*workoutdb.Queries).RawUpdateLiftMuscleMappingLift,
				Convert: func(r *http.Request, value string) (*workoutdb.RawUpdateLiftMuscleMappingLiftParams, error) {
					return &workoutdb.RawUpdateLiftMuscleMappingLiftParams{
						Out:      value,
						In:       r.PathValue("lift"),
						Muscle:   r.PathValue("muscle"),
						Movement: r.PathValue("movement"),
					}, nil
				},
			},
			"muscle": &base.PatchReqParams[workoutdb.RawUpdateLiftMuscleMappingMuscleParams]{
				Query: (*workoutdb.Queries).RawUpdateLiftMuscleMappingMuscle,
				Convert: func(r *http.Request, value string) (*workoutdb.RawUpdateLiftMuscleMappingMuscleParams, error) {
					return &workoutdb.RawUpdateLiftMuscleMappingMuscleParams{
						Out:      value,
						In:       r.PathValue("muscle"),
						Lift:     r.PathValue("lift"),
						Movement: r.PathValue("movement"),
					}, nil
				},
			},
			"movement": &base.PatchReqParams[workoutdb.RawUpdateLiftMuscleMappingMovementParams]{
				Query: (*workoutdb.Queries).RawUpdateLiftMuscleMappingMovement,
				Convert: func(r *http.Request, value string) (*workoutdb.RawUpdateLiftMuscleMappingMovementParams, error) {
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

// HandlePostLiftMuscleView godoc
//
//	@Summary		Create new lift muscle mapping
//	@Description	Creates a new lift-muscle-movement mapping in the database
//	@Tags			rawdata
//	@Accept			x-www-form-urlencoded
//	@Produce		html
//	@Param			lift		formData	string	true	"Lift ID"
//	@Param			muscle		formData	string	true	"Muscle ID"
//	@Param			movement	formData	string	true	"Movement ID"
//	@Success		201			{string}	string	"HTML content"
//	@Failure		400			{string}	string	"Bad request"
//	@Failure		500			{string}	string	"Internal server error"
//	@Router			/view/data/lift_muscle_mapping [post]
func HandlePostLiftMuscleView(_ *config.Data, state *config.State) http.HandlerFunc {
	return base.HandlePostDataTableView(
		state.RDB,
		state.WDB,
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
				PatchEndpoint:  util.UrlPathJoin("/view/data/lift_muscle_mapping", item.Lift, item.Muscle, item.Movement),
				DeleteEndpoint: util.UrlPathJoin("/view/data/lift_muscle_mapping", item.Lift, item.Muscle, item.Movement),
				Values: []templates.DataTableValue{
					{Name: "lift", Type: templates.Select, Value: item.Lift, SelectOptions: lifts},
					{Name: "muscle", Type: templates.Select, Value: item.Muscle, SelectOptions: muscles},
					{Name: "movement", Type: templates.Select, Value: item.Movement, SelectOptions: movements},
				},
			}, nil
		},
	)
}

// HandleDeleteLiftMuscleView godoc
//
//	@Summary		Delete lift muscle mapping
//	@Description	Deletes a lift-muscle-movement mapping
//	@Tags			rawdata
//	@Param			lift		path		string	true	"Lift ID"
//	@Param			muscle		path		string	true	"Muscle ID"
//	@Param			movement	path		string	true	"Movement ID"
//	@Success		200			{string}	string	"OK"
//	@Failure		400			{string}	string	"Bad request"
//	@Failure		500			{string}	string	"Internal server error"
//	@Router			/view/data/lift_muscle_mapping/{lift}/{muscle}/{movement} [delete]
func HandleDeleteLiftMuscleView(_ *config.Data, state *config.State) http.HandlerFunc {
	return base.HandleDeleteTableRowViewRequest(
		state.WDB,
		(*workoutdb.Queries).RawDeleteLiftMuscle,
		func(r *http.Request) (*workoutdb.RawDeleteLiftMuscleParams, error) {
			return &workoutdb.RawDeleteLiftMuscleParams{
				Lift:     r.PathValue("lift"),
				Muscle:   r.PathValue("muscle"),
				Movement: r.PathValue("movement"),
			}, nil
		},
	)
}
