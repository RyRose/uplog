package service

import (
	"database/sql"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
	"github.com/RyRose/uplog/internal/templates"
)

func handleMainTab(roDB *sql.DB) http.HandlerFunc {
	roundToNearest := func(num, unit float64) float64 {
		return math.Round(num/unit) * unit
	}
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

		workout, err := queries.GetWorkoutForDate(ctx, date.Format(time.DateOnly))
		if err != nil {
			slog.WarnContext(ctx, "failed to get workout for date", "date", date, "error", err)
			workout = nil
		}

		if workout != nil {
			templates.ProgressTitle(fmt.Sprint(workout)).Render(ctx, w)
		}

		routines, err := queries.ListRoutinesForDate(ctx, date.Format(time.DateOnly))
		if err != nil {
			slog.ErrorContext(ctx, "failed to list routines for date", "date", date, "error", err)
			http.Error(w, "failed to list routines for date", http.StatusInternalServerError)
			return
		}

		var tables []templates.RoutineTable
		for _, routine := range routines {
			progress, err := queries.GetLatestProgressForLift(ctx, routine.Lift)
			if err != nil {
				slog.ErrorContext(ctx,
					"failed to get latest progress for lift", "lift", routine.Lift, "error", err)
				http.Error(w, "failed to get latest progress for lift", http.StatusInternalServerError)
				return
			}
			rows := templates.RoutineTable{SideWeight: progress.SideWeight.ID, Lift: routine.Lift}
			steps := strings.Split(routine.Steps, ",")
			for _, step := range steps {
				repsSetsPercent := strings.Split(step, "@")
				if len(repsSetsPercent) != 2 {
					slog.WarnContext(ctx,
						"length of reps/percent not 2",
						"repsSetsPercent", repsSetsPercent, "step", step, "steps", steps)
					continue
				}
				repsSets := strings.Split(repsSetsPercent[0], "x")
				reps := repsSets[0]
				var sets string
				if len(repsSets) == 1 {
					sets = "1"
				} else {
					sets = repsSets[1]
				}
				pct, err := strconv.Atoi(strings.TrimSuffix(repsSetsPercent[1], "%"))
				if err != nil {
					slog.WarnContext(ctx,
						"pct not an integer",
						"repsPercent", repsSetsPercent, "step", step, "steps", steps)
					continue
				}
				total := (progress.Progress.Weight*progress.SideWeight.Multiplier +
					progress.SideWeight.Addend) * float64(pct) / 100.0
				rows.Rows = append(rows.Rows,
					templates.RoutineTableRow{
						Percent: repsSetsPercent[1],
						Weight: strconv.FormatFloat(
							roundToNearest(total, 5),
							'f', 0, 64,
						),
						SideWeight: strconv.FormatFloat(
							roundToNearest(
								(total-progress.SideWeight.Addend)/progress.SideWeight.Multiplier, 2.5),
							'f', 1, 64),
						Reps: reps,
						Sets: sets,
					},
				)
			}
			total := (progress.Progress.Weight*progress.SideWeight.Multiplier +
				progress.SideWeight.Addend)
			rows.Rows = append(rows.Rows, templates.RoutineTableRow{
				Percent: "100%",
				Weight: strconv.FormatFloat(
					roundToNearest(total, 5),
					'f', 0, 64,
				),
				SideWeight: strconv.FormatFloat(
					roundToNearest((total-progress.SideWeight.Addend)/progress.SideWeight.Multiplier, 2.5),
					'f', 1, 64),
			})
			tables = append(tables, rows)
		}

		lgs, err := queries.QueryLiftGroupsForDate(ctx, date.Format(time.DateOnly))
		if err != nil {
			http.Error(w, "failed to query lift groups", http.StatusInternalServerError)
			slog.ErrorContext(ctx, "failed to query lift groups", "error", err)
			return
		}

		templates.MainView(templates.MainViewData{
			Routines:   tables,
			Progress:   ps,
			LiftGroups: lgs,
		}).Render(ctx, w)
	}
}

