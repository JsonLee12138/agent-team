package internal

import (
	"sync"
	"time"
)

// CatalogStore caches catalog reads with a refresh interval.
type CatalogStore struct {
	path            string
	refreshInterval time.Duration
	nowFn           func() time.Time

	mu       sync.RWMutex
	loaded   bool
	lastLoad time.Time
	catalog  RoleRepoCatalog
}

// NewCatalogStore creates a store for the catalog path.
// refreshInterval <= 0 disables caching (always read from disk).
func NewCatalogStore(path string, refreshInterval time.Duration) *CatalogStore {
	return &CatalogStore{
		path:            path,
		refreshInterval: refreshInterval,
		nowFn:           time.Now,
	}
}

// Get returns the current catalog snapshot.
func (s *CatalogStore) Get() (RoleRepoCatalog, error) {
	if s == nil {
		return RoleRepoCatalog{}, nil
	}
	if s.refreshInterval <= 0 {
		return ReadRoleRepoCatalog(s.path)
	}

	now := s.nowFn()
	s.mu.RLock()
	if s.loaded && now.Sub(s.lastLoad) < s.refreshInterval {
		catalog := cloneCatalog(s.catalog)
		s.mu.RUnlock()
		return catalog, nil
	}
	s.mu.RUnlock()
	return s.Refresh()
}

// Refresh reloads the catalog from disk.
func (s *CatalogStore) Refresh() (RoleRepoCatalog, error) {
	if s == nil {
		return RoleRepoCatalog{}, nil
	}
	if s.refreshInterval <= 0 {
		return ReadRoleRepoCatalog(s.path)
	}

	now := s.nowFn()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.loaded && now.Sub(s.lastLoad) < s.refreshInterval {
		return cloneCatalog(s.catalog), nil
	}
	catalog, err := ReadRoleRepoCatalog(s.path)
	if err != nil {
		return RoleRepoCatalog{}, err
	}
	s.catalog = catalog
	s.loaded = true
	s.lastLoad = now
	return cloneCatalog(catalog), nil
}

// Set updates the cached catalog snapshot.
func (s *CatalogStore) Set(catalog RoleRepoCatalog) {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.catalog = catalog
	s.loaded = true
	s.lastLoad = s.nowFn()
	s.mu.Unlock()
}

func cloneCatalog(catalog RoleRepoCatalog) RoleRepoCatalog {
	copyEntries := make([]RoleRepoCatalogEntry, len(catalog.Entries))
	copy(copyEntries, catalog.Entries)
	catalog.Entries = copyEntries
	return catalog
}
