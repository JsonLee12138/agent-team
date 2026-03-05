package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

func Open(ctx context.Context, dialect, dsn string) (*sql.DB, error) {
	switch dialect {
	case "postgres":
		// pgx uses "pgx" driver name.
		return openWithDriver(ctx, "pgx", dsn)
	case "sqlite":
		return openWithDriver(ctx, "sqlite", dsn)
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", dialect)
	}
}

func openWithDriver(ctx context.Context, driver, dsn string) (*sql.DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

type PoolOptions struct {
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
	MaxIdleTime  time.Duration
}

func ConfigurePool(db *sql.DB, opts PoolOptions) {
	if opts.MaxOpenConns > 0 {
		db.SetMaxOpenConns(opts.MaxOpenConns)
	}
	if opts.MaxIdleConns >= 0 {
		db.SetMaxIdleConns(opts.MaxIdleConns)
	}
	if opts.MaxLifetime > 0 {
		db.SetConnMaxLifetime(opts.MaxLifetime)
	}
	if opts.MaxIdleTime > 0 {
		db.SetConnMaxIdleTime(opts.MaxIdleTime)
	}
}
