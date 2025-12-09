package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresDB implements the Database interface for PostgreSQL
type PostgresDB struct {
	pool *pgxpool.Pool
}

// NewPostgresDatabase creates a new PostgreSQL database connection with connection pooling
func NewPostgresDatabase(connString string) (*PostgresDB, error) {
	// Parse connection string and create pool config
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Set connection pool settings
	config.MaxConns = 25                         // Maximum connections in pool
	config.MinConns = 5                          // Minimum connections to maintain
	config.MaxConnLifetime = time.Hour           // Max lifetime of a connection
	config.MaxConnIdleTime = 30 * time.Minute    // Max idle time before closing
	config.HealthCheckPeriod = 1 * time.Minute   // Health check interval

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresDB{pool: pool}, nil
}

// Close closes all connections in the pool
func (p *PostgresDB) Close() error {
	p.pool.Close()
	return nil
}

// Ping checks the database connection
func (p *PostgresDB) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return p.pool.Ping(ctx)
}

// Migrate runs database migrations using the migration system
func (p *PostgresDB) Migrate() error {
	// Use golang-migrate for proper version tracking
	migrator, err := NewMigrator(p, GetMigrationsPath())
	if err != nil {
		// Fallback to old schema if migrations not available
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, err := p.pool.Exec(ctx, postgresSchema)
		return err
	}
	defer migrator.Close()

	return migrator.Up()
}

// Helper function to get context with timeout
func (p *PostgresDB) getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
