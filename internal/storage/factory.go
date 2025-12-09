package storage

import (
	"fmt"
	"strings"
)

// NewDatabase creates a database connection based on the connection string
// Auto-detects database type from connection string format
func NewDatabase(connString string) (Database, error) {
	dbType := DetectDatabaseType(connString)

	switch dbType {
	case DatabasePostgreSQL:
		return NewPostgresDatabase(connString)
	case DatabaseSQLite:
		return NewSQLiteDatabase(connString)
	default:
		return nil, fmt.Errorf("unsupported database type")
	}
}

// DetectDatabaseType detects the database type from connection string
func DetectDatabaseType(connString string) DatabaseType {
	connLower := strings.ToLower(connString)

	// PostgreSQL patterns
	if strings.HasPrefix(connLower, "postgres://") ||
		strings.HasPrefix(connLower, "postgresql://") ||
		strings.Contains(connLower, "host=") ||
		strings.Contains(connLower, "port=5432") {
		return DatabasePostgreSQL
	}

	// SQLite patterns (file path or :memory:)
	if strings.HasSuffix(connLower, ".db") ||
		strings.HasSuffix(connLower, ".sqlite") ||
		strings.HasSuffix(connLower, ".sqlite3") ||
		connString == ":memory:" ||
		!strings.Contains(connString, "://") {
		return DatabaseSQLite
	}

	// Default to SQLite for backward compatibility
	return DatabaseSQLite
}

// GetDatabaseInfo returns information about the connected database
func GetDatabaseInfo(db Database) string {
	switch db.(type) {
	case *PostgresDB:
		return "PostgreSQL (with pgx connection pooling)"
	case *SQLiteDB:
		return "SQLite3 (single-user mode)"
	default:
		return "Unknown database type"
	}
}
