package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/RyRose/uplog/internal/calendar"
	"github.com/RyRose/uplog/internal/service/index"
	"github.com/RyRose/uplog/internal/service/rawdata"
	"github.com/RyRose/uplog/internal/service/schedule"
	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
	"github.com/RyRose/uplog/internal/templates"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// TODO: Parameterize today's date for testing support.
func todaysDate() time.Time {
	return time.Now()
}

type errorResponseWriter struct {
	http.ResponseWriter
	requestCtx context.Context
	statusCode int
}

func (w *errorResponseWriter) Write(b []byte) (int, error) {
	if w.statusCode < 400 {
		return w.ResponseWriter.Write(b)
	}

	w.ResponseWriter.WriteHeader(http.StatusUnprocessableEntity)
	var s strings.Builder
	_, err := fmt.Fprint(&s, w.statusCode, ": ")
	if err != nil {
		return 0, fmt.Errorf("failed to write status code: %w", err)
	}
	n, _ := s.Write(b)
	return n, templates.Alert(s.String()).Render(w.requestCtx, w.ResponseWriter)
}

func (w *errorResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode

	if statusCode < 400 {
		w.ResponseWriter.WriteHeader(statusCode)
	}
}

func alertMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(&errorResponseWriter{ResponseWriter: w, requestCtx: r.Context()}, r)
	})
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

func handleIndexPage(tab, cssQuery string, calendarSrv *calendar.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		code := r.URL.Query().Get("code")
		if code != "" {
			// FIXME: Include error in the html response.
			if err := calendarSrv.Init(ctx, code, r.URL.Query().Get("state")); err != nil {
				slog.WarnContext(ctx, "failed to initialize calendar service", "error", err)
			}
		}
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

func handleRawPost[dataType, retType any](
	db *sql.DB,
	insert func(*workoutdb.Queries, context.Context, dataType) (retType, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		decoder := json.NewDecoder(r.Body)

		urlquery, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			slog.ErrorContext(ctx, "failed to parse url query", "error", err)
			http.Error(w, fmt.Sprintf("failed to parse url query: %v", err), http.StatusBadRequest)
		}

		var allData []dataType
		if urlquery.Has("bulk") {
			if err := decoder.Decode(&allData); err != nil {
				slog.ErrorContext(ctx, "failed to decode bulk request", "error", err)
				http.Error(w,
					fmt.Sprintf("failed to decode bulk request: %v", err), http.StatusBadRequest)
				return
			}
		} else {
			var data dataType
			if err := decoder.Decode(&data); err != nil {
				slog.ErrorContext(ctx, "failed to decode request", "error", err)
				http.Error(w,
					fmt.Sprintf("failed to decode request: %v", err), http.StatusBadRequest)
				return
			}
			allData = append(allData, data)
		}

		queries := workoutdb.New(db)
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			slog.ErrorContext(ctx, "failed to start transaction", "error", err)
			http.Error(w, fmt.Sprintf("failed to start transaction: %v", err), http.StatusBadRequest)
			return
		}
		defer tx.Rollback()
		queries = queries.WithTx(tx)

		for i, data := range allData {
			if _, err := insert(queries, ctx, data); err != nil {
				if urlquery.Has("continue") {
					slog.WarnContext(
						ctx, "failed to insert request",
						"index", i, "error", err, "data", data)
					continue
				}
				slog.ErrorContext(
					ctx, "failed to insert request", "index", i, "error", err, "data", data)
				http.Error(w,
					fmt.Sprintf("failed to insert request %d: %v", i, err), http.StatusBadRequest)
				return
			}
		}
		if err := tx.Commit(); err != nil {
			slog.ErrorContext(ctx, "failed to commit request", "error", err)
			http.Error(w, fmt.Sprintf("failed to commit request: %v", err), http.StatusBadRequest)
			return
		}
	}
}

func handleGetCalendarAuthURL(calendarService *calendar.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !calendarService.Initializable() || calendarService.Initialized() {
			w.WriteHeader(http.StatusOK)
			return
		}

		ctx := r.Context()
		url, err := calendarService.AuthCodeURL()
		if err != nil {
			http.Error(w, "failed to get auth code url", http.StatusInternalServerError)
			slog.ErrorContext(ctx, "failed to get auth code url", "error", err)
			return
		}
		if err := templates.AuthorizationURL(url).Render(ctx, w); err != nil {
			http.Error(w, "failed to write response", http.StatusInternalServerError)
			slog.ErrorContext(ctx, "failed to write response", "error", err)
		}
	}
}

func AddRoutes(
	_ context.Context,
	mux *http.ServeMux,
	wDB, roDB *sql.DB,
	calendarService *calendar.Service) {

	// Prometheus metrics.
	mux.Handle("/metrics", promhttp.Handler())

	// Swagger docs.
	// TODO: Only enable swagger docs in "dev" mode.
	// TODO: Make swagger URL configurable and use better default fqdn based on hostname when serving locally.
	mux.Handle("GET /docs/", http.FileServer(http.Dir(".")))
	mux.Handle("GET /swagger/", httpSwagger.Handler(httpSwagger.URL("http://localhost:8080/docs/swagger.json")))

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
	mux.Handle("GET /view/tabs/main", alertMiddleware(index.HandleMainTab(roDB)))
	mux.Handle("GET /view/calendarauthurl", alertMiddleware(handleGetCalendarAuthURL(calendarService)))
	mux.Handle("GET /view/liftgroups", alertMiddleware(index.HandleGetLiftGroupListView(roDB)))

	// Progress table
	mux.Handle("GET /view/progresstable", alertMiddleware(index.HandleGetProgressTable(roDB)))
	mux.Handle("DELETE /view/progresstablerow/{id}", alertMiddleware(index.HandleDeleteProgress(wDB)))
	mux.Handle("POST /view/progresstablerow", alertMiddleware(index.HandleCreateProgress(wDB)))

	// Routine table.
	mux.Handle("GET /view/routinetable", alertMiddleware(index.HandleGetRoutineTable(roDB)))

	// Progress form.
	mux.Handle("GET /view/liftselect", alertMiddleware(index.HandleGetLiftSelect(roDB)))
	mux.Handle("GET /view/sideweightselect", alertMiddleware(index.HandleGetSideWeightSelect(roDB)))
	mux.Handle("GET /view/progressform", alertMiddleware(index.HandleGetProgressForm()))
	mux.Handle("POST /view/progressform", alertMiddleware(index.HandleCreateProgressForm(roDB)))

	// Schedule view.
	mux.Handle("GET /view/tabs/schedule", alertMiddleware(schedule.HandleGetScheduleTable(roDB)))

	// Schedule table.
	mux.Handle("DELETE /view/schedule/{date}", alertMiddleware(schedule.HandleDeleteSchedule(wDB, calendarService)))
	mux.Handle("PATCH /view/schedule/{date}", alertMiddleware(schedule.HandlePatchScheduleTableRow(wDB, calendarService)))
	mux.Handle("POST /view/scheduleappend", alertMiddleware(schedule.HandlePostScheduleTableRow(wDB, calendarService)))
	mux.Handle("POST /view/scheduletablerows", alertMiddleware(schedule.HandlePostScheduleTableRows(wDB, calendarService)))

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
