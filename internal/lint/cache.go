package lint

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/rebeliceyang/schema-lint-mcp-server/pkg/models"
)

// Cache provides caching for AI analysis results
type Cache struct {
	items map[string]cacheItem
	ttl   time.Duration
	mu    sync.RWMutex
}

type cacheItem struct {
	issues    []models.Issue
	timestamp time.Time
}

// NewCache creates a new cache instance
func NewCache(ttl time.Duration) *Cache {
	c := &Cache{
		items: make(map[string]cacheItem),
		ttl:   ttl,
	}

	// Start cleanup goroutine
	go c.cleanup()

	return c
}

// Key generates a cache key from schema content, rule ID, and provider
func (c *Cache) Key(schemaContent, ruleID, provider string) string {
	h := sha256.New()
	h.Write([]byte(schemaContent))
	h.Write([]byte(ruleID))
	h.Write([]byte(provider))
	return hex.EncodeToString(h.Sum(nil))
}

// Get retrieves issues from cache
func (c *Cache) Get(key string) ([]models.Issue, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// Check if item has expired
	if time.Since(item.timestamp) > c.ttl {
		return nil, false
	}

	// Return a copy to prevent modification
	issues := make([]models.Issue, len(item.issues))
	copy(issues, item.issues)

	return issues, true
}

// Set stores issues in cache
func (c *Cache) Set(key string, issues []models.Issue) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Make a copy to prevent external modification
	issuesCopy := make([]models.Issue, len(issues))
	copy(issuesCopy, issues)

	c.items[key] = cacheItem{
		issues:    issuesCopy,
		timestamp: time.Now(),
	}
}

// cleanup removes expired items from cache
func (c *Cache) cleanup() {
	ticker := time.NewTicker(c.ttl / 2)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.items {
			if now.Sub(item.timestamp) > c.ttl {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

// Clear removes all items from cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]cacheItem)
}

// Size returns the number of items in cache
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Stats returns cache statistics
func (c *Cache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := CacheStats{
		Size: len(c.items),
		TTL:  c.ttl,
	}

	// Calculate hit rate if we were tracking it
	// For now, just return basic stats

	return stats
}

// CacheStats holds cache statistics
type CacheStats struct {
	Size     int
	TTL      time.Duration
	HitRate  float64
	Hits     int64
	Misses   int64
}

// String returns a string representation of cache stats
func (cs CacheStats) String() string {
	return fmt.Sprintf("Cache Stats: Size=%d, TTL=%s, Hits=%d, Misses=%d, HitRate=%.2f%%",
		cs.Size, cs.TTL, cs.Hits, cs.Misses, cs.HitRate*100)
}