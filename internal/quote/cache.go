package quote

import (
	"fmt"
	"sync"
	"time"
)

type cacheEntry struct {
	best      Rate
	alts      []Rate
	expiresAt time.Time
}

// Cache memoizes carrier quotes for identical shipments. Carrier tariff APIs
// are slow and rate-limited; the storefront re-quotes the same cart on every
// checkout step, so even a short TTL removes most upstream calls.
type Cache struct {
	mu      sync.RWMutex
	ttl     time.Duration
	entries map[string]cacheEntry
}

func NewCache(ttl time.Duration) *Cache {
	return &Cache{ttl: ttl, entries: make(map[string]cacheEntry)}
}

func cacheKey(s Shipment) string {
	// Two shipments with the same scale weight but different dimensions can
	// have different billable (dimensional) weights, so dims must be part of
	// the key — colliding entries were serving stale cheaper quotes for
	// bulky parcels (SHIP-377).
	return fmt.Sprintf("%s|%s|%s|%.2f|%.0fx%.0fx%.0f",
		s.OriginZip, s.DestZip, s.Service, s.WeightKg,
		s.LengthCm, s.WidthCm, s.HeightCm)
}

func (c *Cache) Get(s Shipment) (Rate, []Rate, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.entries[cacheKey(s)]
	if !ok || time.Now().After(entry.expiresAt) {
		return Rate{}, nil, false
	}
	return entry.best, entry.alts, true
}

func (c *Cache) Set(s Shipment, best Rate, alts []Rate) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[cacheKey(s)] = cacheEntry{
		best:      best,
		alts:      alts,
		expiresAt: time.Now().Add(c.ttl),
	}
}
