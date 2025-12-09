package storage

import (
	"time"

	"github.com/jesus/FCCUR/internal/models"
)

// Database interface abstracts database operations
// Implemented by both SQLite and PostgreSQL
type Database interface {
	// Package operations
	CreatePackage(pkg *models.Package) (int64, error)
	GetPackage(id int64) (*models.Package, error)
	GetPackages() ([]*models.Package, error)
	ListPackages(limit, offset int, category, platform, contentType, courseName string) ([]*models.Package, error)
	DeletePackage(id int64) error
	FindPackageByHash(hash string) (*models.Package, error)

	// Download tracking
	RecordDownload(packageID int64, ipAddress, userAgent string) error
	GetDownloadCount(packageID int64) (int64, error)
	GetTotalDownloads() (int64, error)

	// Statistics
	GetPackageCount() (int64, error)
	GetTotalSize() (int64, error)
	GetRecentPackages(limit int) ([]*models.Package, error)
	GetStats() ([]*models.DownloadStats, error)

	// User operations
	CreateUser(email, passwordHash, fullName string, role models.UserRole) (*models.User, error)
	GetUserByID(id int64) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserByResetToken(token string) (*models.User, error)
	UpdateUserLastLogin(userID int64) error
	UpdateUserPassword(userID int64, passwordHash string) error
	SetUserResetToken(userID int64, token string, expiry time.Time) error
	SetUserEmailVerified(userID int64) error

	// Session operations
	CreateSession(userID int64, token, refreshToken, ipAddress, userAgent string, expiresAt time.Time) (*models.Session, error)
	GetSessionByID(id int64) (*models.Session, error)
	GetSessionByToken(token string) (*models.Session, error)
	GetSessionByRefreshToken(refreshToken string) (*models.Session, error)
	DeleteSession(token string) error
	DeleteUserSessions(userID int64) error
	CleanExpiredSessions() error
	ListUserSessions(userID int64) ([]*models.Session, error)

	// Database management
	Migrate() error
	Close() error
	Ping() error
}

// DatabaseType represents the type of database
type DatabaseType string

const (
	DatabaseSQLite     DatabaseType = "sqlite"
	DatabasePostgreSQL DatabaseType = "postgresql"
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type             DatabaseType
	ConnectionString string
	MaxOpenConns     int
	MaxIdleConns     int
	ConnMaxLifetime  time.Duration
}
