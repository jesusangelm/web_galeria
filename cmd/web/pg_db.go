package main

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Creates the Connection Pool to the DB and return it
func openDB(cfg config) (*pgxpool.Pool, error) {
	poolcfg, err := pgxpool.ParseConfig(cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	poolcfg.MaxConns = int32(cfg.db.maxConns)

	dbPool, err := pgxpool.NewWithConfig(context.Background(), poolcfg)
	if err != nil {
		return nil, err
	}

	err = dbPool.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	return dbPool, nil
}
