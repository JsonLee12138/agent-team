package ingest

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	rps   rate.Limit
	burst int
	mu    sync.Mutex
	items map[string]*rateLimiterEntry
}

type rateLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func NewRateLimiter(rps float64, burst int) *RateLimiter {
	return &RateLimiter{
		rps:   rate.Limit(rps),
		burst: burst,
		items: make(map[string]*rateLimiterEntry),
	}
}

func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, ok := r.items[key]
	if !ok {
		entry = &rateLimiterEntry{limiter: rate.NewLimiter(r.rps, r.burst), lastSeen: time.Now()}
		r.items[key] = entry
	}
	entry.lastSeen = time.Now()
	return entry.limiter.Allow()
}

func (r *RateLimiter) Cleanup(maxAge time.Duration) {
	cutoff := time.Now().Add(-maxAge)
	r.mu.Lock()
	defer r.mu.Unlock()
	for key, entry := range r.items {
		if entry.lastSeen.Before(cutoff) {
			delete(r.items, key)
		}
	}
}
