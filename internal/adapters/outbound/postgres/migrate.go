package postgres

import (
	"context"
	"embed"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// RunMigrations applies embedded SQL files in lexical order, once per filename.
func RunMigrations(ctx context.Context, dbPool *pgxpool.Pool) error {
	_, execErr := dbPool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`)
	if execErr != nil {
		return fmt.Errorf("create schema_migrations: %w", execErr)
	}

	dirEntries, readDirErr := migrationFiles.ReadDir("migrations")
	if readDirErr != nil {
		return fmt.Errorf("read migrations dir: %w", readDirErr)
	}
	sort.Slice(dirEntries, func(left, right int) bool {
		return dirEntries[left].Name() < dirEntries[right].Name()
	})

	for _, entry := range dirEntries {
		if entry.IsDir() {
			continue
		}
		version := entry.Name()
		var alreadyApplied bool
		scanErr := dbPool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`, version).Scan(&alreadyApplied)
		if scanErr != nil {
			return fmt.Errorf("check migration %s: %w", version, scanErr)
		}
		if alreadyApplied {
			continue
		}

		migrationSQL, readErr := migrationFiles.ReadFile("migrations/" + version)
		if readErr != nil {
			return fmt.Errorf("read migration file %s: %w", version, readErr)
		}

		transaction, beginErr := dbPool.Begin(ctx)
		if beginErr != nil {
			return beginErr
		}
		if _, runErr := transaction.Exec(ctx, string(migrationSQL)); runErr != nil {
			_ = transaction.Rollback(ctx)
			return fmt.Errorf("run migration %s: %w", version, runErr)
		}
		if _, insertErr := transaction.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version); insertErr != nil {
			_ = transaction.Rollback(ctx)
			return fmt.Errorf("record migration %s: %w", version, insertErr)
		}
		if commitErr := transaction.Commit(ctx); commitErr != nil {
			return commitErr
		}
	}
	return nil
}
