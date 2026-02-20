package store

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// Compile-time interface check.
var _ Store = (*SQLiteStore)(nil)

// SQLiteStore is a persistent Store backed by SQLite.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore opens (or creates) a SQLite database at the given path and
// initialises the schema. Use ":memory:" for an in-memory SQLite database.
func NewSQLiteStore(dsn string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("erl/store: open sqlite: %w", err)
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS erl_counters (
			key            TEXT PRIMARY KEY,
			count          INTEGER NOT NULL DEFAULT 0,
			bucket_key     TEXT NOT NULL DEFAULT '',
			window_seconds INTEGER NOT NULL DEFAULT 0
		)
	`); err != nil {
		db.Close()
		return nil, fmt.Errorf("erl/store: create table: %w", err)
	}

	return &SQLiteStore{db: db}, nil
}

// Increment atomically adds one to the counter for key in the current window bucket.
// If the bucket has rolled over, the counter is reset before incrementing.
func (s *SQLiteStore) Increment(ctx context.Context, key string, w Window) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var count int64
	var bucketKey string

	err = tx.QueryRowContext(ctx,
		`SELECT count, bucket_key FROM erl_counters WHERE key = ?`, key,
	).Scan(&count, &bucketKey)

	if err == sql.ErrNoRows {
		// New key, insert.
		_, err = tx.ExecContext(ctx,
			`INSERT INTO erl_counters (key, count, bucket_key, window_seconds) VALUES (?, 1, ?, ?)`,
			key, w.BucketKey, int64(w.Duration.Seconds()),
		)
		if err != nil {
			return 0, err
		}
		return 1, tx.Commit()
	}
	if err != nil {
		return 0, err
	}

	if bucketKey != w.BucketKey {
		// Window rolled over, reset.
		count = 0
	}

	count++
	_, err = tx.ExecContext(ctx,
		`UPDATE erl_counters SET count = ?, bucket_key = ?, window_seconds = ? WHERE key = ?`,
		count, w.BucketKey, int64(w.Duration.Seconds()), key,
	)
	if err != nil {
		return 0, err
	}

	return count, tx.Commit()
}

// Get returns the current counter value for key in the active window bucket.
func (s *SQLiteStore) Get(ctx context.Context, key string, w Window) (int64, error) {
	var count int64
	var bucketKey string

	err := s.db.QueryRowContext(ctx,
		`SELECT count, bucket_key FROM erl_counters WHERE key = ?`, key,
	).Scan(&count, &bucketKey)

	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	if bucketKey != w.BucketKey {
		return 0, nil
	}

	return count, nil
}

// Reset removes the counter for the given key.
func (s *SQLiteStore) Reset(ctx context.Context, key string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM erl_counters WHERE key = ?`, key)
	return err
}

// Close closes the underlying SQLite database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
