package cache

import (
	"fmt"
	"sync"
	"time"

	"github.com/miekg/dns"
)



type Cache struct {
	entries map[string]*CacheEntry
	mu sync.RWMutex

	maxSize int

	// Stats
	hits int64
	misses int64
	evictions int64
}

type CacheEntry struct {
	Response *dns.Msg
	CachedAt time.Time
	ExpiredAt time.Time
}


func NewCache(maxSize int) *Cache {
	return &Cache{
		entries: make(map[string]*CacheEntry),
		maxSize: maxSize,
		hits: 0,
		misses: 0,
		evictions: 0,
	}
}

func (c *Cache) Add(domain string, qtype uint16, response *dns.Msg) {
	// Don't cache response with no answers
	if len(response.Answer) == 0 {
		return
	}

	// Extract TTL from response
	ttl := c.getMinTTL(response)

	//Don't cache if TTL is 0 (no caching requested)
	if ttl == 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Check if we need to enforce size limit
	if c.maxSize > 0 && len(c.entries) >- c.maxSize {
		c.evictOne()
	}

	key := generateKey(domain, qtype)
	now := time.Now()

	c.entries[key] = &CacheEntry{
		Response: response.Copy(),
		CachedAt: now,
		ExpiredAt: now.Add(time.Duration(ttl) *time.Second),
	}
}

func (c *Cache) Get(domain string, qtype uint16) (*dns.Msg, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := generateKey(domain, qtype)
	entry, exists := c.entries[key]

	if !exists {
		c.misses++
		return nil, false
	}

	// Check if entry has expired
	if time.Now().After(entry.ExpiredAt) {
		c.misses++
		return nil, false
	}

	// Cache hit!
	c.hits++

	// Return a copy of the response to avoid modification
	return entry.Response.Copy(), true
}

// Delete removes a specific entry from cache
func (c *Cache) Delete(domain string, qtype uint16) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := generateKey(domain, qtype)
	delete(c.entries, key)
}

// Clear removes all entries from cache
func (c *Cache) Clear()  {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)
	c.hits = 0
	c.misses = 0
}

// CleanExpired - removes all expired entries from cache
// If maxSize is set, also enforces size limit
// Returns the number of entries removed

func (c *Cache) CleanExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	removed := 0
	now := time.Now()

	for key, entry := range c.entries {
		if now.After(entry.ExpiredAt) {
			delete(c.entries, key)
			removed++
		}
	}

	// If we have a size limit and still over it, evict oldest
	if c.maxSize > 0 && len(c.entries) > c.maxSize {
		target := int(float64(c.maxSize) * 0.9) // Reduce to 90% of max
		additionalRemoved := c.evictExpired(target)
		removed += additionalRemoved
	}

	return removed
}


// Size returns the current number of cached entries
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.entries)
}

// CacheStats contains cache performance statistics
type CacheStats struct {
	Size      int     // Number of cached entries
	MaxSize   int     // Maximum size (0 = unlimited)
	Hits      int64   // Cache hits
	Misses    int64   // Cache misses
	Evictions int64   // Number of evictions
	HitRate   float64 // Hit rate percentage
}

// Stats returns current cache statistics
func (c *Cache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(c.hits) / float64(total) * 100.0
	}

	return CacheStats{
		Size:      len(c.entries),
		MaxSize:   c.maxSize,
		Hits:      c.hits,
		Misses:    c.misses,
		Evictions: c.evictions,
		HitRate:   hitRate,
	}
}

// String formats cache statistics for display
func (cs CacheStats) String() string {
	if cs.MaxSize > 0 {
		return fmt.Sprintf("Size: %d/%d | Hits: %d | Misses: %d | Evictions: %d | Hit Rate: %.1f%%",
			cs.Size, cs.MaxSize, cs.Hits, cs.Misses, cs.Evictions, cs.HitRate)
	}
	return fmt.Sprintf("Size: %d | Hits: %d | Misses: %d | Evictions: %d | Hit Rate: %.1f%%",
		cs.Size, cs.Hits, cs.Misses, cs.Evictions, cs.HitRate)
}

// creates a unique cache key from domain and query type
func generateKey(domain string, qtype uint16) string {
	return fmt.Sprintf("%s:%d", domain, qtype)
}


func (c *Cache) getMinTTL(response *dns.Msg) uint32 {
	if len(response.Answer) == 0 {
		return 0
	}

	minTTL := response.Answer[0].Header().Ttl

	for _, answer := range response.Answer {
		ttl := answer.Header().Ttl
		if ttl < minTTL {
			minTTL = ttl
		}
	}
	return minTTL
}


// evictOne removes one entry from cache (oldest first)
// Must be called with lock held
func (c *Cache) evictOne() {
	if len(c.entries) == 0 {
		return
	}

	// Find the oldest entry (earliest expiration time)
	var oldestKey string
	var oldestExpiration time.Time
	first := true

	for key, entry := range c.entries {
		if first || entry.ExpiredAt.Before(oldestExpiration) {
			oldestKey = key
			oldestExpiration = entry.ExpiredAt
			first = false
		}
	}

	// Remove the oldest entry
	if oldestKey != "" {
		delete(c.entries, oldestKey)
		c.evictions++
	}
}

// evictExpired removes all expired entries and returns count
// If more space needed, removes oldest entries
// Must be called with lock held
func (c *Cache) evictExpired(targetSize int) int {
	removed := 0
	now := time.Now()

	// First pass: remove expired entries
	for key, entry := range c.entries {
		if now.After(entry.ExpiredAt) {
			delete(c.entries, key)
			removed++
			c.evictions++
		}
	}

	// If still over target size, remove oldest entries
	if targetSize > 0 {
		for len(c.entries) > targetSize {
			c.evictOne()
			removed++
		}
	}

	return removed
}