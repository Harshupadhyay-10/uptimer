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
	CREATE TABLE IF NOT EXISTS targets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		url TEXT NOT NULL UNIQUE,
		created_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS checks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		target_id INTEGER NOT NULL,
		status_code INTEGER,
		latency_ms INTEGER,
		error TEXT,
		checked_at DATETIME NOT NULL,
		FOREIGN KEY (target_id) REFERENCES targets(id)
	);
	`
	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

// SaveCheck inserts a single check result into the database.
func (s *Store) SaveCheck(targetID int64, statusCode int, latency time.Duration, checkErr error, checkedAt time.Time) error {
	errMsg := ""
	if checkErr != nil {
		errMsg = checkErr.Error()
	}

	_, err := s.db.Exec(
		`INSERT INTO checks (target_id, status_code, latency_ms, error, checked_at) VALUES (?, ?, ?, ?, ?)`,
		targetID, statusCode, latency.Milliseconds(), errMsg, checkedAt,
	)
	return err
}

// Close closes the underlying database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// Target represents a URL being monitored.
type Target struct {
	ID        int64
	URL       string
	CreatedAt time.Time
}

// AddTarget inserts a new URL to monitor and returns its ID.
func (s *Store) AddTarget(url string) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO targets (url, created_at) VALUES (?, ?)`,
		url, time.Now(),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ListTargets returns all monitored targets.
func (s *Store) ListTargets() ([]Target, error) {
	rows, err := s.db.Query(`SELECT id, url, created_at FROM targets`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var targets []Target
	for rows.Next() {
		var t Target
		if err := rows.Scan(&t.ID, &t.URL, &t.CreatedAt); err != nil {
			return nil, err
		}
		targets = append(targets, t)
	}
	return targets, nil
}

// Check represents a single historical check result.
type Check struct {
	ID         int64
	TargetID   int64
	StatusCode int
	LatencyMs  int64
	Error      string
	CheckedAt  time.Time
}

// ChecksForTarget returns the most recent checks for a given target, newest first.
func (s *Store) ChecksForTarget(targetID int64, limit int) ([]Check, error) {
	rows, err := s.db.Query(
		`SELECT id, target_id, status_code, latency_ms, error, checked_at 
		 FROM checks WHERE target_id = ? ORDER BY checked_at DESC LIMIT ?`,
		targetID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checks []Check
	for rows.Next() {
		var c Check
		if err := rows.Scan(&c.ID, &c.TargetID, &c.StatusCode, &c.LatencyMs, &c.Error, &c.CheckedAt); err != nil {
			return nil, err
		}
		checks = append(checks, c)
	}
	return checks, nil
}
