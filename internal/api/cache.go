package api

import (
	"sync"
	"time"

	"github.com/jesus/FCCUR/internal/models"
)

// PackageCache caches package metadata in memory
type PackageCache struct {
	mu           sync.RWMutex
	packages     []*models.Package
	lastUpdated  time.Time
	enabled      bool
}

// NewPackageCache creates a new package cache
func NewPackageCache() *PackageCache {
	return &PackageCache{
		packages:    make([]*models.Package, 0),
		lastUpdated: time.Time{},
		enabled:     true,
	}
}

// Get returns cached packages if available
func (c *PackageCache) Get() ([]*models.Package, bool) {
	if !c.enabled {
		return nil, false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.packages == nil || len(c.packages) == 0 {
		return nil, false
	}

	// Return copy to prevent external modifications
	result := make([]*models.Package, len(c.packages))
	copy(result, c.packages)
	return result, true
}

// Set updates the cache with new packages
func (c *PackageCache) Set(packages []*models.Package) {
	if !c.enabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.packages = packages
	c.lastUpdated = time.Now()
}

// Invalidate clears the cache
func (c *PackageCache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.packages = nil
	c.lastUpdated = time.Time{}
}

// Enable enables the cache
func (c *PackageCache) Enable() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enabled = true
}

// Disable disables the cache
func (c *PackageCache) Disable() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enabled = false
	c.packages = nil
}

// LastUpdated returns when the cache was last updated
func (c *PackageCache) LastUpdated() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastUpdated
}

// Size returns the number of cached packages
func (c *PackageCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.packages)
}
