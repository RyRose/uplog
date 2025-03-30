package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/RyRose/uplog/internal/calendar"
	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
	"github.com/RyRose/uplog/internal/templates"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/exp/constraints"
)

// TODO: Parameterize today's date for testing support.
func todaysDate() time.Time {
	return time.Now()
}

func zero[T any](v *T) T {
	if v == nil {
		var zero T
		return zero
	}
	return *v
}

func deZero[T comparable](v T) *T {
	var zero T
	if v == zero {
		return nil
	}
	return &v
}

func urlPathJoin(base string, parts ...string) string {
	p := []string{base}
	for _, part := range parts {
		p = append(p, url.PathEscape(part))
	}
	return path.Join(p...)
}

func handlePostDataTableView[dataType, paramType any](
	roDB *sql.DB,
	wDB *sql.DB,
	insertQ func(*workoutdb.Queries, context.Context, paramType) (dataType, error),
	convertParams func(context.Context, url.Values) (*paramType, error),
	toRow func(context.Context, *workoutdb.Queries, dataType) (*templates.DataTableRow, error),
) http.HandlerFunc {
	roQ := workoutdb.New(roDB)
	wQ := workoutdb.New(wDB)
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if err := r.ParseForm(); err != nil {
			http.Error(w, fmt.Sprintf("failed to parse form: %v", err), http.StatusBadRequest)
			slog.ErrorContext(ctx, "failed to parse form", "error", err)
			return
		}
		params, err := convertParams(ctx, r.Form)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to convert data: %v", err), http.StatusBadRequest)
			slog.ErrorContext(ctx, "failed to convert data", "error", err)
			return
		}
		data, err := insertQ(wQ, ctx, *params)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to insert data: %v", err), http.StatusInternalServerError)
			slog.ErrorContext(ctx, "failed to insert data", "error", err)
			return
		}
		row, err := toRow(ctx, roQ, data)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to convert data to row: %v", err),
				http.StatusInternalServerError)
			slog.ErrorContext(ctx, "failed to convert data to row", "error", err)
			return
		}
		if err := templates.DataTableRowView(*row).Render(ctx, w); err != nil {
			http.Error(w, fmt.Sprintf("failed to render row: %v", err), http.StatusInternalServerError)
			slog.ErrorContext(ctx, "failed to render row", "error", err)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

type tableViewMetadata struct {
	headers []string
	post    string
}

func minimum[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func maximum[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func handleGetDataTableView[dataType any, limitParams any](
	roDB *sql.DB,
	metadata tableViewMetadata,
	selectQ func(*workoutdb.Queries, context.Context, limitParams) ([]dataType, error),
	convertQ func(int64, int64) limitParams,
	convert func(context.Context, *sql.DB, []dataType) ([]templates.DataTableRow, error),
) http.HandlerFunc {
	queries := workoutdb.New(roDB)
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var limit, offset int64
		limit, offset = 50, 0
		u, err := url.Parse(r.Header.Get("HX-Current-URL"))
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to parse current URL: %v", err), http.StatusBadRequest)
			slog.ErrorContext(ctx, "failed to parse current URL", "error", err, "url", r.Header.Get("HX-Current-URL"))
			return
		}
		rawOffset := u.Query().Get("offset")
		if rawOffset != "" {
			var err error
			offset, err = strconv.ParseInt(rawOffset, 10, 64)
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to parse offset: %v", err), http.StatusBadRequest)
				slog.ErrorContext(ctx, "failed to parse offset", "error", err)
				return
			}
		}
		rawValues, err := selectQ(queries, ctx, convertQ(limit, offset))
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to select data: %v", err), http.StatusInternalServerError)
			slog.ErrorContext(ctx, "failed to select data", "error", err)
			return
		}
		rows, err := convert(ctx, roDB, rawValues)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to convert data: %v", err), http.StatusInternalServerError)
			slog.ErrorContext(ctx, "failed to convert data", "error", err)
			return
		}
		var viewLimit int64
		viewLimit = minimum(limit, int64(len(rows)-1))
		tbl := templates.DataTable{
			Header: templates.DataTableHeader{
				Values: metadata.headers,
			},
			Rows: rows,
			Footer: templates.DataTableFooter{
				PostEndpoint: metadata.post,
				FormID:       "datatableform",
			},
			Start:       fmt.Sprint(maximum(offset, 0) + 1),
			StartOffset: fmt.Sprint(maximum(offset-limit, 0)),
			End:         fmt.Sprint(offset + viewLimit),
			LastPage:    viewLimit != limit,
		}
		templates.DataTableView(tbl).Render(ctx, w)
	}
}

