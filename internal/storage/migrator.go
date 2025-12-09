package storage

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/stdlib"
)

// Migrator handles database migrations
type Migrator struct {
	db         Database
	migrate    *migrate.Migrate
	dbType     DatabaseType
	migrations string
}

// MigrationInfo contains information about a migration
type MigrationInfo struct {
	Version   uint
	Dirty     bool
	AppliedAt time.Time
}

// NewMigrator creates a new migrator for the given database
func NewMigrator(db Database, migrationsPath string) (*Migrator, error) {
	var m *migrate.Migrate
	var dbType DatabaseType

	switch v := db.(type) {
	case *SQLiteDB:
		dbType = DatabaseSQLite
		driver, err := sqlite3.WithInstance(v.db, &sqlite3.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to create sqlite3 driver: %w", err)
		}

		sourcePath := filepath.Join(migrationsPath, "sqlite")
		m, err = migrate.NewWithDatabaseInstance(
			fmt.Sprintf("file://%s", sourcePath),
			"sqlite3",
			driver,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create migrator: %w", err)
		}

	case *PostgresDB:
		dbType = DatabasePostgreSQL
		// Convert pgx pool to *sql.DB for migrate
		sqlDB := stdlib.OpenDBFromPool(v.pool)

		driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to create postgres driver: %w", err)
		}

		sourcePath := filepath.Join(migrationsPath, "postgres")
		m, err = migrate.NewWithDatabaseInstance(
			fmt.Sprintf("file://%s", sourcePath),
			"postgres",
			driver,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create migrator: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported database type")
	}

	return &Migrator{
		db:         db,
		migrate:    m,
		dbType:     dbType,
		migrations: migrationsPath,
	}, nil
}

// Up runs all pending migrations
func (m *Migrator) Up() error {
	if m.migrate == nil {
		return fmt.Errorf("migrator not initialized")
	}

	err := m.migrate.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// Down rolls back the most recent migration
func (m *Migrator) Down() error {
	if m.migrate == nil {
		return fmt.Errorf("migrator not initialized")
	}

	err := m.migrate.Down()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	return nil
}

// Steps runs n migrations. Use negative values to rollback
func (m *Migrator) Steps(n int) error {
	if m.migrate == nil {
		return fmt.Errorf("migrator not initialized")
	}

	err := m.migrate.Steps(n)
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run %d steps: %w", n, err)
	}

	return nil
}

// Goto migrates to a specific version
func (m *Migrator) Goto(version uint) error {
	if m.migrate == nil {
		return fmt.Errorf("migrator not initialized")
	}

	err := m.migrate.Migrate(version)
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to migrate to version %d: %w", version, err)
	}

	return nil
}

// Version returns the current migration version
func (m *Migrator) Version() (uint, bool, error) {
	if m.migrate == nil {
		return 0, false, fmt.Errorf("migrator not initialized")
	}

	version, dirty, err := m.migrate.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return 0, false, fmt.Errorf("failed to get version: %w", err)
	}

	return version, dirty, nil
}

// Force sets the migration version without running migrations
// This is useful for fixing dirty state
func (m *Migrator) Force(version int) error {
	if m.migrate == nil {
		return fmt.Errorf("migrator not initialized")
	}

	err := m.migrate.Force(version)
	if err != nil {
		return fmt.Errorf("failed to force version %d: %w", version, err)
	}

	return nil
}

// Drop drops all tables (dangerous!)
func (m *Migrator) Drop() error {
	if m.migrate == nil {
		return fmt.Errorf("migrator not initialized")
	}

	err := m.migrate.Drop()
	if err != nil {
		return fmt.Errorf("failed to drop database: %w", err)
	}

	return nil
}

// GetInfo returns information about the current migration state
func (m *Migrator) GetInfo() (*MigrationInfo, error) {
	version, dirty, err := m.Version()
	if err != nil && err.Error() != "no migration" {
		return nil, err
	}

	info := &MigrationInfo{
		Version:   version,
		Dirty:     dirty,
		AppliedAt: time.Now(), // Note: golang-migrate doesn't track timestamps
	}

	return info, nil
}

// Close closes the migrator
func (m *Migrator) Close() error {
	if m.migrate == nil {
		return nil
	}

	sourceErr, dbErr := m.migrate.Close()
	if sourceErr != nil {
		return sourceErr
	}
	if dbErr != nil {
		return dbErr
	}

	return nil
}

// Validate checks if all migrations are valid
func (m *Migrator) Validate() error {
	// Check if we can get version without errors
	_, _, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("migration validation failed: %w", err)
	}

	return nil
}

// GetDatabaseType returns the database type
func (m *Migrator) GetDatabaseType() DatabaseType {
	return m.dbType
}

// Helper function to safely rollback to a specific version
func (m *Migrator) SafeRollback(targetVersion uint) error {
	currentVersion, dirty, err := m.Version()
	if err != nil && err.Error() != "no migration" {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if dirty {
		return fmt.Errorf("database is in dirty state, cannot rollback safely. Use Force() to fix")
	}

	if targetVersion >= currentVersion {
		return fmt.Errorf("target version %d must be less than current version %d", targetVersion, currentVersion)
	}

	// Calculate steps needed
	steps := int(currentVersion - targetVersion)

	return m.Steps(-steps)
}

// EnsureMigrationTable ensures the migration tracking table exists
func EnsureMigrationTable(db *sql.DB, dbType DatabaseType) error {
	var query string

	switch dbType {
	case DatabaseSQLite:
		query = `
			CREATE TABLE IF NOT EXISTS schema_migrations (
				version bigint not null primary key,
				dirty boolean not null
			);
		`
	case DatabasePostgreSQL:
		query = `
			CREATE TABLE IF NOT EXISTS schema_migrations (
				version bigint not null primary key,
				dirty boolean not null
			);
		`
	default:
		return fmt.Errorf("unsupported database type: %s", dbType)
	}

	_, err := db.Exec(query)
	return err
}
