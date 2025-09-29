package rawdata

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/RyRose/uplog/internal/service/rawdata/util"
	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
	"github.com/RyRose/uplog/internal/templates"
)

func HandleGetSideWeightView(roDB *sql.DB) http.HandlerFunc {
	return handleGetDataTableView(
		roDB,
		tableViewMetadata{
			headers: []string{"ID", "Mult", "Addend", "Format"},
			post:    "/view/data/side_weight",
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

func HandlePatchSideWeightView(wDB *sql.DB) http.HandlerFunc {
	return handlePatchTableRowViewID(
		wDB,
		map[string]patcherID{
			"id": &patchIDParams[workoutdb.RawUpdateSideWeightIdParams]{
				query: (*workoutdb.Queries).RawUpdateSideWeightId,
				convert: func(id, value string) (*workoutdb.RawUpdateSideWeightIdParams, error) {
					return &workoutdb.RawUpdateSideWeightIdParams{
						In:  id,
						Out: value,
					}, nil
				},
			},
			"multiplier": &patchIDParams[workoutdb.RawUpdateSideWeightMultiplierParams]{
				query: (*workoutdb.Queries).RawUpdateSideWeightMultiplier,
				convert: func(id, value string) (*workoutdb.RawUpdateSideWeightMultiplierParams, error) {
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
			"addend": &patchIDParams[workoutdb.RawUpdateSideWeightAddendParams]{
				query: (*workoutdb.Queries).RawUpdateSideWeightAddend,
				convert: func(id, value string) (*workoutdb.RawUpdateSideWeightAddendParams, error) {
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
			"format": &patchIDParams[workoutdb.RawUpdateSideWeightFormatParams]{
				query: (*workoutdb.Queries).RawUpdateSideWeightFormat,
				convert: func(id, value string) (*workoutdb.RawUpdateSideWeightFormatParams, error) {
					return &workoutdb.RawUpdateSideWeightFormatParams{
						ID:     id,
						Format: value,
					}, nil
				},
			},
		},
	)
}

func HandlePostSideWeightView(roDB, wDB *sql.DB) http.HandlerFunc {
	return handlePostDataTableView(
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
