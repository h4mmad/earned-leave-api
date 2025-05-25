package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InitPool(ctx context.Context, url string) *pgxpool.Pool {
	// Build pool config from the DSN.
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		log.Fatalf("parse dsn: %v", err) // fatal: can't even start
	}

	// Disable statement cache (fixes "already exists" with PgBouncer etc.).
	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	// OR: cfg.ConnConfig.BuildStatementCache = nil

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	return pool

}