func handleGetLiftGroupListView(roDB *sql.DB) http.HandlerFunc {
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

func handleGetProgressTable(roDB *sql.DB) http.HandlerFunc {
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

func handleDeleteProgress(wDB *sql.DB) http.HandlerFunc {
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

func handleCreateProgress(wDB *sql.DB) http.HandlerFunc {
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

func handleGetRoutineTable(roDB *sql.DB) http.HandlerFunc {
	roundToNearest := func(num, unit float64) float64 {
		return math.Round(num/unit) * unit
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		queries := workoutdb.New(roDB)

		date := todaysDate()
		workout, err := queries.GetWorkoutForDate(ctx, date.Format(time.DateOnly))
		if err != nil {
			slog.WarnContext(ctx, "failed to get workout for date", "date", date, "error", err)
			workout = nil
		}

		if workout != nil {
			templates.ProgressTitle(fmt.Sprint(workout)).Render(ctx, w)
		}

		routines, err := queries.ListRoutinesForDate(ctx, date.Format(time.DateOnly))
		if err != nil {
			slog.ErrorContext(ctx, "failed to list routines for date", "date", date, "error", err)
			http.Error(w, "failed to list routines for date", http.StatusInternalServerError)
			return
		}

		for _, routine := range routines {
			progress, err := queries.GetLatestProgressForLift(ctx, routine.Lift)
			if err != nil {
				slog.ErrorContext(ctx,
					"failed to get latest progress for lift", "lift", routine.Lift, "error", err)
				http.Error(w, "failed to get latest progress for lift", http.StatusInternalServerError)
				return
			}
			rows := templates.RoutineTable{SideWeight: progress.SideWeight.ID, Lift: routine.Lift}
			steps := strings.Split(routine.Steps, ",")
			for _, step := range steps {
				repsSetsPercent := strings.Split(step, "@")
				if len(repsSetsPercent) != 2 {
					slog.WarnContext(ctx,
						"length of reps/percent not 2",
						"repsSetsPercent", repsSetsPercent, "step", step, "steps", steps)
					continue
				}
				repsSets := strings.Split(repsSetsPercent[0], "x")
				reps := repsSets[0]
				var sets string
				if len(repsSets) == 1 {
					sets = "1"
				} else {
					sets = repsSets[1]
				}
				pct, err := strconv.Atoi(strings.TrimSuffix(repsSetsPercent[1], "%"))
				if err != nil {
					slog.WarnContext(ctx,
						"pct not an integer",
						"repsPercent", repsSetsPercent, "step", step, "steps", steps)
					continue
				}
				total := (progress.Progress.Weight*progress.SideWeight.Multiplier +
					progress.SideWeight.Addend) * float64(pct) / 100.0
				rows.Rows = append(rows.Rows,
					templates.RoutineTableRow{
						Percent: repsSetsPercent[1],
						Weight: strconv.FormatFloat(
							roundToNearest(total, 5),
							'f', 0, 64,
						),
						SideWeight: strconv.FormatFloat(
							roundToNearest(
								(total-progress.SideWeight.Addend)/progress.SideWeight.Multiplier, 2.5),
							'f', 1, 64),
						Reps: reps,
						Sets: sets,
					},
				)
			}
			total := (progress.Progress.Weight*progress.SideWeight.Multiplier +
				progress.SideWeight.Addend)
			rows.Rows = append(rows.Rows, templates.RoutineTableRow{
				Percent: "100%",
				Weight: strconv.FormatFloat(
					roundToNearest(total, 5),
					'f', 0, 64,
				),
				SideWeight: strconv.FormatFloat(
					roundToNearest((total-progress.SideWeight.Addend)/progress.SideWeight.Multiplier, 2.5),
					'f', 1, 64),
			})
			templates.RoutineTableView(rows).Render(ctx, w)
		}
	}
}

func handleGetLiftSelect(roDB *sql.DB) http.HandlerFunc {
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

func handleGetSideWeightSelect(roDB *sql.DB) http.HandlerFunc {
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

func handleGetProgressForm() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templates.ProgressForm(templates.ProgressFormData{}).Render(r.Context(), w)
	}
}

func handleCreateProgressForm(roDB *sql.DB) http.HandlerFunc {
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
