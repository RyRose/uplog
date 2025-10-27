package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"

	"github.com/RyRose/uplog/internal/service"

	_ "github.com/mattn/go-sqlite3"
)

// @title				Uplog API
// @version			1.0
// @description		A workout tracking and management system with progress logging, routine management, and workout scheduling.
// @description		This API provides endpoints for managing workouts, exercises, progress tracking, and muscle group mappings.
//
// @contact.name		RyRose
// @contact.url		https://github.com/RyRose/uplog
//
// @license.name		MIT
// @license.url		https://github.com/RyRose/uplog/blob/main/LICENSE
//
// @BasePath			/
//
// @schemes			http https
//
// @tag.name			index
// @tag.description	Main application pages and views
//
// @tag.name			rawdata
// @tag.description	CRUD operations for raw data entities (lifts, workouts, progress, etc.)
func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	if err := service.Run(ctx, "./config/main.lua"); err != nil {
		log.Fatalf("server failed to run: %v", err)
	}
	slog.InfoContext(ctx, "server exited gracefully")
}
