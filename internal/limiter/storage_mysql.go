package limiter

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type MySQLStorage struct {
	db *sql.DB
}

func NewMySQLStorage(dsn string) (*MySQLStorage, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed connecting to MySQL: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed pinging MySQL: %w", err)
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS rate_limits (k VARCHAR(255) PRIMARY KEY, count INT NOT NULL, expires_at TIMESTAMP)`); err != nil {
		return nil, fmt.Errorf("create table rate_limits: %w", err)
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS bans (k VARCHAR(255) PRIMARY KEY, expires_at TIMESTAMP)`); err != nil {
		return nil, fmt.Errorf("create table bans: %w", err)
	}
	return &MySQLStorage{db: db}, nil
}

func (m *MySQLStorage) Increment(ctx context.Context, key string, window time.Duration) (int, error) {
	tx, err := m.db.BeginTx(ctx, nil)
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
	err = tx.QueryRowContext(ctx, `SELECT count, expires_at FROM rate_limits WHERE k = ? FOR UPDATE`, key).Scan(&count, &expiresAt)
	if err == sql.ErrNoRows {
		count = 1
		expires := time.Now().Add(window)
		_, err = tx.ExecContext(ctx, `INSERT INTO rate_limits (k, count, expires_at) VALUES (?, ?, ?)`, key, count, expires)
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
	_, err = tx.ExecContext(ctx, `UPDATE rate_limits SET count = ?, expires_at = ? WHERE k = ?`, count, expires, key)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *MySQLStorage) SetBan(ctx context.Context, key string, duration time.Duration) error {
	expiresAt := time.Now().Add(duration)
	_, err := m.db.ExecContext(ctx, `INSERT INTO bans (k, expires_at) VALUES (?, ?) ON DUPLICATE KEY UPDATE expires_at=VALUES(expires_at)`, key, expiresAt)
	return err
}

func (m *MySQLStorage) IsBanned(ctx context.Context, key string) (bool, error) {
	var expiresAt time.Time
	err := m.db.QueryRowContext(ctx, `SELECT expires_at FROM bans WHERE k = ?`, key).Scan(&expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, nil
	}
	if expiresAt.After(time.Now()) {
		return true, nil
	}
	return false, nil
}

func (m *MySQLStorage) Close() error {
	return m.db.Close()
}
