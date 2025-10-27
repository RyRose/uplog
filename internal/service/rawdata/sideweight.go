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

// HandleGetSideWeightView godoc
//
//	@Summary		Get side weight data table view
//	@Description	Renders a paginated table view of side weights
//	@Tags			rawdata
//	@Produce		html
//	@Param			offset	query		integer	false	"Pagination offset"
//	@Success		200		{string}	string	"HTML content"
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/view/data/side_weight [get]
func HandleGetSideWeightView(roDB *sql.DB) http.HandlerFunc {
	return base.HandleGetDataTableView(
		roDB,
		base.TableViewMetadata{
			Headers: []string{"ID", "Mult", "Addend", "Format"},
			Post:    "/view/data/side_weight",
		},
		(*workoutdb.Queries).RawSelectSideWeightPage,
		func(limit, offset int64) workoutdb.RawSelectSideWeightPageParams {
			return workoutdb.RawSelectSideWeightPageParams{
				Limit:  limit,
				Offset: offset,
			}
		},
		func(_ context.Context, _ *sql.DB, items []workoutdb.SideWeight) ([]templates.DataTableRow, error) {
			var rows []templates.DataTableRow
			for _, item := range items {
				rows = append(rows, templates.DataTableRow{
					PatchEndpoint:  util.UrlPathJoin("/view/data/side_weight", item.ID),
					DeleteEndpoint: util.UrlPathJoin("/view/data/side_weight", item.ID),
					Values: []templates.DataTableValue{
						{Name: "id", Type: templates.InputString, Value: item.ID},
						{Name: "multiplier", Type: templates.InputNumber, Value: fmt.Sprint(item.Multiplier)},
						{Name: "addend", Type: templates.InputNumber, Value: fmt.Sprint(item.Addend)},
						{Name: "format", Type: templates.InputString, Value: item.Format},
					},
				})
			}
			rows = append(rows, templates.DataTableRow{
				Values: []templates.DataTableValue{
					{Name: "id", Type: templates.InputString},
					{Name: "multiplier", Type: templates.InputNumber},
					{Name: "addend", Type: templates.InputNumber},
					{Name: "format", Type: templates.InputString},
				},
			})
			return rows, nil
		},
	)
}

// HandlePatchSideWeightView godoc
//
//	@Summary		Update side weight data
//	@Description	Updates specific fields of a side weight entry by ID
//	@Tags			rawdata
//	@Accept			x-www-form-urlencoded
//	@Param			id			path		string	true	"Side weight ID"
//	@Param			id			formData	string	false	"New side weight ID"
//	@Param			multiplier	formData	number	false	"Multiplier value"
//	@Param			addend		formData	number	false	"Addend value"
//	@Param			format		formData	string	false	"Format string"
//	@Success		200			{string}	string	"OK"
//	@Failure		400			{string}	string	"Bad request"
//	@Failure		500			{string}	string	"Internal server error"
//	@Router			/view/data/side_weight/{id} [patch]
func HandlePatchSideWeightView(wDB *sql.DB) http.HandlerFunc {
	return base.HandlePatchTableRowViewID(
		wDB,
		map[string]base.PatcherID{
			"id": &base.PatchIDParams[workoutdb.RawUpdateSideWeightIdParams]{
				Query: (*workoutdb.Queries).RawUpdateSideWeightId,
				Convert: func(id, value string) (*workoutdb.RawUpdateSideWeightIdParams, error) {
					return &workoutdb.RawUpdateSideWeightIdParams{
						In:  id,
						Out: value,
					}, nil
				},
			},
			"multiplier": &base.PatchIDParams[workoutdb.RawUpdateSideWeightMultiplierParams]{
				Query: (*workoutdb.Queries).RawUpdateSideWeightMultiplier,
				Convert: func(id, value string) (*workoutdb.RawUpdateSideWeightMultiplierParams, error) {
					v, err := strconv.ParseFloat(value, 64)
					if err != nil {
						return nil, fmt.Errorf("failed to parse multiplier: %w", err)
					}
					return &workoutdb.RawUpdateSideWeightMultiplierParams{
						ID:         id,
						Multiplier: v,
					}, nil
				},
			},
			"addend": &base.PatchIDParams[workoutdb.RawUpdateSideWeightAddendParams]{
				Query: (*workoutdb.Queries).RawUpdateSideWeightAddend,
				Convert: func(id, value string) (*workoutdb.RawUpdateSideWeightAddendParams, error) {
					v, err := strconv.ParseFloat(value, 64)
					if err != nil {
						return nil, fmt.Errorf("failed to parse addend: %w", err)
					}
					return &workoutdb.RawUpdateSideWeightAddendParams{
						ID:     id,
						Addend: v,
					}, nil
				},
			},
			"format": &base.PatchIDParams[workoutdb.RawUpdateSideWeightFormatParams]{
				Query: (*workoutdb.Queries).RawUpdateSideWeightFormat,
				Convert: func(id, value string) (*workoutdb.RawUpdateSideWeightFormatParams, error) {
					return &workoutdb.RawUpdateSideWeightFormatParams{
						ID:     id,
						Format: value,
					}, nil
				},
			},
		},
	)
}

