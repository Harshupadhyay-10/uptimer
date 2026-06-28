package storage

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

// Store wraps a database connection for persisting check results.
type Store struct {
	db *sql.DB
}

// New opens (or creates) a SQLite database at the given path
// and ensures the required tables exist.
func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS checks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		url TEXT NOT NULL,
		status_code INTEGER,
		latency_ms INTEGER,
		error TEXT,
		checked_at DATETIME NOT NULL
	);
	`
	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

// SaveCheck inserts a single check result into the database.
func (s *Store) SaveCheck(url string, statusCode int, latency time.Duration, checkErr error, checkedAt time.Time) error {
	errMsg := ""
	if checkErr != nil {
		errMsg = checkErr.Error()
	}

	_, err := s.db.Exec(
		`INSERT INTO checks (url, status_code, latency_ms, error, checked_at) VALUES (?, ?, ?, ?, ?)`,
		url, statusCode, latency.Milliseconds(), errMsg, checkedAt,
	)
	return err
}

// Close closes the underlying database connection.
func (s *Store) Close() error {
	return s.db.Close()
}
