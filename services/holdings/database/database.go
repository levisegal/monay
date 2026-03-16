package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	_ "modernc.org/sqlite"
)

//go:embed sql/schema.sql
var schema string

func Open(ctx context.Context, dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	if _, err := db.ExecContext(ctx, schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	for _, col := range []string{
		"ALTER TABLE securities ADD COLUMN expense_ratio_bps integer",
		"ALTER TABLE securities ADD COLUMN fund_family text",
		"ALTER TABLE securities ADD COLUMN fund_category text",
		"ALTER TABLE lots ADD COLUMN estimated_basis integer not null default 0",
	} {
		db.ExecContext(ctx, col)
	}

	return db, nil
}
