// goose is a custom migration tool for Uplog.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/RyRose/uplog/internal/config"
	"github.com/RyRose/uplog/internal/sqlc"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

func run(ctx context.Context, args []string) error {
	command := args[0]
	arguments := []string{}
	if len(args) > 1 {
		arguments = append(arguments, args[1:]...)
	}

	cfg, err := config.Load(ctx, "./config/main.lua")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	dsn := fmt.Sprintf("file:%s?mode=rwc&_journal_mode=WAL&_txlock=immediate", url.QueryEscape(cfg.DatabasePath))
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return fmt.Errorf("failed to open %v: %w", dsn, err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()

	goose.SetBaseFS(sqlc.EmbedMigrations)
	return goose.RunContext(ctx, command, db, "migrations", arguments...)
}

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("Usage: goose [OPTIONS] COMMAND [COMMAND_ARGS]")
		return
	}
	ctx := context.Background()
	if err := run(ctx, args); err != nil {
		log.Fatalf("goose: %v", err)
	}
}
