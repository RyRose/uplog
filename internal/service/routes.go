package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/RyRose/uplog/internal/calendar"
	"github.com/RyRose/uplog/internal/service/rawdata"
	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
	"github.com/RyRose/uplog/internal/templates"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// TODO: Parameterize today's date for testing support.
func todaysDate() time.Time {
	return time.Now()
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
	mux.Handle("POST /view/data/lift", alertMiddleware(rawdata.HandlePostLiftView(roDB, wDB)))
	mux.Handle("PATCH /view/data/lift", alertMiddleware(rawdata.HandlePatchLiftView(wDB)))
	mux.Handle("PATCH /view/data/lift/{id}", alertMiddleware(rawdata.HandlePatchLiftView(wDB)))
	mux.Handle("DELETE /view/data/lift",
		alertMiddleware(rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteLift)))
	mux.Handle("DELETE /view/data/lift/{id}",
		alertMiddleware(rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteLift)))
	mux.Handle("GET /view/data/lift", alertMiddleware(rawdata.HandleGetLiftView(roDB)))

	// Movement table view
	mux.Handle("POST /view/data/movement", alertMiddleware(rawdata.HandlePostMovementView(roDB, wDB)))
	mux.Handle("PATCH /view/data/movement", alertMiddleware(rawdata.HandlePatchMovementView(wDB)))
	mux.Handle("PATCH /view/data/movement/{id}", alertMiddleware(rawdata.HandlePatchMovementView(wDB)))
	mux.Handle("DELETE /view/data/movement",
		alertMiddleware(rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteMovement)))
	mux.Handle("DELETE /view/data/movement/{id}",
		alertMiddleware(rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteMovement)))
	mux.Handle("GET /view/data/movement", alertMiddleware(rawdata.HandleGetMovementView(roDB)))

	// Muscle table view
	mux.Handle("POST /view/data/muscle", alertMiddleware(rawdata.HandlePostMuscleView(roDB, wDB)))
	mux.Handle("PATCH /view/data/muscle", alertMiddleware(rawdata.HandlePatchMuscleView(wDB)))
	mux.Handle("PATCH /view/data/muscle/{id}", alertMiddleware(rawdata.HandlePatchMuscleView(wDB)))
	mux.Handle("DELETE /view/data/muscle",
		alertMiddleware(rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteMuscle)))
	mux.Handle("DELETE /view/data/muscle/{id}",
		alertMiddleware(rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteMuscle)))
	mux.Handle("GET /view/data/muscle", alertMiddleware(rawdata.HandleGetMuscleView(roDB)))

	// Routine table view
	mux.Handle("POST /view/data/routine", alertMiddleware(rawdata.HandlePostRoutineView(roDB, wDB)))
	mux.Handle("PATCH /view/data/routine", alertMiddleware(rawdata.HandlePatchRoutineView(wDB)))
	mux.Handle("PATCH /view/data/routine/{id}", alertMiddleware(rawdata.HandlePatchRoutineView(wDB)))
	mux.Handle("DELETE /view/data/routine",
		alertMiddleware(rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteRoutine)))
	mux.Handle("DELETE /view/data/routine/{id}",
		alertMiddleware(rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteRoutine)))
	mux.Handle("GET /view/data/routine", alertMiddleware(rawdata.HandleGetRoutineView(roDB)))

	// Schedule list view
	mux.Handle("POST /view/data/schedule_list", alertMiddleware(rawdata.HandlePostScheduleListView(roDB, wDB)))
	mux.Handle("PATCH /view/data/schedule_list", alertMiddleware(rawdata.HandlePatchScheduleListView(wDB)))
	mux.Handle("PATCH /view/data/schedule_list/{id}", alertMiddleware(rawdata.HandlePatchScheduleListView(wDB)))
	mux.Handle("DELETE /view/data/schedule_list",
		alertMiddleware(rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteScheduleList)))
	mux.Handle("DELETE /view/data/schedule_list/{id}",
		alertMiddleware(rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteScheduleList)))
	mux.Handle("GET /view/data/schedule_list", alertMiddleware(rawdata.HandleGetScheduleListView(roDB)))

	// Side weight view
	mux.Handle("POST /view/data/side_weight", alertMiddleware(rawdata.HandlePostSideWeightView(roDB, wDB)))
	mux.Handle("PATCH /view/data/side_weight", alertMiddleware(rawdata.HandlePatchSideWeightView(wDB)))
	mux.Handle("PATCH /view/data/side_weight/{id}", alertMiddleware(rawdata.HandlePatchSideWeightView(wDB)))
	mux.Handle("DELETE /view/data/side_weight",
		alertMiddleware(rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteSideWeight)))
	mux.Handle("DELETE /view/data/side_weight/{id}",
		alertMiddleware(rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteSideWeight)))
	mux.Handle("GET /view/data/side_weight", alertMiddleware(rawdata.HandleGetSideWeightView(roDB)))

	// Template variable view
	mux.Handle("POST /view/data/template_variable", alertMiddleware(rawdata.HandlePostTemplateVariableView(roDB, wDB)))
	mux.Handle("PATCH /view/data/template_variable", alertMiddleware(rawdata.HandlePatchTemplateVariableView(wDB)))
	mux.Handle("PATCH /view/data/template_variable/{id}", alertMiddleware(rawdata.HandlePatchTemplateVariableView(wDB)))
	mux.Handle("DELETE /view/data/template_variable",
		alertMiddleware(rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteTemplateVariable)))
	mux.Handle("DELETE /view/data/template_variable/{id}",
		alertMiddleware(rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteTemplateVariable)))
	mux.Handle("GET /view/data/template_variable", alertMiddleware(rawdata.HandleGetTemplateVariableView(roDB)))

	// Workout view
	mux.Handle("POST /view/data/workout", alertMiddleware(rawdata.HandlePostWorkoutView(roDB, wDB)))
	mux.Handle("PATCH /view/data/workout", alertMiddleware(rawdata.HandlePatchWorkoutView(wDB)))
	mux.Handle("PATCH /view/data/workout/{id}", alertMiddleware(rawdata.HandlePatchWorkoutView(wDB)))
	mux.Handle("DELETE /view/data/workout",
		alertMiddleware(rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteWorkout)))
	mux.Handle("DELETE /view/data/workout/{id}",
		alertMiddleware(rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteWorkout)))
	mux.Handle("GET /view/data/workout", alertMiddleware(rawdata.HandleGetWorkoutView(roDB)))

	// Schedule view
	mux.Handle("POST /view/data/schedule", alertMiddleware(rawdata.HandlePostScheduleView(roDB, wDB)))
	mux.Handle("PATCH /view/data/schedule", alertMiddleware(rawdata.HandlePatchScheduleView(wDB)))
	mux.Handle("PATCH /view/data/schedule/{id}", alertMiddleware(rawdata.HandlePatchScheduleView(wDB)))
	mux.Handle("DELETE /view/data/schedule",
		alertMiddleware(rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteSchedule)))
	mux.Handle("DELETE /view/data/schedule/{id}",
		alertMiddleware(rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteSchedule)))
	mux.Handle("GET /view/data/schedule", alertMiddleware(rawdata.HandleGetScheduleView(roDB)))

	// Progress view
	mux.Handle("POST /view/data/progress", alertMiddleware(rawdata.HandlePostProgressView(roDB, wDB)))
	mux.Handle("PATCH /view/data/progress", alertMiddleware(rawdata.HandlePatchProgressView(wDB)))
	mux.Handle("PATCH /view/data/progress/{id}", alertMiddleware(rawdata.HandlePatchProgressView(wDB)))
	mux.Handle("DELETE /view/data/progress/{id}",
		alertMiddleware(rawdata.HandleDeleteTableRowViewRequest(
			wDB, (*workoutdb.Queries).RawDeleteProgress,
			func(r *http.Request) (*int64, error) {
				id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
				if err != nil {
					return nil, fmt.Errorf("failed to parse id: %w", err)
				}
				return &id, nil
			},
		)))
	mux.Handle("GET /view/data/progress", alertMiddleware(rawdata.HandleGetProgressView(roDB)))

	// Lift muscle mapping view
	mux.Handle("POST /view/data/lift_muscle_mapping", alertMiddleware(rawdata.HandlePostLiftMuscleView(roDB, wDB)))
	mux.Handle("PATCH /view/data/lift_muscle_mapping/{lift}/{muscle}/{movement}", alertMiddleware(rawdata.HandlePatchLiftMuscleView(wDB)))
	mux.Handle("DELETE /view/data/lift_muscle_mapping/{lift}/{muscle}/{movement}",
		alertMiddleware(rawdata.HandleDeleteTableRowViewRequest(
			wDB, (*workoutdb.Queries).RawDeleteLiftMuscle,
			func(r *http.Request) (*workoutdb.RawDeleteLiftMuscleParams, error) {
				return &workoutdb.RawDeleteLiftMuscleParams{
					Lift:     r.PathValue("lift"),
					Muscle:   r.PathValue("muscle"),
					Movement: r.PathValue("movement"),
				}, nil
			},
		)))
	mux.Handle("GET /view/data/lift_muscle_mapping", alertMiddleware(rawdata.HandleGetLiftMuscleView(roDB)))

	// Lift workout mapping view
	mux.Handle("POST /view/data/lift_workout_mapping", alertMiddleware(rawdata.HandlePostLiftWorkoutView(roDB, wDB)))
	mux.Handle("PATCH /view/data/lift_workout_mapping/{lift}/{workout}", alertMiddleware(rawdata.HandlePatchLiftWorkoutView(wDB)))
	mux.Handle("DELETE /view/data/lift_workout_mapping/{lift}/{workout}",
		alertMiddleware(rawdata.HandleDeleteTableRowViewRequest(
			wDB, (*workoutdb.Queries).RawDeleteLiftWorkout,
			func(r *http.Request) (*workoutdb.RawDeleteLiftWorkoutParams, error) {
				return &workoutdb.RawDeleteLiftWorkoutParams{
					Lift:    r.PathValue("lift"),
					Workout: r.PathValue("workout"),
				}, nil
			},
		)))
	mux.Handle("GET /view/data/lift_workout_mapping", alertMiddleware(rawdata.HandleGetLiftWorkoutView(roDB)))

	// Routine workout mapping view
	mux.Handle("POST /view/data/routine_workout_mapping", alertMiddleware(rawdata.HandlePostRoutineWorkoutView(roDB, wDB)))
	mux.Handle("PATCH /view/data/routine_workout_mapping/{routine}/{workout}", alertMiddleware(rawdata.HandlePatchRoutineWorkoutView(wDB)))
	mux.Handle("DELETE /view/data/routine_workout_mapping/{routine}/{workout}",
		alertMiddleware(rawdata.HandleDeleteTableRowViewRequest(
			wDB, (*workoutdb.Queries).RawDeleteRoutineWorkout,
			func(r *http.Request) (*workoutdb.RawDeleteRoutineWorkoutParams, error) {
				return &workoutdb.RawDeleteRoutineWorkoutParams{
					Routine: r.PathValue("routine"),
					Workout: r.PathValue("workout"),
				}, nil
			},
		)))
	mux.Handle("GET /view/data/routine_workout_mapping", alertMiddleware(rawdata.HandleGetRoutineWorkoutView(roDB)))

	// Subworkout mapping view
	mux.Handle("POST /view/data/subworkout", alertMiddleware(rawdata.HandlePostSubworkoutView(roDB, wDB)))
	mux.Handle("PATCH /view/data/subworkout/{subworkout}/{superworkout}", alertMiddleware(rawdata.HandlePatchSubworkoutView(wDB)))
	mux.Handle("DELETE /view/data/subworkout/{subworkout}/{superworkout}",
		alertMiddleware(rawdata.HandleDeleteTableRowViewRequest(
			wDB, (*workoutdb.Queries).RawDeleteSubworkout,
			func(r *http.Request) (*workoutdb.RawDeleteSubworkoutParams, error) {
				return &workoutdb.RawDeleteSubworkoutParams{
					Subworkout:   r.PathValue("subworkout"),
					Superworkout: r.PathValue("superworkout"),
				}, nil
			},
		)))
	mux.Handle("GET /view/data/subworkout", alertMiddleware(rawdata.HandleGetSubworkoutView(roDB)))

	// Lift group table view
	mux.Handle("POST /view/data/lift_group", alertMiddleware(rawdata.HandlePostLiftGroupView(roDB, wDB)))
	mux.Handle("PATCH /view/data/lift_group", alertMiddleware(rawdata.HandlePatchLiftGroupView(wDB)))
	mux.Handle("PATCH /view/data/lift_group/{id}", alertMiddleware(rawdata.HandlePatchLiftGroupView(wDB)))
	mux.Handle("DELETE /view/data/lift_group",
		alertMiddleware(rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteLiftGroup)))
	mux.Handle("DELETE /view/data/lift_group/{id}",
		alertMiddleware(rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteLiftGroup)))
	mux.Handle("GET /view/data/lift_group", alertMiddleware(rawdata.HandleGetLiftGroupView(roDB)))
}
