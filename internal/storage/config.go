package storage

import "sync"

var (
	migrationsPath string
	mu             sync.RWMutex
)

// SetMigrationsPath sets the global migrations directory path
func SetMigrationsPath(path string) {
	mu.Lock()
	defer mu.Unlock()
	migrationsPath = path
}

// GetMigrationsPath returns the migrations directory path
func GetMigrationsPath() string {
	mu.RLock()
	defer mu.RUnlock()
	if migrationsPath == "" {
		return "./migrations"
	}
	return migrationsPath
}