// HandlePostSideWeightView godoc
//
//	@Summary		Create new side weight
//	@Description	Creates a new side weight entry in the database
//	@Tags			rawdata
//	@Accept			x-www-form-urlencoded
//	@Produce		html
//	@Param			id			formData	string	true	"Side weight ID"
//	@Param			multiplier	formData	number	true	"Multiplier value"
//	@Param			addend		formData	number	true	"Addend value"
//	@Param			format		formData	string	true	"Format string"
//	@Success		201			{string}	string	"HTML content"
//	@Failure		400			{string}	string	"Bad request"
//	@Failure		500			{string}	string	"Internal server error"
//	@Router			/view/data/side_weight [post]
func HandlePostSideWeightView(roDB, wDB *sql.DB) http.HandlerFunc {
	return base.HandlePostDataTableView(
		roDB,
		wDB,
		(*workoutdb.Queries).RawInsertSideWeight,
		func(_ context.Context, values url.Values) (*workoutdb.RawInsertSideWeightParams, error) {
			multiplier, err := strconv.ParseFloat(values.Get("multiplier"), 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse multiplier: %w", err)
			}
			addend, err := strconv.ParseFloat(values.Get("addend"), 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse addend: %w", err)
			}
			return &workoutdb.RawInsertSideWeightParams{
				ID:         values.Get("id"),
				Multiplier: multiplier,
				Addend:     addend,
				Format:     values.Get("format"),
			}, nil
		},
		func(_ context.Context, _ *workoutdb.Queries, item workoutdb.SideWeight) (*templates.DataTableRow, error) {
			return &templates.DataTableRow{
				PatchEndpoint:  util.UrlPathJoin("/view/data/side_weight", item.ID),
				DeleteEndpoint: util.UrlPathJoin("/view/data/side_weight", item.ID),
				Values: []templates.DataTableValue{
					{Name: "id", Type: templates.InputString, Value: item.ID},
					{Name: "multiplier", Type: templates.InputNumber, Value: fmt.Sprint(item.Multiplier)},
					{Name: "addend", Type: templates.InputNumber, Value: fmt.Sprint(item.Addend)},
					{Name: "format", Type: templates.InputString, Value: item.Format},
				},
			}, nil
		},
	)
}

// HandleDeleteSideWeightView godoc
//
//	@Summary		Delete side weight
//	@Description	Deletes a side weight entry by ID
//	@Tags			rawdata
//	@Param			id	path		string	true	"Side weight ID"
//	@Success		200	{string}	string	"OK"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/view/data/side_weight/{id} [delete]
func HandleDeleteSideWeightView(wDB *sql.DB) http.HandlerFunc {
	return base.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteSideWeight)
}
