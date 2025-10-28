package index

import (
	"fmt"
	"log/slog"
	"net/http"
	"path"
	"slices"
	"strconv"
	"time"

	"github.com/RyRose/uplog/internal/config"
	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
	"github.com/RyRose/uplog/internal/templates"
)

func todaysDate() time.Time {
	return time.Now()
}

// HandleIndexPage godoc
//
//	@Summary		Get index page
//	@Description	Renders the main index page with specified tab and CSS query parameters
//	@Tags			index
//	@Produce		html
//	@Param			tabX	path		string	false	"Tab X parameter"
//	@Param			tabY	path		string	false	"Tab Y parameter"
//	@Success		200		{string}	string	"HTML content"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/ [get]
//	@Router			/data [get]
//	@Router			/data/{tabX}/{tabY} [get]
func HandleIndexPage(tab string, cfg *config.Data, state *config.State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		err := templates.IndexPage(
			*cfg.Version,
			path.Join("/view/tabs", tab, r.PathValue("tabX"), r.PathValue("tabY")),
		).Render(ctx, w)
		if err != nil {
			http.Error(w, "failed to write response", http.StatusInternalServerError)
			slog.ErrorContext(ctx, "failed to write response", "error", err)
		}
	}
}

// HandleMainTab godoc
//
//	@Summary		Get main tab view
//	@Description	Renders the main tab view with progress for today and lift groups
//	@Tags			index
//	@Produce		html
//	@Success		200	{string}	string	"HTML content"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/view/tabs/main [get]
func HandleMainTab(_ *config.Data, state *config.State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		date := todaysDate()
		queries := workoutdb.New(state.RDB)

		ps, err := queries.ListProgressForDay(ctx, date.Format(time.DateOnly))
		if err != nil {
			slog.ErrorContext(ctx, "failed to retrieve progress", "date", date, "error", err)
			http.Error(
				w,
				fmt.Sprintf("failed to retrieve progress for %v: %v", date, err),
				http.StatusInternalServerError)
			return
		}

		lgs, err := queries.QueryLiftGroupsForDate(ctx, date.Format(time.DateOnly))
		if err != nil {
			http.Error(w, "failed to query lift groups", http.StatusInternalServerError)
			slog.ErrorContext(ctx, "failed to query lift groups", "error", err)
			return
		}

		templates.MainView(templates.MainViewData{
			Progress:   ps,
			LiftGroups: lgs,
		}).Render(ctx, w)
	}
}

// HandleGetLiftGroupListView godoc
//
//	@Summary		Get lift group list view
//	@Description	Renders the list of lift groups for today
//	@Tags			index
//	@Produce		html
//	@Success		200	{string}	string	"HTML content"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/view/liftgroups [get]
func HandleGetLiftGroupListView(_ *config.Data, state *config.State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		date := todaysDate()
		queries := workoutdb.New(state.RDB)

		lgs, err := queries.QueryLiftGroupsForDate(ctx, date.Format(time.DateOnly))
		if err != nil {
			http.Error(w, "failed to query lift groups", http.StatusInternalServerError)
			slog.ErrorContext(ctx, "failed to query lift groups", "error", err)
			return
		}

		templates.LiftGroupList(lgs).Render(ctx, w)
	}
}

// HandleGetProgressTable godoc
//
//	@Summary		Get progress table
//	@Description	Renders the progress table for today
//	@Tags			index
//	@Produce		html
//	@Success		200	{string}	string	"HTML content"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/view/progresstable [get]
func HandleGetProgressTable(_ *config.Data, state *config.State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		date := todaysDate()
		queries := workoutdb.New(state.RDB)
		ps, err := queries.ListProgressForDay(ctx, date.Format(time.DateOnly))
		if err != nil {
			slog.ErrorContext(ctx, "failed to retrieve progress", "date", date, "error", err)
			http.Error(
				w,
				fmt.Sprintf("failed to retrieve progress for %v: %v", date, err),
				http.StatusInternalServerError)
			return
		}
		templates.ProgressTable(ps).Render(ctx, w)
	}
}

