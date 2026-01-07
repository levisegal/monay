package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/levisegal/monay/services/holdings/database/sql/migrations"
)

func MigrateUp(ctx context.Context, connString string) error {
	db, err := sql.Open("pgx", connString)
	if err != nil {
		return fmt.Errorf("failed to open database for migrations: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	goose.SetLogger(goose.NopLogger())

	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations.Embed)
	if err != nil {
		return fmt.Errorf("failed to create migration provider: %w", err)
	}

	results, err := provider.Up(ctx)
	if err != nil {
		return fmt.Errorf("failed to apply database migrations: %w", err)
	}

	if len(results) == 0 {
		slog.Info("database migrations up-to-date")
		return nil
	}

	for _, r := range results {
		slog.Info("applied migration", "version", r.Source.Version, "name", r.Source.Path)
	}

	return nil
}

