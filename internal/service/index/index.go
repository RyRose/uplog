package index

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"path"
	"slices"
	"strconv"
	"time"

	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
	"github.com/RyRose/uplog/internal/templates"
)

func todaysDate() time.Time {
	return time.Now()
}

func HandleIndexPage(tab, cssQuery string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		err := templates.IndexPage(
			cssQuery,
			path.Join("/view/tabs", tab, r.PathValue("tabX"), r.PathValue("tabY")),
		).Render(ctx, w)
		if err != nil {
			http.Error(w, "failed to write response", http.StatusInternalServerError)
			slog.ErrorContext(ctx, "failed to write response", "error", err)
		}
	}
}

func HandleMainTab(roDB *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		date := todaysDate()
		queries := workoutdb.New(roDB)

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

func HandleGetLiftGroupListView(roDB *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		date := todaysDate()
		queries := workoutdb.New(roDB)

		lgs, err := queries.QueryLiftGroupsForDate(ctx, date.Format(time.DateOnly))
		if err != nil {
			http.Error(w, "failed to query lift groups", http.StatusInternalServerError)
			slog.ErrorContext(ctx, "failed to query lift groups", "error", err)
			return
		}

		templates.LiftGroupList(lgs).Render(ctx, w)
	}
}

func HandleGetProgressTable(roDB *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		date := todaysDate()
		queries := workoutdb.New(roDB)
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

func HandleDeleteProgress(wDB *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		idStr := r.PathValue("id")
		queries := workoutdb.New(wDB)
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

func HandleCreateProgress(wDB *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		queries := workoutdb.New(wDB)

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

func HandleGetRoutineTable(roDB *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Routine table functionality removed - schedule concept eliminated
		w.WriteHeader(http.StatusOK)
	}
}

func HandleGetLiftSelect(roDB *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		queries := workoutdb.New(roDB)
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

func HandleGetSideWeightSelect(roDB *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		queries := workoutdb.New(roDB)
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

func HandleGetProgressForm() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templates.ProgressForm(templates.ProgressFormData{}).Render(r.Context(), w)
	}
}

func HandleCreateProgressForm(roDB *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		queries := workoutdb.New(roDB)

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