// HandleDeleteProgress godoc
//
//	@Summary		Delete progress entry
//	@Description	Deletes a progress entry by ID
//	@Tags			index
//	@Param			id	path		int		true	"Progress ID"
//	@Success		200	{string}	string	"OK"
//	@Failure		400	{string}	string	"Bad request"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/view/progresstablerow/{id} [delete]
func HandleDeleteProgress(_ *config.Data, state *config.State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		idStr := r.PathValue("id")
		queries := workoutdb.New(state.WDB)
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("failed to parse id (%v): %v", idStr, err),
				http.StatusBadRequest)
			return
		}
		if err := queries.DeleteProgress(ctx, idInt); err != nil {
			http.Error(
				w,
				fmt.Sprintf("failed to delete progress with id (%v): %v", idInt, err),
				http.StatusInternalServerError,
			)
			return
		}
		w.Header().Set("HX-Trigger", "deleteProgress")
	}
}

// HandleCreateProgress godoc
//
//	@Summary		Create progress entry
//	@Description	Creates a new progress entry for today
//	@Tags			index
//	@Accept			x-www-form-urlencoded
//	@Produce		html
//	@Param			lift	formData	string	true	"Lift ID"
//	@Param			weight	formData	number	true	"Weight"
//	@Param			sets	formData	integer	true	"Number of sets"
//	@Param			reps	formData	integer	true	"Number of reps"
//	@Param			side	formData	string	false	"Side weight"
//	@Success		200		{string}	string	"HTML content"
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/view/progresstablerow [post]
func HandleCreateProgress(_ *config.Data, state *config.State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		queries := workoutdb.New(state.WDB)

		weight, err := strconv.ParseFloat(r.PostFormValue("weight"), 64)
		if err != nil {
			http.Error(w, "failed to parse weight", http.StatusBadRequest)
			return
		}

		sets, err := strconv.Atoi(r.PostFormValue("sets"))
		if err != nil {
			http.Error(w, "failed to parse sets", http.StatusBadRequest)
			return
		}

		reps, err := strconv.Atoi(r.PostFormValue("reps"))
		if err != nil {
			http.Error(w, "failed to parse reps", http.StatusBadRequest)
			return
		}

		params := workoutdb.InsertProgressParams{
			Lift:       r.PostFormValue("lift"),
			Date:       todaysDate().Format(time.DateOnly),
			Weight:     weight,
			Sets:       int64(sets),
			Reps:       int64(reps),
			SideWeight: r.PostFormValue("side"),
		}
		progress, err := queries.InsertProgress(ctx, params)
		if err != nil {
			slog.ErrorContext(ctx, "failed to insert progress", "error", err, "params", params)
			http.Error(w, fmt.Sprintf("Failed to insert progress: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("HX-Trigger", "newProgress")
		templates.ProgressTableRow(progress).Render(ctx, w)
	}
}

// HandleGetRoutineTable godoc
//
//	@Summary		Get routine table
//	@Description	Returns empty response as routine table functionality was removed
//	@Tags			index
//	@Success		200	{string}	string	"OK"
//	@Router			/view/routinetable [get]
func HandleGetRoutineTable(_ *config.Data, state *config.State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Routine table functionality removed - schedule concept eliminated
		w.WriteHeader(http.StatusOK)
	}
}

// HandleGetLiftSelect godoc
//
//	@Summary		Get lift select dropdown
//	@Description	Renders a select dropdown of lifts grouped by lift group
//	@Tags			index
//	@Produce		html
//	@Param			name	query		string	false	"Input name attribute"
//	@Param			lift	query		string	false	"Selected lift ID"
//	@Success		200		{string}	string	"HTML content"
//	@Router			/view/liftselect [get]
func HandleGetLiftSelect(_ *config.Data, state *config.State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		queries := workoutdb.New(state.RDB)
		lifts, err := queries.RawSelectLift(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "failed to list all lifts", "error", err)
			return
		}

		mapLifts := make(map[string][]string)
		for _, lift := range lifts {
			var lg string
			if lift.LiftGroup != nil {
				lg = *lift.LiftGroup
			}
			if lg == "" {
				continue
			}
			mapLifts[lg] = append(mapLifts[lg], lift.ID)
		}

		var lgs []string
		for lg := range mapLifts {
			lgs = append(lgs, lg)
		}

		slices.Sort(lgs)
		var groups []templates.LiftSelectGroup
		for _, lg := range lgs {
			groups = append(groups, templates.LiftSelectGroup{
				Name:    lg,
				Options: mapLifts[lg],
			})
		}

		templates.LiftSelect(
			r.URL.Query().Get("name"),
			r.URL.Query().Get("lift"),
			groups,
		).Render(ctx, w)
	}
}

