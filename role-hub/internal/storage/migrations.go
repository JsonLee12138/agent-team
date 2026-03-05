package storage

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

//go:embed migrations/postgres/*.sql migrations/sqlite/*.sql
var migrationFS embed.FS

func ApplyMigrations(ctx context.Context, db *sql.DB, dialect string) error {
	if err := ensureSchemaMigrations(ctx, db); err != nil {
		return err
	}

	files, err := readMigrationFiles(dialect)
	if err != nil {
		return err
	}
	for _, file := range files {
		applied, err := isMigrationApplied(ctx, db, dialect, file.version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		if err := applyMigration(ctx, db, dialect, file); err != nil {
			return err
		}
	}
	return nil
}

type migrationFile struct {
	version string
	path    string
	name    string
}

func readMigrationFiles(dialect string) ([]migrationFile, error) {
	base := "migrations/" + dialect
	entries, err := fs.ReadDir(migrationFS, base)
	if err != nil {
		return nil, fmt.Errorf("read migrations: %w", err)
	}
	files := make([]migrationFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		version := strings.SplitN(name, "_", 2)[0]
		files = append(files, migrationFile{
			version: version,
			path:    base + "/" + name,
			name:    name,
		})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].name < files[j].name })
	return files, nil
}

func ensureSchemaMigrations(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
  version TEXT PRIMARY KEY,
  applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);`)
	return err
}

func isMigrationApplied(ctx context.Context, db *sql.DB, dialect, version string) (bool, error) {
	placeholder := "$1"
	if dialect == "sqlite" {
		placeholder = "?"
	}
	row := db.QueryRowContext(ctx, "SELECT 1 FROM schema_migrations WHERE version = "+placeholder, version)
	var dummy int
	if scanErr := row.Scan(&dummy); scanErr == nil {
		return true, nil
	} else if scanErr == sql.ErrNoRows {
		return false, nil
	} else {
		return false, scanErr
	}
}

func applyMigration(ctx context.Context, db *sql.DB, dialect string, file migrationFile) error {
	content, err := fs.ReadFile(migrationFS, file.path)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", file.name, err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, string(content)); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("apply migration %s: %w", file.name, err)
	}
	placeholder := "$1"
	if dialect == "sqlite" {
		placeholder = "?"
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES ("+placeholder+")", file.version); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("record migration %s: %w", file.name, err)
	}
	return tx.Commit()
}
