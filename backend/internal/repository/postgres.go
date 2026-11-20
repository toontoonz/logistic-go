// Package repository talks to PostgreSQL. SQL stays here so services
// can be tested without a database.
package repository

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/toontoonz/logistic-go/backend/internal/boot"
)

// NewPool dials Postgres, retrying for about a minute. depends_on can
// say the container is healthy before this process can open a TCP
// connection, and the first dials fail.
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	cfg.MaxConns = 10

	var pool *pgxpool.Pool
	err = boot.UntilReady(ctx, "postgres", 40, time.Second, func(pingCtx context.Context) error {
		next, err := pgxpool.NewWithConfig(pingCtx, cfg)
		if err != nil {
			return err
		}
		if err := next.Ping(pingCtx); err != nil {
			next.Close()
			return err
		}
		pool = next
		return nil
	})
	if err != nil {
		return nil, err
	}
	return pool, nil
}

// Migrate applies *.sql files in lexical order, once each.
// schema_migrations is the record of what already ran, so a restart
// does not try to recreate tables.
func Migrate(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`)
	if err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)

	for _, name := range names {
		var applied bool
		err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`, name).Scan(&applied)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}
		if applied {
			continue
		}

		body, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if err := execScript(ctx, pool, string(body)); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
		if _, err := pool.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, name); err != nil {
			return fmt.Errorf("record migration %s: %w", name, err)
		}
	}
	return nil
}

// execScript uses the simple query protocol, which accepts several
// SQL statements in one string. pool.Exec uses the extended protocol
// and rejects that.
func execScript(ctx context.Context, pool *pgxpool.Pool, sql string) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	// ReadAll returns the first statement error, if any.
	_, err = conn.Conn().PgConn().Exec(ctx, sql).ReadAll()
	return err
}
