package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/RyRose/uplog/internal/config"
	"github.com/RyRose/uplog/internal/service"

	_ "github.com/mattn/go-sqlite3"
)

func run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(
		ctx, os.Interrupt,
		syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	defer cancel()

	cfg, err := config.Load(ctx, "./config/main.lua")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	state, err := service.NewState(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create service state: %w", err)
	}
	defer func(ctx context.Context) {
		if err := state.Close(); err != nil {
			slog.WarnContext(ctx, "failed to close service state", "error", err)
		}
	}(ctx)

	srv := &http.Server{
		Addr:    net.JoinHostPort("", cfg.Port),
		Handler: service.NewServer(ctx, cfg, state),
	}
	go func(ctx context.Context) {
		<-ctx.Done()
		slog.InfoContext(ctx, "server context done, shutting down server")
		if err := srv.Shutdown(ctx); err != nil {
			slog.WarnContext(ctx, "failed to shutdown server", "error", err)
		}
	}(ctx)

	slog.InfoContext(ctx, "start listening", "addr", srv.Addr, "cfg", cfg)
	err = srv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return fmt.Errorf("failed to listen and serve: %w", srv.ListenAndServe())
}

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		log.Fatalf("server failed to run: %v", err)
	}
	slog.InfoContext(ctx, "server exited gracefully")
}
