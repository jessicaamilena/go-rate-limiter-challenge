package limiter

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(dsn string) (*PostgresStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS rate_limits (k TEXT PRIMARY KEY, count INT NOT NULL, expires_at TIMESTAMP)`); err != nil {
		return nil, fmt.Errorf("create table rate_limits: %w", err)
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS bans (k TEXT PRIMARY KEY, expires_at TIMESTAMP)`); err != nil {
		return nil, fmt.Errorf("create table bans: %w", err)
	}
	return &PostgresStorage{db: db}, nil
}

func (p *PostgresStorage) Increment(ctx context.Context, key string, window time.Duration) (int, error) {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	var count int
	var expiresAt sql.NullTime
	err = tx.QueryRowContext(ctx, `SELECT count, expires_at FROM rate_limits WHERE k = $1 FOR UPDATE`, key).Scan(&count, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		count = 1
		expiresAt := time.Now().Add(window)
		_, err = tx.ExecContext(ctx, `INSERT INTO rate_limits (k, count, expires_at) VALUES ($1, $2, $3)`, key, count, expiresAt)
		return count, err
	}
	if err != nil {
		return 0, err
	}

	now := time.Now()
	if !expiresAt.Valid || expiresAt.Time.Before(now) {
		count = 1
	} else {
		count++
	}
	expires := now.Add(window)
	_, err = tx.ExecContext(ctx, `UPDATE rate_limits SET count=$1, expires_at=$2 WHERE k=$3`, count, expires, key)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (p *PostgresStorage) SetBan(ctx context.Context, key string, duration time.Duration) error {
	expiresAt := time.Now().Add(duration)
	_, err := p.db.ExecContext(ctx, `INSERT INTO bans (k, expires_at) VALUES ($1, $2) ON CONFLICT (k) DO UPDATE SET expires_at=EXCLUDED.expires_at`, key, expiresAt)
	return err
}

func (p *PostgresStorage) IsBanned(ctx context.Context, key string) (bool, error) {
	var expiresAt time.Time
	err := p.db.QueryRowContext(ctx, `SELECT expires_at FROM bans WHERE k = $1`, key).Scan(&expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if expiresAt.After(time.Now()) {
		return true, nil
	}
	return false, nil
}

func (p *PostgresStorage) Close() error {
	return p.db.Close()
}