func handleDeleteTableRowViewID(
	wDB *sql.DB,
	deleteQ func(*workoutdb.Queries, context.Context, string) error,
) http.HandlerFunc {
	queries := workoutdb.New(wDB)
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := r.PathValue("id")
		if err := deleteQ(queries, ctx, id); err != nil {
			http.Error(w, fmt.Sprintf("failed to delete row: %v", err), http.StatusInternalServerError)
			slog.ErrorContext(ctx, "failed to delete row", "error", err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
func handleDeleteTableRowViewRequest[idType any](
	wDB *sql.DB,
	deleteQ func(*workoutdb.Queries, context.Context, idType) error,
	convert func(*http.Request) (*idType, error),
) http.HandlerFunc {
	queries := workoutdb.New(wDB)
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id, err := convert(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to convert id: %v", err), http.StatusBadRequest)
			slog.ErrorContext(ctx, "failed to convert id", "error", err)
			return
		}
		if err := deleteQ(queries, ctx, *id); err != nil {
			http.Error(w, fmt.Sprintf("failed to delete row: %v", err), http.StatusInternalServerError)
			slog.ErrorContext(ctx, "failed to delete row", "error", err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

type patchIDParams[dataType any] struct {
	query   func(*workoutdb.Queries, context.Context, dataType) error
	convert func(string, string) (*dataType, error)
}

func (p *patchIDParams[dataType]) Patch(
	ctx context.Context, queries *workoutdb.Queries, key, value string) error {
	data, err := p.convert(key, value)
	if err != nil {
		return fmt.Errorf("failed to convert patch data: %w", err)
	}
	return p.query(queries, ctx, *data)
}

type patcherID interface {
	Patch(ctx context.Context, queries *workoutdb.Queries, key, value string) error
}

type patchReqParams[dataType any] struct {
	query   func(*workoutdb.Queries, context.Context, dataType) error
	convert func(*http.Request, string) (*dataType, error)
}

func (p *patchReqParams[dataType]) Patch(
	ctx context.Context, queries *workoutdb.Queries, r *http.Request, value string) error {
	data, err := p.convert(r, value)
	if err != nil {
		return fmt.Errorf("failed to convert patch data: %w", err)
	}
	return p.query(queries, ctx, *data)
}

type patcherReq interface {
	Patch(context.Context, *workoutdb.Queries, *http.Request, string) error
}

func handlePatchTableRowViewID(
	wDB *sql.DB, patchQ map[string]patcherID) http.HandlerFunc {
	queries := workoutdb.New(wDB)
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := r.PathValue("id")
		if err := r.ParseForm(); err != nil {
			http.Error(w, fmt.Sprintf("failed to parse form: %v", err), http.StatusBadRequest)
			slog.ErrorContext(ctx, "failed to parse form", "error", err)
			return
		}
		for param, values := range r.Form {
			patcher, ok := patchQ[param]
			if !ok {
				slog.WarnContext(ctx, "unknown patch parameter", "param", param)
				continue
			}
			if len(values) != 1 {
				slog.WarnContext(ctx, "unexpected number of values", "param", param, "values", values)
				continue
			}
			if err := patcher.Patch(ctx, queries, id, values[0]); err != nil {
				http.Error(w, fmt.Sprintf("failed to patch row: %v", err), http.StatusInternalServerError)
				slog.ErrorContext(ctx, "failed to patch row", "error", err)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, "no patch data provided", http.StatusBadRequest)
		slog.ErrorContext(ctx, "no patch data provided", "id", id)
	}
}

func handlePatchTableRowViewRequest(
	wDB *sql.DB, patchQ map[string]patcherReq) http.HandlerFunc {
	queries := workoutdb.New(wDB)
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if err := r.ParseForm(); err != nil {
			http.Error(w, fmt.Sprintf("failed to parse form: %v", err), http.StatusBadRequest)
			slog.ErrorContext(ctx, "failed to parse form", "error", err)
			return
		}
		for param, values := range r.Form {
			patcher, ok := patchQ[param]
			if !ok {
				slog.WarnContext(ctx, "unknown patch parameter", "param", param)
				continue
			}
			if len(values) != 1 {
				slog.WarnContext(ctx, "unexpected number of values", "param", param, "values", values)
				continue
			}
			if err := patcher.Patch(ctx, queries, r, values[0]); err != nil {
				http.Error(w, fmt.Sprintf("failed to patch row: %v", err), http.StatusInternalServerError)
				slog.ErrorContext(ctx, "failed to patch row", "error", err)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, "no patch data provided", http.StatusBadRequest)
	}
}

func handleGetDataTabView() http.HandlerFunc {
	tabs := [][]templates.DataTab{
		{
			{Title: "Lifts", Endpoint: "/view/data/lift"},
			{Title: "Movements", Endpoint: "/view/data/movement"},
			{Title: "Muscles", Endpoint: "/view/data/muscle"},
			{Title: "Routines", Endpoint: "/view/data/routine"},
			{Title: "Workouts", Endpoint: "/view/data/workout"},
		},
		{
			{Title: "Schedules", Endpoint: "/view/data/schedule_list"},
			{Title: "Variables", Endpoint: "/view/data/template_variable"},
			{Title: "Schedule", Endpoint: "/view/data/schedule"},
			{Title: "Lift Groups", Endpoint: "/view/data/lift_group"},
		},
		{
			{Title: "Side Weight", Endpoint: "/view/data/side_weight"},
			{Title: "Progress", Endpoint: "/view/data/progress"},
			{Title: "Subworkouts", Endpoint: "/view/data/subworkout"},
		},
		{
			{Title: "Routine:Workout", Endpoint: "/view/data/routine_workout_mapping"},
			{Title: "Lift:Muscle", Endpoint: "/view/data/lift_muscle_mapping"},
			{Title: "Lift:Workout", Endpoint: "/view/data/lift_workout_mapping"},
		},
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		rawTabX, rawTabY := r.PathValue("tabX"), r.PathValue("tabY")
		if rawTabX == "" && rawTabY == "" {
			templates.DataTabView(tabs, 0, 0).Render(ctx, w)
			return
		}
		tabX, err := strconv.Atoi(rawTabX)
		if err != nil || tabX < 0 || tabX >= len(tabs) {
			http.Error(w, "invalid tab index", http.StatusBadRequest)
			slog.ErrorContext(ctx, "invalid tab index", "tabX", r.PathValue("tabX"), "error", err)
			return
		}
		tabY, err := strconv.Atoi(rawTabY)
		if err != nil || tabY < 0 || tabY >= len(tabs[tabX]) {
			http.Error(w, "invalid tab index", http.StatusBadRequest)
			slog.ErrorContext(ctx, "invalid tab index", "tabY", r.PathValue("tabY"), "error", err)
			return
		}
		templates.DataTabView(tabs, tabX, tabY).Render(ctx, w)
	}
}

func handleGetLiftView(roDB *sql.DB) http.HandlerFunc {
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
					PatchEndpoint:  urlPathJoin("/view/data/lift", lift.ID),
					DeleteEndpoint: urlPathJoin("/view/data/lift", lift.ID),
					Values: []templates.DataTableValue{
						{Name: "id", Value: lift.ID, Type: templates.InputString},
						{Name: "link", Value: lift.Link, Type: templates.InputString},
						{Name: "default_side_weight",
							Value: zero(lift.DefaultSideWeight),
							Type:  templates.Select, SelectOptions: append(sideWeightOpts, "")},
						{Name: "notes", Value: zero(lift.Notes), Type: templates.InputString},
						{Name: "lift_group",
							Value: zero(lift.LiftGroup),
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

func handlePatchLiftView(wDB *sql.DB) http.HandlerFunc {
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
						DefaultSideWeight: deZero(value),
					}, nil
				},
			},
			"notes": &patchIDParams[workoutdb.RawUpdateLiftNotesParams]{
				query: (*workoutdb.Queries).RawUpdateLiftNotes,
				convert: func(id, value string) (*workoutdb.RawUpdateLiftNotesParams, error) {
					return &workoutdb.RawUpdateLiftNotesParams{
						ID:    id,
						Notes: deZero(value),
					}, nil
				},
			},
			"lift_group": &patchIDParams[workoutdb.RawUpdateLiftLiftGroupParams]{
				query: (*workoutdb.Queries).RawUpdateLiftLiftGroup,
				convert: func(id, value string) (*workoutdb.RawUpdateLiftLiftGroupParams, error) {
					return &workoutdb.RawUpdateLiftLiftGroupParams{
						ID:        id,
						LiftGroup: deZero(value),
					}, nil
				},
			},
		},
	)
}

func handlePostLiftView(roDB, wDB *sql.DB) http.HandlerFunc {
	return handlePostDataTableView(
		roDB,
		wDB,
		(*workoutdb.Queries).RawInsertLift,
		func(_ context.Context, values url.Values) (*workoutdb.RawInsertLiftParams, error) {
			return &workoutdb.RawInsertLiftParams{
				ID:                values.Get("id"),
				Link:              values.Get("link"),
				DefaultSideWeight: deZero(values.Get("default_side_weight")),
				Notes:             deZero(values.Get("notes")),
				LiftGroup:         deZero(values.Get("lift_group")),
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
				PatchEndpoint:  urlPathJoin("/view/data/lift", lift.ID),
				DeleteEndpoint: urlPathJoin("/view/data/lift", lift.ID),
				Values: []templates.DataTableValue{
					{Name: "id", Value: lift.ID, Type: templates.InputString},
					{Name: "link", Value: lift.Link, Type: templates.InputString},
					{Name: "default_side_weight",
						Value: zero(lift.DefaultSideWeight),
						Type:  templates.Select, SelectOptions: append(sideWeightOpts, "")},
					{Name: "notes", Value: zero(lift.Notes), Type: templates.InputString},
					{Name: "lift_group",
						Value: zero(lift.LiftGroup),
						Type:  templates.Select, SelectOptions: append(liftGroups, "")},
				},
			}, nil
		},
	)
}

func handleGetMovementView(roDB *sql.DB) http.HandlerFunc {
	return handleGetDataTableView(
		roDB,
		tableViewMetadata{
			headers: []string{"ID", "Alias"},
			post:    "/view/data/movement",
		},
		(*workoutdb.Queries).RawSelectMovementPage,
		func(limit, offset int64) workoutdb.RawSelectMovementPageParams {
			return workoutdb.RawSelectMovementPageParams{
				Limit:  limit,
				Offset: offset,
			}
		},
		func(_ context.Context, _ *sql.DB, movements []workoutdb.Movement) ([]templates.DataTableRow, error) {
			var rows []templates.DataTableRow
			for _, movement := range movements {
				rows = append(rows, templates.DataTableRow{
					PatchEndpoint:  urlPathJoin("/view/data/movement", movement.ID),
					DeleteEndpoint: urlPathJoin("/view/data/movement", movement.ID),
					Values: []templates.DataTableValue{
						{Name: "id", Value: movement.ID, Type: templates.InputString},
						{Name: "alias", Value: movement.Alias, Type: templates.InputString},
					},
				})
			}
			rows = append(rows, templates.DataTableRow{
				Values: []templates.DataTableValue{
					{Name: "id", Type: templates.InputString},
					{Name: "alias", Type: templates.InputString},
				},
			})
			return rows, nil
		},
	)
}

func handlePatchMovementView(wDB *sql.DB) http.HandlerFunc {
	return handlePatchTableRowViewID(
		wDB,
		map[string]patcherID{
			"id": &patchIDParams[workoutdb.RawUpdateMovementIdParams]{
				query: (*workoutdb.Queries).RawUpdateMovementId,
				convert: func(id, value string) (*workoutdb.RawUpdateMovementIdParams, error) {
					return &workoutdb.RawUpdateMovementIdParams{
						In:  id,
						Out: value,
					}, nil
				},
			},
			"alias": &patchIDParams[workoutdb.RawUpdateMovementAliasParams]{
				query: (*workoutdb.Queries).RawUpdateMovementAlias,
				convert: func(id, value string) (*workoutdb.RawUpdateMovementAliasParams, error) {
					return &workoutdb.RawUpdateMovementAliasParams{
						ID:    id,
						Alias: value,
					}, nil
				},
			},
		},
	)
}

func handlePostMovementView(roDB, wDB *sql.DB) http.HandlerFunc {
	return handlePostDataTableView(
		roDB,
		wDB,
		(*workoutdb.Queries).RawInsertMovement,
		func(_ context.Context, values url.Values) (*workoutdb.RawInsertMovementParams, error) {
			return &workoutdb.RawInsertMovementParams{
				ID:    values.Get("id"),
				Alias: values.Get("alias"),
			}, nil
		},
		func(ctx context.Context, q *workoutdb.Queries, movement workoutdb.Movement) (*templates.DataTableRow, error) {
			return &templates.DataTableRow{
				PatchEndpoint:  urlPathJoin("/view/data/movement", movement.ID),
				DeleteEndpoint: urlPathJoin("/view/data/movement", movement.ID),
				Values: []templates.DataTableValue{
					{Name: "id", Value: movement.ID, Type: templates.InputString},
					{Name: "alias", Value: movement.Alias, Type: templates.InputString},
				},
			}, nil
		},
	)
}

func handleGetMuscleView(roDB *sql.DB) http.HandlerFunc {
	return handleGetDataTableView(
		roDB,
		tableViewMetadata{
			headers: []string{"ID", "Link", "Message"},
			post:    "/view/data/muscle",
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
					PatchEndpoint:  urlPathJoin("/view/data/muscle", item.ID),
					DeleteEndpoint: urlPathJoin("/view/data/muscle", item.ID),
					Values: []templates.DataTableValue{
						{Name: "id", Value: item.ID, Type: templates.InputString},
						{Name: "link", Value: item.Link, Type: templates.InputString},
						{Name: "message", Value: zero(item.Message), Type: templates.InputString},
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

func handlePatchMuscleView(wDB *sql.DB) http.HandlerFunc {
	return handlePatchTableRowViewID(
		wDB,
		map[string]patcherID{
			"id": &patchIDParams[workoutdb.RawUpdateMuscleIdParams]{
				query: (*workoutdb.Queries).RawUpdateMuscleId,
				convert: func(id, value string) (*workoutdb.RawUpdateMuscleIdParams, error) {
					return &workoutdb.RawUpdateMuscleIdParams{
						In:  id,
						Out: value,
					}, nil
				},
			},
			"link": &patchIDParams[workoutdb.RawUpdateMuscleLinkParams]{
				query: (*workoutdb.Queries).RawUpdateMuscleLink,
				convert: func(id, value string) (*workoutdb.RawUpdateMuscleLinkParams, error) {
					return &workoutdb.RawUpdateMuscleLinkParams{
						ID:   id,
						Link: value,
					}, nil
				},
			},
			"message": &patchIDParams[workoutdb.RawUpdateMuscleMessageParams]{
				query: (*workoutdb.Queries).RawUpdateMuscleMessage,
				convert: func(id, value string) (*workoutdb.RawUpdateMuscleMessageParams, error) {
					return &workoutdb.RawUpdateMuscleMessageParams{
						ID:      id,
						Message: deZero(value),
					}, nil
				},
			},
		},
	)
}

func handlePostMuscleView(roDB, wDB *sql.DB) http.HandlerFunc {
	return handlePostDataTableView(
		roDB,
		wDB,
		(*workoutdb.Queries).RawInsertMuscle,
		func(_ context.Context, values url.Values) (*workoutdb.RawInsertMuscleParams, error) {
			return &workoutdb.RawInsertMuscleParams{
				ID:      values.Get("id"),
				Link:    values.Get("link"),
				Message: deZero(values.Get("message")),
			}, nil
		},
		func(ctx context.Context, q *workoutdb.Queries, movement workoutdb.Muscle) (*templates.DataTableRow, error) {
			return &templates.DataTableRow{
				PatchEndpoint:  urlPathJoin("/view/data/muscle", movement.ID),
				DeleteEndpoint: urlPathJoin("/view/data/muscle", movement.ID),
				Values: []templates.DataTableValue{
					{Name: "id", Value: movement.ID, Type: templates.InputString},
					{Name: "link", Value: movement.Link, Type: templates.InputString},
					{Name: "message", Value: zero(movement.Message), Type: templates.InputString},
				},
			}, nil
		},
	)
}

func handleGetRoutineView(roDB *sql.DB) http.HandlerFunc {
	return handleGetDataTableView(
		roDB,
		tableViewMetadata{
			headers: []string{"ID", "Steps", "Lift"},
			post:    "/view/data/routine",
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
					PatchEndpoint:  urlPathJoin("/view/data/routine", item.ID),
					DeleteEndpoint: urlPathJoin("/view/data/routine", item.ID),
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

func handlePatchRoutineView(wDB *sql.DB) http.HandlerFunc {
	return handlePatchTableRowViewID(
		wDB,
		map[string]patcherID{
			"id": &patchIDParams[workoutdb.RawUpdateRoutineIdParams]{
				query: (*workoutdb.Queries).RawUpdateRoutineId,
				convert: func(id, value string) (*workoutdb.RawUpdateRoutineIdParams, error) {
					return &workoutdb.RawUpdateRoutineIdParams{
						In:  id,
						Out: value,
					}, nil
				},
			},
			"steps": &patchIDParams[workoutdb.RawUpdateRoutineStepsParams]{
				query: (*workoutdb.Queries).RawUpdateRoutineSteps,
				convert: func(id, value string) (*workoutdb.RawUpdateRoutineStepsParams, error) {
					return &workoutdb.RawUpdateRoutineStepsParams{
						ID:    id,
						Steps: value,
					}, nil
				},
			},
			"lift": &patchIDParams[workoutdb.RawUpdateRoutineLiftParams]{
				query: (*workoutdb.Queries).RawUpdateRoutineLift,
				convert: func(id, value string) (*workoutdb.RawUpdateRoutineLiftParams, error) {
					return &workoutdb.RawUpdateRoutineLiftParams{
						ID:   id,
						Lift: value,
					}, nil
				},
			},
		},
	)
}

func handlePostRoutineView(roDB, wDB *sql.DB) http.HandlerFunc {
	return handlePostDataTableView(
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
				PatchEndpoint:  urlPathJoin("/view/data/routine", item.ID),
				DeleteEndpoint: urlPathJoin("/view/data/routine", item.ID),
				Values: []templates.DataTableValue{
					{Name: "id", Value: item.ID, Type: templates.InputString},
					{Name: "steps", Value: item.Steps, Type: templates.InputString},
					{Name: "lift", Value: item.Lift, Type: templates.Select, SelectOptions: lifts},
				},
			}, nil
		},
	)
}

func handleGetScheduleListView(roDB *sql.DB) http.HandlerFunc {
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

func handlePatchScheduleListView(wDB *sql.DB) http.HandlerFunc {
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

func handlePostScheduleListView(roDB, wDB *sql.DB) http.HandlerFunc {
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

func handleGetSideWeightView(roDB *sql.DB) http.HandlerFunc {
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
					PatchEndpoint:  urlPathJoin("/view/data/side_weight", item.ID),
					DeleteEndpoint: urlPathJoin("/view/data/side_weight", item.ID),
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

func handlePatchSideWeightView(wDB *sql.DB) http.HandlerFunc {
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

func handlePostSideWeightView(roDB, wDB *sql.DB) http.HandlerFunc {
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
				PatchEndpoint:  urlPathJoin("/view/data/side_weight", item.ID),
				DeleteEndpoint: urlPathJoin("/view/data/side_weight", item.ID),
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

func handleGetTemplateVariableView(roDB *sql.DB) http.HandlerFunc {
	return handleGetDataTableView(
		roDB,
		tableViewMetadata{
			headers: []string{"ID", "Value"},
			post:    "/view/data/template_variable",
		},
		(*workoutdb.Queries).RawSelectTemplateVariablePage,
		func(limit, offset int64) workoutdb.RawSelectTemplateVariablePageParams {
			return workoutdb.RawSelectTemplateVariablePageParams{
				Limit:  limit,
				Offset: offset,
			}
		},
		func(_ context.Context, _ *sql.DB, items []workoutdb.TemplateVariable) ([]templates.DataTableRow, error) {
			var rows []templates.DataTableRow
			for _, item := range items {
				rows = append(rows, templates.DataTableRow{
					PatchEndpoint:  urlPathJoin("/view/data/template_variable", item.ID),
					DeleteEndpoint: urlPathJoin("/view/data/template_variable", item.ID),
					Values: []templates.DataTableValue{
						{Name: "id", Type: templates.InputString, Value: item.ID},
						{Name: "value", Type: templates.TextArea, Value: item.Value},
					},
				})
			}
			rows = append(rows, templates.DataTableRow{
				Values: []templates.DataTableValue{
					{Name: "id", Type: templates.InputString},
					{Name: "value", Type: templates.TextArea},
				},
			})
			return rows, nil
		},
	)
}

func handlePatchTemplateVariableView(wDB *sql.DB) http.HandlerFunc {
	return handlePatchTableRowViewID(
		wDB,
		map[string]patcherID{
			"id": &patchIDParams[workoutdb.RawUpdateTemplateVariableIdParams]{
				query: (*workoutdb.Queries).RawUpdateTemplateVariableId,
				convert: func(id, value string) (*workoutdb.RawUpdateTemplateVariableIdParams, error) {
					return &workoutdb.RawUpdateTemplateVariableIdParams{
						In:  id,
						Out: value,
					}, nil
				},
			},
			"value": &patchIDParams[workoutdb.RawUpdateTemplateVariableValueParams]{
				query: (*workoutdb.Queries).RawUpdateTemplateVariableValue,
				convert: func(id, value string) (*workoutdb.RawUpdateTemplateVariableValueParams, error) {
					return &workoutdb.RawUpdateTemplateVariableValueParams{
						ID:    id,
						Value: value,
					}, nil
				},
			},
		},
	)
}

func handlePostTemplateVariableView(roDB, wDB *sql.DB) http.HandlerFunc {
	return handlePostDataTableView(
		roDB,
		wDB,
		(*workoutdb.Queries).RawInsertTemplateVariable,
		func(_ context.Context, values url.Values) (*workoutdb.RawInsertTemplateVariableParams, error) {
			return &workoutdb.RawInsertTemplateVariableParams{
				ID:    values.Get("id"),
				Value: values.Get("value"),
			}, nil
		},
		func(_ context.Context, _ *workoutdb.Queries, item workoutdb.TemplateVariable) (*templates.DataTableRow, error) {
			return &templates.DataTableRow{
				PatchEndpoint:  urlPathJoin("/view/data/template_variable", item.ID),
				DeleteEndpoint: urlPathJoin("/view/data/template_variable", item.ID),
				Values: []templates.DataTableValue{
					{Name: "id", Type: templates.InputString, Value: item.ID},
					{Name: "value", Type: templates.TextArea, Value: item.Value},
				},
			}, nil
		},
	)
}

func handleGetWorkoutView(roDB *sql.DB) http.HandlerFunc {
	return handleGetDataTableView(
		roDB,
		tableViewMetadata{
			headers: []string{"ID", "Template"},
			post:    "/view/data/workout",
		},
		(*workoutdb.Queries).RawSelectWorkoutPage,
		func(limit, offset int64) workoutdb.RawSelectWorkoutPageParams {
			return workoutdb.RawSelectWorkoutPageParams{
				Limit:  limit,
				Offset: offset,
			}
		},
		func(_ context.Context, _ *sql.DB, items []workoutdb.Workout) ([]templates.DataTableRow, error) {
			var rows []templates.DataTableRow
			for _, item := range items {
				rows = append(rows, templates.DataTableRow{
					PatchEndpoint:  urlPathJoin("/view/data/workout", item.ID),
					DeleteEndpoint: urlPathJoin("/view/data/workout", item.ID),
					Values: []templates.DataTableValue{
						{Name: "id", Type: templates.InputString, Value: item.ID},
						{Name: "template", Type: templates.TextArea, Value: item.Template},
					},
				})
			}
			rows = append(rows, templates.DataTableRow{
				Values: []templates.DataTableValue{
					{Name: "id", Type: templates.InputString},
					{Name: "template", Type: templates.TextArea},
				},
			})
			return rows, nil
		},
	)
}

func handlePatchWorkoutView(wDB *sql.DB) http.HandlerFunc {
	return handlePatchTableRowViewID(
		wDB,
		map[string]patcherID{
			"id": &patchIDParams[workoutdb.RawUpdateWorkoutIdParams]{
				query: (*workoutdb.Queries).RawUpdateWorkoutId,
				convert: func(id, value string) (*workoutdb.RawUpdateWorkoutIdParams, error) {
					return &workoutdb.RawUpdateWorkoutIdParams{
						In:  id,
						Out: value,
					}, nil
				},
			},
			"template": &patchIDParams[workoutdb.RawUpdateWorkoutTemplateParams]{
				query: (*workoutdb.Queries).RawUpdateWorkoutTemplate,
				convert: func(id, value string) (*workoutdb.RawUpdateWorkoutTemplateParams, error) {
					return &workoutdb.RawUpdateWorkoutTemplateParams{
						ID:       id,
						Template: value,
					}, nil
				},
			},
		},
	)
}

func handlePostWorkoutView(roDB, wDB *sql.DB) http.HandlerFunc {
	return handlePostDataTableView(
		roDB,
		wDB,
		(*workoutdb.Queries).RawInsertWorkout,
		func(_ context.Context, values url.Values) (*workoutdb.RawInsertWorkoutParams, error) {
			return &workoutdb.RawInsertWorkoutParams{
				ID:       values.Get("id"),
				Template: values.Get("template"),
			}, nil
		},
		func(_ context.Context, _ *workoutdb.Queries, item workoutdb.Workout) (*templates.DataTableRow, error) {
			return &templates.DataTableRow{
				PatchEndpoint:  urlPathJoin("/view/data/workout", item.ID),
				DeleteEndpoint: urlPathJoin("/view/data/workout", item.ID),
				Values: []templates.DataTableValue{
					{Name: "id", Type: templates.InputString, Value: item.ID},
					{Name: "template", Type: templates.TextArea, Value: item.Template},
				},
			}, nil
		},
	)
}

func handleGetScheduleView(roDB *sql.DB) http.HandlerFunc {
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

func handlePatchScheduleView(wDB *sql.DB) http.HandlerFunc {
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

func handlePostScheduleView(roDB, wDB *sql.DB) http.HandlerFunc {
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

func handleGetProgressView(roDB *sql.DB) http.HandlerFunc {
	return handleGetDataTableView(
		roDB,
		tableViewMetadata{
			headers: []string{"ID", "Lift", "Date", "Weight", "Sets", "Reps", "SW"},
			post:    "/view/data/progress",
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
					PatchEndpoint:  urlPathJoin("/view/data/progress", fmt.Sprint(item.ID)),
					DeleteEndpoint: urlPathJoin("/view/data/progress", fmt.Sprint(item.ID)),
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

func handlePatchProgressView(wDB *sql.DB) http.HandlerFunc {
	return handlePatchTableRowViewID(
		wDB,
		map[string]patcherID{
			"lift": &patchIDParams[workoutdb.RawUpdateProgressLiftParams]{
				query: (*workoutdb.Queries).RawUpdateProgressLift,
				convert: func(id, value string) (*workoutdb.RawUpdateProgressLiftParams, error) {
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
			"date": &patchIDParams[workoutdb.RawUpdateProgressDateParams]{
				query: (*workoutdb.Queries).RawUpdateProgressDate,
				convert: func(id, value string) (*workoutdb.RawUpdateProgressDateParams, error) {
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
			"weight": &patchIDParams[workoutdb.RawUpdateProgressWeightParams]{
				query: (*workoutdb.Queries).RawUpdateProgressWeight,
				convert: func(id, value string) (*workoutdb.RawUpdateProgressWeightParams, error) {
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
			"sets": &patchIDParams[workoutdb.RawUpdateProgressSetsParams]{
				query: (*workoutdb.Queries).RawUpdateProgressSets,
				convert: func(id, value string) (*workoutdb.RawUpdateProgressSetsParams, error) {
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
			"reps": &patchIDParams[workoutdb.RawUpdateProgressRepsParams]{
				query: (*workoutdb.Queries).RawUpdateProgressReps,
				convert: func(id, value string) (*workoutdb.RawUpdateProgressRepsParams, error) {
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
			"side_weight": &patchIDParams[workoutdb.RawUpdateProgressSideWeightParams]{
				query: (*workoutdb.Queries).RawUpdateProgressSideWeight,
				convert: func(id, value string) (*workoutdb.RawUpdateProgressSideWeightParams, error) {
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

func handlePostProgressView(roDB, wDB *sql.DB) http.HandlerFunc {
	return handlePostDataTableView(
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
				SideWeight: deZero(values.Get("side_weight")),
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
				PatchEndpoint:  urlPathJoin("/view/data/progress", fmt.Sprint(item.ID)),
				DeleteEndpoint: urlPathJoin("/view/data/progress", fmt.Sprint(item.ID)),
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

func handleGetLiftMuscleView(roDB *sql.DB) http.HandlerFunc {
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

func handlePatchLiftMuscleView(wDB *sql.DB) http.HandlerFunc {
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

func handlePostLiftMuscleView(roDB, wDB *sql.DB) http.HandlerFunc {
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

func handleGetLiftWorkoutView(roDB *sql.DB) http.HandlerFunc {
	return handleGetDataTableView(
		roDB,
		tableViewMetadata{
			headers: []string{"Lift", "Workout"},
			post:    "/view/data/lift_workout_mapping",
		},
		(*workoutdb.Queries).RawSelectLiftWorkoutPage,
		func(limit, offset int64) workoutdb.RawSelectLiftWorkoutPageParams {
			return workoutdb.RawSelectLiftWorkoutPageParams{
				Limit:  limit,
				Offset: offset,
			}
		},
		func(ctx context.Context, roDB *sql.DB, items []workoutdb.LiftWorkoutMapping) ([]templates.DataTableRow, error) {
			q := workoutdb.New(roDB)
			lifts, err := q.ListAllIndividualLifts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list lifts: %w", err)
			}
			workouts, err := q.ListAllIndividualWorkouts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list workouts: %w", err)
			}
			var rows []templates.DataTableRow
			for _, item := range items {
				rows = append(rows, templates.DataTableRow{
					PatchEndpoint:  urlPathJoin("/view/data/lift_workout_mapping", item.Lift, item.Workout),
					DeleteEndpoint: urlPathJoin("/view/data/lift_workout_mapping", item.Lift, item.Workout),
					Values: []templates.DataTableValue{
						{Name: "lift", Type: templates.Select, Value: item.Lift, SelectOptions: lifts},
						{Name: "workout", Type: templates.Select, Value: item.Workout, SelectOptions: workouts},
					},
				})
			}
			rows = append(rows, templates.DataTableRow{
				Values: []templates.DataTableValue{
					{Name: "lift", Type: templates.Select, SelectOptions: lifts},
					{Name: "workout", Type: templates.Select, SelectOptions: workouts},
				},
			})
			return rows, nil
		},
	)
}

func handlePatchLiftWorkoutView(wDB *sql.DB) http.HandlerFunc {
	return handlePatchTableRowViewRequest(
		wDB,
		map[string]patcherReq{
			"lift": &patchReqParams[workoutdb.RawUpdateLiftWorkoutMappingLiftParams]{
				query: (*workoutdb.Queries).RawUpdateLiftWorkoutMappingLift,
				convert: func(r *http.Request, value string) (*workoutdb.RawUpdateLiftWorkoutMappingLiftParams, error) {
					return &workoutdb.RawUpdateLiftWorkoutMappingLiftParams{
						Out:     value,
						In:      r.PathValue("lift"),
						Workout: r.PathValue("workout"),
					}, nil
				},
			},
			"workout": &patchReqParams[workoutdb.RawUpdateLiftWorkoutMappingWorkoutParams]{
				query: (*workoutdb.Queries).RawUpdateLiftWorkoutMappingWorkout,
				convert: func(r *http.Request, value string) (*workoutdb.RawUpdateLiftWorkoutMappingWorkoutParams, error) {
					return &workoutdb.RawUpdateLiftWorkoutMappingWorkoutParams{
						Out:  value,
						In:   r.PathValue("workout"),
						Lift: r.PathValue("lift"),
					}, nil
				},
			},
		},
	)
}

func handlePostLiftWorkoutView(roDB, wDB *sql.DB) http.HandlerFunc {
	return handlePostDataTableView(
		roDB,
		wDB,
		(*workoutdb.Queries).RawInsertLiftWorkout,
		func(_ context.Context, values url.Values) (*workoutdb.RawInsertLiftWorkoutParams, error) {
			return &workoutdb.RawInsertLiftWorkoutParams{
				Lift:    values.Get("lift"),
				Workout: values.Get("workout"),
			}, nil
		},
		func(ctx context.Context, q *workoutdb.Queries, item workoutdb.LiftWorkoutMapping) (*templates.DataTableRow, error) {
			lifts, err := q.ListAllIndividualLifts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list lifts: %w", err)
			}
			workouts, err := q.ListAllIndividualWorkouts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list workouts: %w", err)
			}

			return &templates.DataTableRow{
				PatchEndpoint:  urlPathJoin("/view/data/lift_workout_mapping", item.Lift, item.Workout),
				DeleteEndpoint: urlPathJoin("/view/data/lift_workout_mapping", item.Lift, item.Workout),
				Values: []templates.DataTableValue{
					{Name: "lift", Type: templates.Select, Value: item.Lift, SelectOptions: lifts},
					{Name: "workout", Type: templates.Select, Value: item.Workout, SelectOptions: workouts},
				},
			}, nil
		},
	)
}

func handleGetRoutineWorkoutView(roDB *sql.DB) http.HandlerFunc {
	return handleGetDataTableView(
		roDB,
		tableViewMetadata{
			headers: []string{"Routine", "Workout"},
			post:    "/view/data/routine_workout_mapping",
		},
		(*workoutdb.Queries).RawSelectRoutineWorkoutPage,
		func(limit, offset int64) workoutdb.RawSelectRoutineWorkoutPageParams {
			return workoutdb.RawSelectRoutineWorkoutPageParams{
				Limit:  limit,
				Offset: offset,
			}
		},
		func(ctx context.Context, roDB *sql.DB, items []workoutdb.RoutineWorkoutMapping) ([]templates.DataTableRow, error) {
			q := workoutdb.New(roDB)
			routines, err := q.ListAllIndividualRoutines(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list routines: %w", err)
			}
			workouts, err := q.ListAllIndividualWorkouts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list workouts: %w", err)
			}
			var rows []templates.DataTableRow
			for _, item := range items {
				rows = append(rows, templates.DataTableRow{
					PatchEndpoint:  urlPathJoin("/view/data/routine_workout_mapping", item.Routine, item.Workout),
					DeleteEndpoint: urlPathJoin("/view/data/routine_workout_mapping", item.Routine, item.Workout),
					Values: []templates.DataTableValue{
						{Name: "routine", Type: templates.Select, Value: item.Routine, SelectOptions: routines},
						{Name: "workout", Type: templates.Select, Value: item.Workout, SelectOptions: workouts},
					},
				})
			}
			rows = append(rows, templates.DataTableRow{
				Values: []templates.DataTableValue{
					{Name: "routine", Type: templates.Select, SelectOptions: routines},
					{Name: "workout", Type: templates.Select, SelectOptions: workouts},
				},
			})
			return rows, nil
		},
	)
}

func handlePatchRoutineWorkoutView(wDB *sql.DB) http.HandlerFunc {
	return handlePatchTableRowViewRequest(
		wDB,
		map[string]patcherReq{
			"routine": &patchReqParams[workoutdb.RawUpdateRoutineWorkoutMappingRoutineParams]{
				query: (*workoutdb.Queries).RawUpdateRoutineWorkoutMappingRoutine,
				convert: func(r *http.Request, value string) (*workoutdb.RawUpdateRoutineWorkoutMappingRoutineParams, error) {
					return &workoutdb.RawUpdateRoutineWorkoutMappingRoutineParams{
						Out:     value,
						In:      r.PathValue("routine"),
						Workout: r.PathValue("workout"),
					}, nil
				},
			},
			"workout": &patchReqParams[workoutdb.RawUpdateRoutineWorkoutMappingWorkoutParams]{
				query: (*workoutdb.Queries).RawUpdateRoutineWorkoutMappingWorkout,
				convert: func(r *http.Request, value string) (*workoutdb.RawUpdateRoutineWorkoutMappingWorkoutParams, error) {
					return &workoutdb.RawUpdateRoutineWorkoutMappingWorkoutParams{
						Out:     value,
						In:      r.PathValue("workout"),
						Routine: r.PathValue("routine"),
					}, nil
				},
			},
		},
	)
}

func handlePostRoutineWorkoutView(roDB, wDB *sql.DB) http.HandlerFunc {
	return handlePostDataTableView(
		roDB,
		wDB,
		(*workoutdb.Queries).RawInsertRoutineWorkout,
		func(_ context.Context, values url.Values) (*workoutdb.RawInsertRoutineWorkoutParams, error) {
			return &workoutdb.RawInsertRoutineWorkoutParams{
				Routine: values.Get("routine"),
				Workout: values.Get("workout"),
			}, nil
		},
		func(ctx context.Context, q *workoutdb.Queries, item workoutdb.RoutineWorkoutMapping) (*templates.DataTableRow, error) {
			routines, err := q.ListAllIndividualRoutines(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list routines: %w", err)
			}
			workouts, err := q.ListAllIndividualWorkouts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list workouts: %w", err)
			}

			return &templates.DataTableRow{
				PatchEndpoint:  urlPathJoin("/view/data/routine_workout_mapping", item.Routine, item.Workout),
				DeleteEndpoint: urlPathJoin("/view/data/routine_workout_mapping", item.Routine, item.Workout),
				Values: []templates.DataTableValue{
					{Name: "routine", Type: templates.Select, Value: item.Routine, SelectOptions: routines},
					{Name: "workout", Type: templates.Select, Value: item.Workout, SelectOptions: workouts},
				},
			}, nil
		},
	)
}

func handleGetSubworkoutView(roDB *sql.DB) http.HandlerFunc {
	return handleGetDataTableView(
		roDB,
		tableViewMetadata{
			headers: []string{"Subworkout", "Superworkout"},
			post:    "/view/data/subworkout",
		},
		(*workoutdb.Queries).RawSelectSubworkoutPage,
		func(limit, offset int64) workoutdb.RawSelectSubworkoutPageParams {
			return workoutdb.RawSelectSubworkoutPageParams{
				Limit:  limit,
				Offset: offset,
			}
		},
		func(ctx context.Context, roDB *sql.DB, items []workoutdb.Subworkout) ([]templates.DataTableRow, error) {
			q := workoutdb.New(roDB)
			workouts, err := q.ListAllIndividualWorkouts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list workouts: %w", err)
			}
			var rows []templates.DataTableRow
			for _, item := range items {
				rows = append(rows, templates.DataTableRow{
					PatchEndpoint:  urlPathJoin("/view/data/subworkout", item.Subworkout, item.Superworkout),
					DeleteEndpoint: urlPathJoin("/view/data/subworkout", item.Subworkout, item.Superworkout),
					Values: []templates.DataTableValue{
						{Name: "subworkout", Type: templates.Select, Value: item.Subworkout, SelectOptions: workouts},
						{Name: "superworkout", Type: templates.Select, Value: item.Superworkout, SelectOptions: workouts},
					},
				})
			}
			rows = append(rows, templates.DataTableRow{
				Values: []templates.DataTableValue{
					{Name: "subworkout", Type: templates.Select, SelectOptions: workouts},
					{Name: "superworkout", Type: templates.Select, SelectOptions: workouts},
				},
			})
			return rows, nil
		},
	)
}

func handlePatchSubworkoutView(wDB *sql.DB) http.HandlerFunc {
	return handlePatchTableRowViewRequest(
		wDB,
		map[string]patcherReq{
			"subworkout": &patchReqParams[workoutdb.RawUpdateSubworkoutSubworkoutParams]{
				query: (*workoutdb.Queries).RawUpdateSubworkoutSubworkout,
				convert: func(r *http.Request, value string) (*workoutdb.RawUpdateSubworkoutSubworkoutParams, error) {
					return &workoutdb.RawUpdateSubworkoutSubworkoutParams{
						Out:          value,
						In:           r.PathValue("subworkout"),
						Superworkout: r.PathValue("superworkout"),
					}, nil
				},
			},
			"superworkout": &patchReqParams[workoutdb.RawUpdateSubworkoutSuperworkoutParams]{
				query: (*workoutdb.Queries).RawUpdateSubworkoutSuperworkout,
				convert: func(r *http.Request, value string) (*workoutdb.RawUpdateSubworkoutSuperworkoutParams, error) {
					return &workoutdb.RawUpdateSubworkoutSuperworkoutParams{
						Out:        value,
						In:         r.PathValue("superworkout"),
						Subworkout: r.PathValue("subworkout"),
					}, nil
				},
			},
		},
	)
}

func handlePostSubworkoutView(roDB, wDB *sql.DB) http.HandlerFunc {
	return handlePostDataTableView(
		roDB,
		wDB,
		(*workoutdb.Queries).RawInsertSubworkout,
		func(_ context.Context, values url.Values) (*workoutdb.RawInsertSubworkoutParams, error) {
			return &workoutdb.RawInsertSubworkoutParams{
				Subworkout:   values.Get("subworkout"),
				Superworkout: values.Get("superworkout"),
			}, nil
		},
		func(ctx context.Context, q *workoutdb.Queries, item workoutdb.Subworkout) (*templates.DataTableRow, error) {
			workouts, err := q.ListAllIndividualWorkouts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list workouts: %w", err)
			}

			return &templates.DataTableRow{
				PatchEndpoint:  urlPathJoin("/view/data/subworkout", item.Subworkout, item.Superworkout),
				DeleteEndpoint: urlPathJoin("/view/data/subworkout", item.Subworkout, item.Superworkout),
				Values: []templates.DataTableValue{
					{Name: "subworkout", Type: templates.Select, Value: item.Subworkout, SelectOptions: workouts},
					{Name: "superworkout", Type: templates.Select, Value: item.Superworkout, SelectOptions: workouts},
				},
			}, nil
		},
	)
}

func handleGetLiftGroupView(roDB *sql.DB) http.HandlerFunc {
	return handleGetDataTableView(
		roDB,
		tableViewMetadata{
			headers: []string{"ID"},
			post:    "/view/data/lift_group",
		},
		(*workoutdb.Queries).RawSelectLiftGroupPage,
		func(limit, offset int64) workoutdb.RawSelectLiftGroupPageParams {
			return workoutdb.RawSelectLiftGroupPageParams{
				Limit:  limit,
				Offset: offset,
			}
		},
		func(_ context.Context, _ *sql.DB, liftGroups []string) ([]templates.DataTableRow, error) {
			var rows []templates.DataTableRow
			for _, liftGroup := range liftGroups {
				rows = append(rows, templates.DataTableRow{
					PatchEndpoint:  urlPathJoin("/view/data/lift_group", liftGroup),
					DeleteEndpoint: urlPathJoin("/view/data/lift_group", liftGroup),
					Values: []templates.DataTableValue{
						{Name: "id", Value: liftGroup, Type: templates.InputString},
					},
				})
			}
			rows = append(rows, templates.DataTableRow{
				Values: []templates.DataTableValue{
					{Name: "id", Type: templates.InputString},
				},
			})
			return rows, nil
		},
	)
}

func handlePatchLiftGroupView(wDB *sql.DB) http.HandlerFunc {
	return handlePatchTableRowViewID(
		wDB,
		map[string]patcherID{
			"id": &patchIDParams[workoutdb.RawUpdateLiftGroupIdParams]{
				query: (*workoutdb.Queries).RawUpdateLiftGroupId,
				convert: func(id, value string) (*workoutdb.RawUpdateLiftGroupIdParams, error) {
					return &workoutdb.RawUpdateLiftGroupIdParams{
						In:  id,
						Out: value,
					}, nil
				},
			},
		},
	)
}

func handlePostLiftGroupView(roDB, wDB *sql.DB) http.HandlerFunc {
	return handlePostDataTableView(
		roDB,
		wDB,
		(*workoutdb.Queries).RawInsertLiftGroup,
		func(_ context.Context, values url.Values) (*string, error) {
			id := values.Get("id")
			return &id, nil
		},
		func(ctx context.Context, q *workoutdb.Queries, liftGroup string) (*templates.DataTableRow, error) {
			return &templates.DataTableRow{
				PatchEndpoint:  urlPathJoin("/view/data/lift_group", liftGroup),
				DeleteEndpoint: urlPathJoin("/view/data/lift_group", liftGroup),
				Values: []templates.DataTableValue{
					{Name: "id", Value: liftGroup, Type: templates.InputString},
				},
			}, nil
		},
	)
}

func addRoutes(
	_ context.Context,
	mux *http.ServeMux,
	wDB, roDB *sql.DB,
	calendarService *calendar.Service) {

	// Prometheus metrics.
	mux.Handle("/metrics", promhttp.Handler())

	// Vendor and non-vendor static assets.
	mux.Handle("GET /web/static/", http.FileServer(http.Dir(".")))
	mux.Handle("GET /web/vendor/", http.FileServer(http.Dir(".")))

	// Raw single and bulk insert endpoints.
	mux.HandleFunc("POST /api/v1/raw/lift", handleRawPost(wDB, (*workoutdb.Queries).RawInsertLift))
	mux.HandleFunc("POST /api/v1/raw/lift_muscle", handleRawPost(wDB, (*workoutdb.Queries).RawInsertLiftMuscle))
	mux.HandleFunc("POST /api/v1/raw/lift_workout", handleRawPost(wDB, (*workoutdb.Queries).RawInsertLiftWorkout))
	mux.HandleFunc("POST /api/v1/raw/movement", handleRawPost(wDB, (*workoutdb.Queries).RawInsertMovement))
	mux.HandleFunc("POST /api/v1/raw/muscle", handleRawPost(wDB, (*workoutdb.Queries).RawInsertMuscle))
	mux.HandleFunc("POST /api/v1/raw/progress", handleRawPost(wDB, (*workoutdb.Queries).RawInsertProgress))
	mux.HandleFunc("POST /api/v1/raw/routine", handleRawPost(wDB, (*workoutdb.Queries).RawInsertRoutine))
	mux.HandleFunc("POST /api/v1/raw/routine_workout", handleRawPost(wDB, (*workoutdb.Queries).RawInsertRoutineWorkout))
	mux.HandleFunc("POST /api/v1/raw/schedule", handleRawPost(wDB, (*workoutdb.Queries).RawInsertSchedule))
	mux.HandleFunc("POST /api/v1/raw/schedule_list", handleRawPost(wDB, (*workoutdb.Queries).RawInsertScheduleList))
	mux.HandleFunc("POST /api/v1/raw/side_weight", handleRawPost(wDB, (*workoutdb.Queries).RawInsertSideWeight))
	mux.HandleFunc("POST /api/v1/raw/subworkout", handleRawPost(wDB, (*workoutdb.Queries).RawInsertSubworkout))
	mux.HandleFunc("POST /api/v1/raw/template_variable", handleRawPost(wDB, (*workoutdb.Queries).RawInsertTemplateVariable))
	mux.HandleFunc("POST /api/v1/raw/workout", handleRawPost(wDB, (*workoutdb.Queries).RawInsertWorkout))

	ts := fmt.Sprint(todaysDate().Unix())

	// Index pages.
	// TODO: Use hash of css for cache busting instead of date.
	mux.HandleFunc("GET /{$}", handleIndexPage("main", ts, calendarService))
	mux.HandleFunc("GET /schedule/{$}", handleIndexPage("schedule", ts, calendarService))
	mux.HandleFunc("GET /data/{$}", handleIndexPage("data", ts, calendarService))
	mux.HandleFunc("GET /data/{tabX}/{tabY}", handleIndexPage("data", ts, calendarService))

	// Main view.
	mux.Handle("GET /view/tabs/main", alertMiddleware(handleMainTab(roDB)))
	mux.Handle("GET /view/calendarauthurl", alertMiddleware(handleGetCalendarAuthURL(calendarService)))
	mux.Handle("GET /view/liftgroups", alertMiddleware(handleGetLiftGroupListView(roDB)))

	// Progress table
	mux.Handle("GET /view/progresstable", alertMiddleware(handleGetProgressTable(roDB)))
	mux.Handle("DELETE /view/progresstablerow/{id}", alertMiddleware(handleDeleteProgress(wDB)))
	mux.Handle("POST /view/progresstablerow", alertMiddleware(handleCreateProgress(wDB)))

	// Routine table.
	mux.Handle("GET /view/routinetable", alertMiddleware(handleGetRoutineTable(roDB)))

	// Progress form.
	mux.Handle("GET /view/liftselect", alertMiddleware(handleGetLiftSelect(roDB)))
	mux.Handle("GET /view/sideweightselect", alertMiddleware(handleGetSideWeightSelect(roDB)))
	mux.Handle("GET /view/progressform", alertMiddleware(handleGetProgressForm()))
	mux.Handle("POST /view/progressform", alertMiddleware(handleCreateProgressForm(roDB)))

	// Schedule view.
	mux.Handle("GET /view/tabs/schedule", alertMiddleware(handleGetScheduleTable(roDB)))

	// Schedule table.
	mux.Handle("DELETE /view/schedule/{date}", alertMiddleware(handleDeleteSchedule(wDB, calendarService)))
	mux.Handle("PATCH /view/schedule/{date}", alertMiddleware(handlePatchScheduleTableRow(wDB, calendarService)))
	mux.Handle("POST /view/scheduleappend", alertMiddleware(handlePostScheduleTableRow(wDB, calendarService)))
	mux.Handle("POST /view/scheduletablerows", alertMiddleware(handlePostScheduleTableRows(wDB, calendarService)))

	// Data view
	mux.Handle("GET /view/tabs/data/{$}", alertMiddleware(handleGetDataTabView()))
	mux.Handle("GET /view/tabs/data/{tabX}/{tabY}", alertMiddleware(handleGetDataTabView()))

	// Lift table view
	mux.Handle("POST /view/data/lift", alertMiddleware(handlePostLiftView(roDB, wDB)))
	mux.Handle("PATCH /view/data/lift", alertMiddleware(handlePatchLiftView(wDB)))
	mux.Handle("PATCH /view/data/lift/{id}", alertMiddleware(handlePatchLiftView(wDB)))
	mux.Handle("DELETE /view/data/lift",
		alertMiddleware(handleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteLift)))
	mux.Handle("DELETE /view/data/lift/{id}",
		alertMiddleware(handleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteLift)))
	mux.Handle("GET /view/data/lift", alertMiddleware(handleGetLiftView(roDB)))

	// Movement table view
	mux.Handle("POST /view/data/movement", alertMiddleware(handlePostMovementView(roDB, wDB)))
	mux.Handle("PATCH /view/data/movement", alertMiddleware(handlePatchMovementView(wDB)))
	mux.Handle("PATCH /view/data/movement/{id}", alertMiddleware(handlePatchMovementView(wDB)))
	mux.Handle("DELETE /view/data/movement",
		alertMiddleware(handleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteMovement)))
	mux.Handle("DELETE /view/data/movement/{id}",
		alertMiddleware(handleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteMovement)))
	mux.Handle("GET /view/data/movement", alertMiddleware(handleGetMovementView(roDB)))

	// Muscle table view
	mux.Handle("POST /view/data/muscle", alertMiddleware(handlePostMuscleView(roDB, wDB)))
	mux.Handle("PATCH /view/data/muscle", alertMiddleware(handlePatchMuscleView(wDB)))
	mux.Handle("PATCH /view/data/muscle/{id}", alertMiddleware(handlePatchMuscleView(wDB)))
	mux.Handle("DELETE /view/data/muscle",
		alertMiddleware(handleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteMuscle)))
	mux.Handle("DELETE /view/data/muscle/{id}",
		alertMiddleware(handleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteMuscle)))
	mux.Handle("GET /view/data/muscle", alertMiddleware(handleGetMuscleView(roDB)))

	// Routine table view
	mux.Handle("POST /view/data/routine", alertMiddleware(handlePostRoutineView(roDB, wDB)))
	mux.Handle("PATCH /view/data/routine", alertMiddleware(handlePatchRoutineView(wDB)))
	mux.Handle("PATCH /view/data/routine/{id}", alertMiddleware(handlePatchRoutineView(wDB)))
	mux.Handle("DELETE /view/data/routine",
		alertMiddleware(handleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteRoutine)))
	mux.Handle("DELETE /view/data/routine/{id}",
		alertMiddleware(handleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteRoutine)))
	mux.Handle("GET /view/data/routine", alertMiddleware(handleGetRoutineView(roDB)))

	// Schedule list view
	mux.Handle("POST /view/data/schedule_list", alertMiddleware(handlePostScheduleListView(roDB, wDB)))
	mux.Handle("PATCH /view/data/schedule_list", alertMiddleware(handlePatchScheduleListView(wDB)))
	mux.Handle("PATCH /view/data/schedule_list/{id}", alertMiddleware(handlePatchScheduleListView(wDB)))
	mux.Handle("DELETE /view/data/schedule_list",
		alertMiddleware(handleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteScheduleList)))
	mux.Handle("DELETE /view/data/schedule_list/{id}",
		alertMiddleware(handleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteScheduleList)))
	mux.Handle("GET /view/data/schedule_list", alertMiddleware(handleGetScheduleListView(roDB)))

	// Side weight view
	mux.Handle("POST /view/data/side_weight", alertMiddleware(handlePostSideWeightView(roDB, wDB)))
	mux.Handle("PATCH /view/data/side_weight", alertMiddleware(handlePatchSideWeightView(wDB)))
	mux.Handle("PATCH /view/data/side_weight/{id}", alertMiddleware(handlePatchSideWeightView(wDB)))
	mux.Handle("DELETE /view/data/side_weight",
		alertMiddleware(handleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteSideWeight)))
	mux.Handle("DELETE /view/data/side_weight/{id}",
		alertMiddleware(handleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteSideWeight)))
	mux.Handle("GET /view/data/side_weight", alertMiddleware(handleGetSideWeightView(roDB)))

	// Template variable view
	mux.Handle("POST /view/data/template_variable", alertMiddleware(handlePostTemplateVariableView(roDB, wDB)))
	mux.Handle("PATCH /view/data/template_variable", alertMiddleware(handlePatchTemplateVariableView(wDB)))
	mux.Handle("PATCH /view/data/template_variable/{id}", alertMiddleware(handlePatchTemplateVariableView(wDB)))
	mux.Handle("DELETE /view/data/template_variable",
		alertMiddleware(handleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteTemplateVariable)))
	mux.Handle("DELETE /view/data/template_variable/{id}",
		alertMiddleware(handleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteTemplateVariable)))
	mux.Handle("GET /view/data/template_variable", alertMiddleware(handleGetTemplateVariableView(roDB)))

	// Workout view
	mux.Handle("POST /view/data/workout", alertMiddleware(handlePostWorkoutView(roDB, wDB)))
	mux.Handle("PATCH /view/data/workout", alertMiddleware(handlePatchWorkoutView(wDB)))
	mux.Handle("PATCH /view/data/workout/{id}", alertMiddleware(handlePatchWorkoutView(wDB)))
	mux.Handle("DELETE /view/data/workout",
		alertMiddleware(handleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteWorkout)))
	mux.Handle("DELETE /view/data/workout/{id}",
		alertMiddleware(handleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteWorkout)))
	mux.Handle("GET /view/data/workout", alertMiddleware(handleGetWorkoutView(roDB)))

	// Schedule view
	mux.Handle("POST /view/data/schedule", alertMiddleware(handlePostScheduleView(roDB, wDB)))
	mux.Handle("PATCH /view/data/schedule", alertMiddleware(handlePatchScheduleView(wDB)))
	mux.Handle("PATCH /view/data/schedule/{id}", alertMiddleware(handlePatchScheduleView(wDB)))
	mux.Handle("DELETE /view/data/schedule",
		alertMiddleware(handleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteSchedule)))
	mux.Handle("DELETE /view/data/schedule/{id}",
		alertMiddleware(handleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteSchedule)))
	mux.Handle("GET /view/data/schedule", alertMiddleware(handleGetScheduleView(roDB)))

	// Progress view
	mux.Handle("POST /view/data/progress", alertMiddleware(handlePostProgressView(roDB, wDB)))
	mux.Handle("PATCH /view/data/progress", alertMiddleware(handlePatchProgressView(wDB)))
	mux.Handle("PATCH /view/data/progress/{id}", alertMiddleware(handlePatchProgressView(wDB)))
	mux.Handle("DELETE /view/data/progress/{id}",
		alertMiddleware(handleDeleteTableRowViewRequest(
			wDB, (*workoutdb.Queries).RawDeleteProgress,
			func(r *http.Request) (*int64, error) {
				id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
				if err != nil {
					return nil, fmt.Errorf("failed to parse id: %w", err)
				}
				return &id, nil
			},
		)))
	mux.Handle("GET /view/data/progress", alertMiddleware(handleGetProgressView(roDB)))

	// Lift muscle mapping view
	mux.Handle("POST /view/data/lift_muscle_mapping", alertMiddleware(handlePostLiftMuscleView(roDB, wDB)))
	mux.Handle("PATCH /view/data/lift_muscle_mapping/{lift}/{muscle}/{movement}", alertMiddleware(handlePatchLiftMuscleView(wDB)))
	mux.Handle("DELETE /view/data/lift_muscle_mapping/{lift}/{muscle}/{movement}",
		alertMiddleware(handleDeleteTableRowViewRequest(
			wDB, (*workoutdb.Queries).RawDeleteLiftMuscle,
			func(r *http.Request) (*workoutdb.RawDeleteLiftMuscleParams, error) {
				return &workoutdb.RawDeleteLiftMuscleParams{
					Lift:     r.PathValue("lift"),
					Muscle:   r.PathValue("muscle"),
					Movement: r.PathValue("movement"),
				}, nil
			},
		)))
	mux.Handle("GET /view/data/lift_muscle_mapping", alertMiddleware(handleGetLiftMuscleView(roDB)))

	// Lift workout mapping view
	mux.Handle("POST /view/data/lift_workout_mapping", alertMiddleware(handlePostLiftWorkoutView(roDB, wDB)))
	mux.Handle("PATCH /view/data/lift_workout_mapping/{lift}/{workout}", alertMiddleware(handlePatchLiftWorkoutView(wDB)))
	mux.Handle("DELETE /view/data/lift_workout_mapping/{lift}/{workout}",
		alertMiddleware(handleDeleteTableRowViewRequest(
			wDB, (*workoutdb.Queries).RawDeleteLiftWorkout,
			func(r *http.Request) (*workoutdb.RawDeleteLiftWorkoutParams, error) {
				return &workoutdb.RawDeleteLiftWorkoutParams{
					Lift:    r.PathValue("lift"),
					Workout: r.PathValue("workout"),
				}, nil
			},
		)))
	mux.Handle("GET /view/data/lift_workout_mapping", alertMiddleware(handleGetLiftWorkoutView(roDB)))

	// Routine workout mapping view
	mux.Handle("POST /view/data/routine_workout_mapping", alertMiddleware(handlePostRoutineWorkoutView(roDB, wDB)))
	mux.Handle("PATCH /view/data/routine_workout_mapping/{routine}/{workout}", alertMiddleware(handlePatchRoutineWorkoutView(wDB)))
	mux.Handle("DELETE /view/data/routine_workout_mapping/{routine}/{workout}",
		alertMiddleware(handleDeleteTableRowViewRequest(
			wDB, (*workoutdb.Queries).RawDeleteRoutineWorkout,
			func(r *http.Request) (*workoutdb.RawDeleteRoutineWorkoutParams, error) {
				return &workoutdb.RawDeleteRoutineWorkoutParams{
					Routine: r.PathValue("routine"),
					Workout: r.PathValue("workout"),
				}, nil
			},
		)))
	mux.Handle("GET /view/data/routine_workout_mapping", alertMiddleware(handleGetRoutineWorkoutView(roDB)))

	// Subworkout mapping view
	mux.Handle("POST /view/data/subworkout", alertMiddleware(handlePostSubworkoutView(roDB, wDB)))
	mux.Handle("PATCH /view/data/subworkout/{subworkout}/{superworkout}", alertMiddleware(handlePatchSubworkoutView(wDB)))
	mux.Handle("DELETE /view/data/subworkout/{subworkout}/{superworkout}",
		alertMiddleware(handleDeleteTableRowViewRequest(
			wDB, (*workoutdb.Queries).RawDeleteSubworkout,
			func(r *http.Request) (*workoutdb.RawDeleteSubworkoutParams, error) {
				return &workoutdb.RawDeleteSubworkoutParams{
					Subworkout:   r.PathValue("subworkout"),
					Superworkout: r.PathValue("superworkout"),
				}, nil
			},
		)))
	mux.Handle("GET /view/data/subworkout", alertMiddleware(handleGetSubworkoutView(roDB)))

	// Lift group table view
	mux.Handle("POST /view/data/lift_group", alertMiddleware(handlePostLiftGroupView(roDB, wDB)))
	mux.Handle("PATCH /view/data/lift_group", alertMiddleware(handlePatchLiftGroupView(wDB)))
	mux.Handle("PATCH /view/data/lift_group/{id}", alertMiddleware(handlePatchLiftGroupView(wDB)))
	mux.Handle("DELETE /view/data/lift_group",
		alertMiddleware(handleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteLiftGroup)))
	mux.Handle("DELETE /view/data/lift_group/{id}",
		alertMiddleware(handleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteLiftGroup)))
	mux.Handle("GET /view/data/lift_group", alertMiddleware(handleGetLiftGroupView(roDB)))
}
