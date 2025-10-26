package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/RyRose/uplog/internal/config"
	"github.com/RyRose/uplog/internal/service/index"
	"github.com/RyRose/uplog/internal/service/mux"
	"github.com/RyRose/uplog/internal/service/rawdata"
	"github.com/RyRose/uplog/internal/service/schedule"
	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpswagger "github.com/swaggo/http-swagger/v2"
)

func AddRoutes(
	_ context.Context,
	rawMux *http.ServeMux,
	cfg *config.Data,
	state *State) {

	ts := fmt.Sprint(time.Now().Unix())
	traceMux := &mux.Trace{Mux: rawMux}
	webMux := &mux.Web{Mux: traceMux}
	roDB := state.ReadonlyDB
	wDB := state.WriteDB

	// Prometheus metrics.
	traceMux.Handle("/metrics", promhttp.Handler())

	// Swagger docs.
	// TODO: Only enable swagger docs in "dev" mode.
	traceMux.Handle("GET /docs/", http.FileServer(http.Dir(".")))
	traceMux.Handle("GET /swagger/", httpswagger.Handler(httpswagger.URL(cfg.SwaggerURL)))

	// Vendor and non-vendor static assets.
	traceMux.Handle("GET /web/static/", http.FileServer(http.Dir(".")))
	traceMux.Handle("GET /web/vendor/", http.FileServer(http.Dir(".")))

	// Index pages.
	// TODO: Use hash of css for cache busting instead of date.
	traceMux.HandleFunc("GET /{$}", index.HandleIndexPage("main", ts))
	traceMux.HandleFunc("GET /schedule/{$}", index.HandleIndexPage("schedule", ts))
	traceMux.HandleFunc("GET /data/{$}", index.HandleIndexPage("data", ts))
	traceMux.HandleFunc("GET /data/{tabX}/{tabY}", index.HandleIndexPage("data", ts))

	// Main view.
	webMux.Handle("GET /view/tabs/main", index.HandleMainTab(roDB))
	webMux.Handle("GET /view/liftgroups", index.HandleGetLiftGroupListView(roDB))

	// Progress table
	webMux.Handle("GET /view/progresstable", index.HandleGetProgressTable(roDB))
	webMux.Handle("DELETE /view/progresstablerow/{id}", index.HandleDeleteProgress(wDB))
	webMux.Handle("POST /view/progresstablerow", index.HandleCreateProgress(wDB))

	// Routine table.
	webMux.Handle("GET /view/routinetable", index.HandleGetRoutineTable(roDB))

	// Progress form.
	webMux.Handle("GET /view/liftselect", index.HandleGetLiftSelect(roDB))
	webMux.Handle("GET /view/sideweightselect", index.HandleGetSideWeightSelect(roDB))
	webMux.Handle("GET /view/progressform", index.HandleGetProgressForm())
	webMux.Handle("POST /view/progressform", index.HandleCreateProgressForm(roDB))

	// Schedule view.
	webMux.Handle("GET /view/tabs/schedule", schedule.HandleGetScheduleTable(roDB))

	// Schedule table.
	webMux.Handle("DELETE /view/schedule/{date}", schedule.HandleDeleteSchedule(wDB))
	webMux.Handle("PATCH /view/schedule/{date}", schedule.HandlePatchScheduleTableRow(wDB))
	webMux.Handle("POST /view/scheduleappend", schedule.HandlePostScheduleTableRow(wDB))
	webMux.Handle("POST /view/scheduletablerows", schedule.HandlePostScheduleTableRows(wDB))

	// Data view
	webMux.Handle("GET /view/tabs/data/{$}", index.HandleGetDataTabView())
	webMux.Handle("GET /view/tabs/data/{tabX}/{tabY}", index.HandleGetDataTabView())

	// Lift table view
	webMux.Handle("POST /view/data/lift", rawdata.HandlePostLiftView(roDB, wDB))
	webMux.Handle("PATCH /view/data/lift", rawdata.HandlePatchLiftView(wDB))
	webMux.Handle("PATCH /view/data/lift/{id}", rawdata.HandlePatchLiftView(wDB))
	webMux.Handle("DELETE /view/data/lift", rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteLift))
	webMux.Handle("DELETE /view/data/lift/{id}", rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteLift))
	webMux.Handle("GET /view/data/lift", rawdata.HandleGetLiftView(roDB))

	// Movement table view
	webMux.Handle("POST /view/data/movement", rawdata.HandlePostMovementView(roDB, wDB))
	webMux.Handle("PATCH /view/data/movement", rawdata.HandlePatchMovementView(wDB))
	webMux.Handle("PATCH /view/data/movement/{id}", rawdata.HandlePatchMovementView(wDB))
	webMux.Handle("DELETE /view/data/movement", rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteMovement))
	webMux.Handle("DELETE /view/data/movement/{id}", rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteMovement))
	webMux.Handle("GET /view/data/movement", rawdata.HandleGetMovementView(roDB))

	// Muscle table view
	webMux.Handle("POST /view/data/muscle", rawdata.HandlePostMuscleView(roDB, wDB))
	webMux.Handle("PATCH /view/data/muscle", rawdata.HandlePatchMuscleView(wDB))
	webMux.Handle("PATCH /view/data/muscle/{id}", rawdata.HandlePatchMuscleView(wDB))
	webMux.Handle("DELETE /view/data/muscle", rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteMuscle))
	webMux.Handle("DELETE /view/data/muscle/{id}", rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteMuscle))
	webMux.Handle("GET /view/data/muscle", rawdata.HandleGetMuscleView(roDB))

	// Routine table view
	webMux.Handle("POST /view/data/routine", rawdata.HandlePostRoutineView(roDB, wDB))
	webMux.Handle("PATCH /view/data/routine", rawdata.HandlePatchRoutineView(wDB))
	webMux.Handle("PATCH /view/data/routine/{id}", rawdata.HandlePatchRoutineView(wDB))
	webMux.Handle("DELETE /view/data/routine", rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteRoutine))
	webMux.Handle("DELETE /view/data/routine/{id}", rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteRoutine))
	webMux.Handle("GET /view/data/routine", rawdata.HandleGetRoutineView(roDB))

	// Schedule list view
	webMux.Handle("POST /view/data/schedule_list", rawdata.HandlePostScheduleListView(roDB, wDB))
	webMux.Handle("PATCH /view/data/schedule_list", rawdata.HandlePatchScheduleListView(wDB))
	webMux.Handle("PATCH /view/data/schedule_list/{id}", rawdata.HandlePatchScheduleListView(wDB))
	webMux.Handle("DELETE /view/data/schedule_list", rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteScheduleList))
	webMux.Handle("DELETE /view/data/schedule_list/{id}", rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteScheduleList))
	webMux.Handle("GET /view/data/schedule_list", rawdata.HandleGetScheduleListView(roDB))

	// Side weight view
	webMux.Handle("POST /view/data/side_weight", rawdata.HandlePostSideWeightView(roDB, wDB))
	webMux.Handle("PATCH /view/data/side_weight", rawdata.HandlePatchSideWeightView(wDB))
	webMux.Handle("PATCH /view/data/side_weight/{id}", rawdata.HandlePatchSideWeightView(wDB))
	webMux.Handle("DELETE /view/data/side_weight", rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteSideWeight))
	webMux.Handle("DELETE /view/data/side_weight/{id}", rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteSideWeight))
	webMux.Handle("GET /view/data/side_weight", rawdata.HandleGetSideWeightView(roDB))

	// Template variable view
	webMux.Handle("POST /view/data/template_variable", rawdata.HandlePostTemplateVariableView(roDB, wDB))
	webMux.Handle("PATCH /view/data/template_variable", rawdata.HandlePatchTemplateVariableView(wDB))
	webMux.Handle("PATCH /view/data/template_variable/{id}", rawdata.HandlePatchTemplateVariableView(wDB))
	webMux.Handle("DELETE /view/data/template_variable", rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteTemplateVariable))
	webMux.Handle("DELETE /view/data/template_variable/{id}", rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteTemplateVariable))
	webMux.Handle("GET /view/data/template_variable", rawdata.HandleGetTemplateVariableView(roDB))

	// Workout view
	webMux.Handle("POST /view/data/workout", rawdata.HandlePostWorkoutView(roDB, wDB))
	webMux.Handle("PATCH /view/data/workout", rawdata.HandlePatchWorkoutView(wDB))
	webMux.Handle("PATCH /view/data/workout/{id}", rawdata.HandlePatchWorkoutView(wDB))
	webMux.Handle("DELETE /view/data/workout", rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteWorkout))
	webMux.Handle("DELETE /view/data/workout/{id}", rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteWorkout))
	webMux.Handle("GET /view/data/workout", rawdata.HandleGetWorkoutView(roDB))

	// Schedule view
	webMux.Handle("POST /view/data/schedule", rawdata.HandlePostScheduleView(roDB, wDB))
	webMux.Handle("PATCH /view/data/schedule", rawdata.HandlePatchScheduleView(wDB))
	webMux.Handle("PATCH /view/data/schedule/{id}", rawdata.HandlePatchScheduleView(wDB))
	webMux.Handle("DELETE /view/data/schedule", rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteSchedule))
	webMux.Handle("DELETE /view/data/schedule/{id}", rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteSchedule))
	webMux.Handle("GET /view/data/schedule", rawdata.HandleGetScheduleView(roDB))

	// Progress view
	webMux.Handle("POST /view/data/progress", rawdata.HandlePostProgressView(roDB, wDB))
	webMux.Handle("PATCH /view/data/progress", rawdata.HandlePatchProgressView(wDB))
	webMux.Handle("PATCH /view/data/progress/{id}", rawdata.HandlePatchProgressView(wDB))
	webMux.Handle("DELETE /view/data/progress/{id}",
		rawdata.HandleDeleteTableRowViewRequest(
			wDB, (*workoutdb.Queries).RawDeleteProgress,
			func(r *http.Request) (*int64, error) {
				id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
				if err != nil {
					return nil, fmt.Errorf("failed to parse id: %w", err)
				}
				return &id, nil
			},
		))
	webMux.Handle("GET /view/data/progress", rawdata.HandleGetProgressView(roDB))

	// Lift muscle mapping view
	webMux.Handle("POST /view/data/lift_muscle_mapping", rawdata.HandlePostLiftMuscleView(roDB, wDB))
	webMux.Handle("PATCH /view/data/lift_muscle_mapping/{lift}/{muscle}/{movement}", rawdata.HandlePatchLiftMuscleView(wDB))
	webMux.Handle("DELETE /view/data/lift_muscle_mapping/{lift}/{muscle}/{movement}",
		rawdata.HandleDeleteTableRowViewRequest(
			wDB, (*workoutdb.Queries).RawDeleteLiftMuscle,
			func(r *http.Request) (*workoutdb.RawDeleteLiftMuscleParams, error) {
				return &workoutdb.RawDeleteLiftMuscleParams{
					Lift:     r.PathValue("lift"),
					Muscle:   r.PathValue("muscle"),
					Movement: r.PathValue("movement"),
				}, nil
			},
		))
	webMux.Handle("GET /view/data/lift_muscle_mapping", rawdata.HandleGetLiftMuscleView(roDB))

	// Lift workout mapping view
	webMux.Handle("POST /view/data/lift_workout_mapping", rawdata.HandlePostLiftWorkoutView(roDB, wDB))
	webMux.Handle("PATCH /view/data/lift_workout_mapping/{lift}/{workout}", rawdata.HandlePatchLiftWorkoutView(wDB))
	webMux.Handle("DELETE /view/data/lift_workout_mapping/{lift}/{workout}",
		rawdata.HandleDeleteTableRowViewRequest(
			wDB, (*workoutdb.Queries).RawDeleteLiftWorkout,
			func(r *http.Request) (*workoutdb.RawDeleteLiftWorkoutParams, error) {
				return &workoutdb.RawDeleteLiftWorkoutParams{
					Lift:    r.PathValue("lift"),
					Workout: r.PathValue("workout"),
				}, nil
			},
		))
	webMux.Handle("GET /view/data/lift_workout_mapping", rawdata.HandleGetLiftWorkoutView(roDB))

	// Routine workout mapping view
	webMux.Handle("POST /view/data/routine_workout_mapping", rawdata.HandlePostRoutineWorkoutView(roDB, wDB))
	webMux.Handle("PATCH /view/data/routine_workout_mapping/{routine}/{workout}", rawdata.HandlePatchRoutineWorkoutView(wDB))
	webMux.Handle("DELETE /view/data/routine_workout_mapping/{routine}/{workout}",
		rawdata.HandleDeleteTableRowViewRequest(
			wDB, (*workoutdb.Queries).RawDeleteRoutineWorkout,
			func(r *http.Request) (*workoutdb.RawDeleteRoutineWorkoutParams, error) {
				return &workoutdb.RawDeleteRoutineWorkoutParams{
					Routine: r.PathValue("routine"),
					Workout: r.PathValue("workout"),
				}, nil
			},
		))
	webMux.Handle("GET /view/data/routine_workout_mapping", rawdata.HandleGetRoutineWorkoutView(roDB))

	// Subworkout mapping view
	webMux.Handle("POST /view/data/subworkout", rawdata.HandlePostSubworkoutView(roDB, wDB))
	webMux.Handle("PATCH /view/data/subworkout/{subworkout}/{superworkout}", rawdata.HandlePatchSubworkoutView(wDB))
	webMux.Handle("DELETE /view/data/subworkout/{subworkout}/{superworkout}",
		rawdata.HandleDeleteTableRowViewRequest(
			wDB, (*workoutdb.Queries).RawDeleteSubworkout,
			func(r *http.Request) (*workoutdb.RawDeleteSubworkoutParams, error) {
				return &workoutdb.RawDeleteSubworkoutParams{
					Subworkout:   r.PathValue("subworkout"),
					Superworkout: r.PathValue("superworkout"),
				}, nil
			},
		))
	webMux.Handle("GET /view/data/subworkout", rawdata.HandleGetSubworkoutView(roDB))

	// Lift group table view
	webMux.Handle("POST /view/data/lift_group", rawdata.HandlePostLiftGroupView(roDB, wDB))
	webMux.Handle("PATCH /view/data/lift_group", rawdata.HandlePatchLiftGroupView(wDB))
	webMux.Handle("PATCH /view/data/lift_group/{id}", rawdata.HandlePatchLiftGroupView(wDB))
	webMux.Handle("DELETE /view/data/lift_group",
		rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteLiftGroup))
	webMux.Handle("DELETE /view/data/lift_group/{id}",
		rawdata.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteLiftGroup))
	webMux.Handle("GET /view/data/lift_group", rawdata.HandleGetLiftGroupView(roDB))
}
