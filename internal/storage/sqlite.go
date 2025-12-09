package storage

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteDB implements the Database interface for SQLite
type SQLiteDB struct {
	db *sql.DB
}

// NewSQLiteDatabase creates a new SQLite database connection
func NewSQLiteDatabase(path string) (*SQLiteDB, error) {
	// Add parseTime parameter to handle datetime properly
	db, err := sql.Open("sqlite3", path+"?parseTime=true")
	if err != nil {
		return nil, err
	}

	// SQLite optimizations
	db.SetMaxOpenConns(1) // SQLite only supports 1 writer
	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA synchronous=NORMAL")
	db.Exec("PRAGMA cache_size=10000")

	return &SQLiteDB{db: db}, nil
}

// Close closes the database connection
func (s *SQLiteDB) Close() error {
	return s.db.Close()
}

// Ping checks the database connection
func (s *SQLiteDB) Ping() error {
	return s.db.Ping()
}

// Migrate runs database migrations using the migration system
func (s *SQLiteDB) Migrate() error {
	// Use golang-migrate for proper version tracking
	migrator, err := NewMigrator(s, GetMigrationsPath())
	if err != nil {
		// Fallback to old schema if migrations not available
		_, err := s.db.Exec(schema)
		return err
	}
	defer migrator.Close()

	return migrator.Up()
}
