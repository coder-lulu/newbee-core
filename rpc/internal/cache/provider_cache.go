package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/coder-lulu/newbee-core/rpc/internal/interfaces"
)

// CacheEntry represents a cached provider configuration
type CacheEntry struct {
	Config    *interfaces.OAuthProviderConfig
	ExpiresAt time.Time
	CreatedAt time.Time
}

// IsExpired checks if the cache entry has expired
func (e *CacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// ProviderConfigCache manages cached OAuth provider configurations
type ProviderConfigCache struct {
	mu       sync.RWMutex
	cache    map[string]*CacheEntry
	ttl      time.Duration
	maxSize  int
	stats    CacheStats
	onEvict  func(key string, entry *CacheEntry)
}

// CacheStats tracks cache performance metrics
type CacheStats struct {
	Hits       int64 `json:"hits"`
	Misses     int64 `json:"misses"`
	Evictions  int64 `json:"evictions"`
	Expirations int64 `json:"expirations"`
	Size       int   `json:"size"`
}

// GetHitRate returns the cache hit rate as a percentage
func (s *CacheStats) GetHitRate() float64 {
	total := s.Hits + s.Misses
	if total == 0 {
		return 0
	}
	return float64(s.Hits) / float64(total) * 100
}

// NewProviderConfigCache creates a new provider configuration cache
func NewProviderConfigCache(ttl time.Duration, maxSize int) *ProviderConfigCache {
	return &ProviderConfigCache{
		cache:   make(map[string]*CacheEntry),
		ttl:     ttl,
		maxSize: maxSize,
		stats:   CacheStats{},
	}
}

// SetEvictionCallback sets a callback function that's called when entries are evicted
func (c *ProviderConfigCache) SetEvictionCallback(callback func(key string, entry *CacheEntry)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onEvict = callback
}

// Get retrieves a provider configuration from cache
func (c *ProviderConfigCache) Get(ctx context.Context, key string) (*interfaces.OAuthProviderConfig, bool) {
	c.mu.RLock()
	entry, exists := c.cache[key]
	c.mu.RUnlock()

	if !exists {
		c.mu.Lock()
		c.stats.Misses++
		c.mu.Unlock()
		return nil, false
	}

	if entry.IsExpired() {
		c.mu.Lock()
		delete(c.cache, key)
		c.stats.Expirations++
		c.stats.Size--
		c.mu.Unlock()
		
		if c.onEvict != nil {
			c.onEvict(key, entry)
		}
		
		return nil, false
	}

	c.mu.Lock()
	c.stats.Hits++
	c.mu.Unlock()

	// Return a copy to prevent external modifications
	configCopy := *entry.Config
	return &configCopy, true
}

// Set stores a provider configuration in cache
func (c *ProviderConfigCache) Set(ctx context.Context, key string, config *interfaces.OAuthProviderConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to evict entries to make room
	if len(c.cache) >= c.maxSize {
		c.evictOldest()
	}

	// Store a copy to prevent external modifications
	configCopy := *config
	entry := &CacheEntry{
		Config:    &configCopy,
		ExpiresAt: time.Now().Add(c.ttl),
		CreatedAt: time.Now(),
	}

	c.cache[key] = entry
	c.stats.Size = len(c.cache)

	return nil
}

// Delete removes a provider configuration from cache
func (c *ProviderConfigCache) Delete(ctx context.Context, key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.cache[key]
	if !exists {
		return false
	}

	delete(c.cache, key)
	c.stats.Size = len(c.cache)

	if c.onEvict != nil {
		c.onEvict(key, entry)
	}

	return true
}

// Clear removes all entries from cache
func (c *ProviderConfigCache) Clear(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.onEvict != nil {
		for key, entry := range c.cache {
			c.onEvict(key, entry)
		}
	}

	c.cache = make(map[string]*CacheEntry)
	c.stats.Size = 0
}

// GetStats returns cache performance statistics
func (c *ProviderConfigCache) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stats
}

// GetSize returns the current number of entries in cache
func (c *ProviderConfigCache) GetSize() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}

// GetKeys returns all keys currently in cache
func (c *ProviderConfigCache) GetKeys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.cache))
	for key := range c.cache {
		keys = append(keys, key)
	}
	return keys
}

// CleanupExpired removes all expired entries from cache
func (c *ProviderConfigCache) CleanupExpired(ctx context.Context) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiredCount := 0
	now := time.Now()

	for key, entry := range c.cache {
		if now.After(entry.ExpiresAt) {
			delete(c.cache, key)
			c.stats.Expirations++
			expiredCount++

			if c.onEvict != nil {
				c.onEvict(key, entry)
			}
		}
	}

	c.stats.Size = len(c.cache)
	return expiredCount
}

// evictOldest removes the oldest entry from cache (LRU-like behavior)
func (c *ProviderConfigCache) evictOldest() {
	if len(c.cache) == 0 {
		return
	}

	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.cache {
		if oldestKey == "" || entry.CreatedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.CreatedAt
		}
	}

	if oldestKey != "" {
		entry := c.cache[oldestKey]
		delete(c.cache, oldestKey)
		c.stats.Evictions++

		if c.onEvict != nil {
			c.onEvict(oldestKey, entry)
		}
	}
}

// SetTTL updates the cache TTL
func (c *ProviderConfigCache) SetTTL(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ttl = ttl
}

// GetTTL returns the current cache TTL
func (c *ProviderConfigCache) GetTTL() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ttl
}

// SetMaxSize updates the maximum cache size
func (c *ProviderConfigCache) SetMaxSize(maxSize int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.maxSize = maxSize
	
	// Evict entries if current size exceeds new max size
	for len(c.cache) > maxSize && len(c.cache) > 0 {
		c.evictOldest()
	}
	
	c.stats.Size = len(c.cache)
}

// GetMaxSize returns the maximum cache size
func (c *ProviderConfigCache) GetMaxSize() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.maxSize
}

// GenerateProviderKey generates a cache key for a provider configuration
func GenerateProviderKey(tenantID uint64, providerType interfaces.OAuthProviderType) string {
	return fmt.Sprintf("provider:%d:%s", tenantID, providerType)
}

// GenerateProviderIDKey generates a cache key for a provider by ID
func GenerateProviderIDKey(providerID int64) string {
	return fmt.Sprintf("provider_id:%d", providerID)
}

// Global cache instance
var globalProviderCache *ProviderConfigCache

// InitGlobalProviderCache initializes the global provider cache
func InitGlobalProviderCache(ttl time.Duration, maxSize int) {
	globalProviderCache = NewProviderConfigCache(ttl, maxSize)
}

// GetGlobalProviderCache returns the global provider cache instance
func GetGlobalProviderCache() *ProviderConfigCache {
	if globalProviderCache == nil {
		// Default configuration: 1 hour TTL, 1000 entries max
		globalProviderCache = NewProviderConfigCache(time.Hour, 1000)
	}
	return globalProviderCache
}

// SetGlobalProviderCache sets the global provider cache instance
func SetGlobalProviderCache(cache *ProviderConfigCache) {
	globalProviderCache = cache
}