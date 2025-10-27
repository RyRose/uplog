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

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	if err := service.Run(ctx, "./config/main.lua"); err != nil {
		log.Fatalf("server failed to run: %v", err)
	}
	slog.InfoContext(ctx, "server exited gracefully")
}
