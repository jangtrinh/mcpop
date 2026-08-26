package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB wraps standard database/sql DB for MCPOp
type DB struct {
	*sql.DB
	Path string
}

// GetDefaultDBPath returns ~/.mcpop/data.db
func GetDefaultDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(home, ".mcpop", "data.db"), nil
}

// Open initializes SQLite connection with WAL mode and creates tables
func Open(dbPath string) (*DB, error) {
	if dbPath == "" {
		var err error
		dbPath, err = GetDefaultDBPath()
		if err != nil {
			return nil, err
		}
	}

	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory %s: %w", dir, err)
	}

	// SQLite connection string with Pragmas for high concurrency & performance
	dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)", dbPath)
	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Recommended connection pool settings for SQLite
	sqlDB.SetMaxOpenConns(1) // modernc sqlite single writer
	sqlDB.SetMaxIdleConns(1)

	db := &DB{
		DB:   sqlDB,
		Path: dbPath,
	}

	if err := db.migrate(); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return db, nil
}

func (db *DB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		command TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'running',
		started_at DATETIME NOT NULL,
		ended_at DATETIME
	);

	CREATE TABLE IF NOT EXISTS messages (
		id TEXT PRIMARY KEY,
		session_id TEXT NOT NULL,
		direction TEXT NOT NULL,
		method TEXT,
		rpc_id TEXT,
		is_error BOOLEAN NOT NULL DEFAULT 0,
		payload TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS tool_calls (
		id TEXT PRIMARY KEY,
		session_id TEXT NOT NULL,
		rpc_id TEXT NOT NULL,
		tool_name TEXT NOT NULL,
		arguments TEXT NOT NULL,
		result TEXT,
		is_error BOOLEAN NOT NULL DEFAULT 0,
		error_message TEXT,
		latency_ms INTEGER DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'pending',
		created_at DATETIME NOT NULL,
		completed_at DATETIME,
		FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS failures (
		id TEXT PRIMARY KEY,
		session_id TEXT NOT NULL,
		tool_call_id TEXT,
		failure_type TEXT NOT NULL,
		description TEXT NOT NULL,
		severity TEXT NOT NULL DEFAULT 'warning',
		created_at DATETIME NOT NULL,
		FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE,
		FOREIGN KEY (tool_call_id) REFERENCES tool_calls(id) ON DELETE SET NULL
	);

	CREATE INDEX IF NOT EXISTS idx_messages_session ON messages(session_id, created_at);
	CREATE INDEX IF NOT EXISTS idx_tool_calls_session ON tool_calls(session_id, created_at);
	CREATE INDEX IF NOT EXISTS idx_tool_calls_lookup ON tool_calls(session_id, rpc_id);
	CREATE INDEX IF NOT EXISTS idx_failures_session ON failures(session_id, created_at);
	`

	_, err := db.Exec(schema)
	return err
}