// HandleGetSideWeightSelect godoc
//
//	@Summary		Get side weight select dropdown
//	@Description	Renders a select dropdown of side weights with default selection based on lift
//	@Tags			index
//	@Produce		html
//	@Param			name	query		string	false	"Input name attribute"
//	@Param			lift	query		string	false	"Lift ID to get default side weight"
//	@Success		200		{string}	string	"HTML content"
//	@Router			/view/sideweightselect [get]
func HandleGetSideWeightSelect(_ *config.Data, state *config.State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		queries := workoutdb.New(state.RDB)
		name := r.URL.Query().Get("name")
		sideWeights, err := queries.ListAllIndividualSideWeights(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "failed to list all side weights", "error", err)
			return
		}

		liftParam := r.URL.Query().Get("lift")
		if liftParam == "" {
			templates.SideweightSelect(name, "", sideWeights).Render(ctx, w)
			return
		}

		lift, err := queries.GetLift(ctx, liftParam)
		if err != nil {
			slog.WarnContext(ctx, "failed to get lift", "lift", liftParam, "error", err)
			templates.SideweightSelect(name, "", sideWeights).Render(ctx, w)
			return
		}

		selected := ""
		if lift.DefaultSideWeight != nil {
			selected = *lift.DefaultSideWeight
		}
		templates.SideweightSelect(
			name,
			selected,
			sideWeights).Render(ctx, w)
	}
}

// HandleGetProgressForm godoc
//
//	@Summary		Get progress form
//	@Description	Renders an empty progress form
//	@Tags			index
//	@Produce		html
//	@Success		200	{string}	string	"HTML content"
//	@Router			/view/progressform [get]
func HandleGetProgressForm() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templates.ProgressForm(templates.ProgressFormData{}).Render(r.Context(), w)
	}
}

// HandleCreateProgressForm godoc
//
//	@Summary		Create progress form with recent data
//	@Description	Renders a progress form pre-filled with recent progress data for the selected lift
//	@Tags			index
//	@Accept			x-www-form-urlencoded
//	@Produce		html
//	@Param			lift	formData	string	false	"Lift ID"
//	@Param			side	formData	string	false	"Side weight"
//	@Param			weight	formData	string	false	"Weight"
//	@Param			sets	formData	string	false	"Sets"
//	@Param			reps	formData	string	false	"Reps"
//	@Success		200		{string}	string	"HTML content"
//	@Router			/view/progressform [post]
func HandleCreateProgressForm(_ *config.Data, state *config.State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		queries := workoutdb.New(state.RDB)

		lift := r.PostFormValue("lift")
		var progress []workoutdb.Progress
		if lift != "" {
			var err error
			progress, err = queries.ListMostRecentProgressForLift(ctx,
				workoutdb.ListMostRecentProgressForLiftParams{
					Lift:  lift,
					Limit: 5,
				})
			if err != nil {
				slog.WarnContext(ctx, "failed to list most recent progress for lift", "lift", lift, "error", err)
				progress = nil
			}
		}
		templates.ProgressForm(templates.ProgressFormData{
			Lift:       lift,
			SideWeight: r.PostFormValue("side"),
			Weight:     r.PostFormValue("weight"),
			Sets:       r.PostFormValue("sets"),
			Reps:       r.PostFormValue("reps"),
			Progress:   progress,
		}).Render(ctx, w)
	}
}
