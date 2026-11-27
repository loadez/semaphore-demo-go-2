// Package store wraps postgres and redis access for the integration tests.
package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
)

type Store struct {
	DB  *sql.DB
	RDB *redis.Client
}

func Open(pgDSN, redisAddr string) (*Store, error) {
	db, err := sql.Open("postgres", pgDSN)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(8)
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	return &Store{DB: db, RDB: rdb}, nil
}

func (s *Store) Close() { s.DB.Close(); s.RDB.Close() }

func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.DB.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS events (
		id BIGSERIAL PRIMARY KEY,
		kind TEXT NOT NULL,
		payload TEXT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now())`)
	return err
}

// Record inserts an event and caches its count per kind in redis.
func (s *Store) Record(ctx context.Context, kind, payload string) (int64, error) {
	var id int64
	err := s.DB.QueryRowContext(ctx,
		`INSERT INTO events(kind, payload) VALUES($1,$2) RETURNING id`, kind, payload).Scan(&id)
	if err != nil {
		return 0, err
	}
	if err := s.RDB.Incr(ctx, "count:"+kind).Err(); err != nil {
		return id, fmt.Errorf("redis incr: %w", err)
	}
	return id, nil
}

func (s *Store) CountDB(ctx context.Context, kind string) (int64, error) {
	var n int64
	err := s.DB.QueryRowContext(ctx, `SELECT count(*) FROM events WHERE kind=$1`, kind).Scan(&n)
	return n, err
}

func (s *Store) CountCache(ctx context.Context, kind string) (int64, error) {
	return s.RDB.Get(ctx, "count:"+kind).Int64()
}

func (s *Store) Cache(ctx context.Context, k, v string, ttl time.Duration) error {
	return s.RDB.Set(ctx, k, v, ttl).Err()
}
