package provider

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/coder-lulu/newbee-core/rpc/internal/interfaces"
)

// ProviderRegistry manages OAuth provider instances and their lifecycle
type ProviderRegistry struct {
	mu             sync.RWMutex
	providers      map[string]interfaces.OAuthAdapter // key: tenantID:providerType
	factory        *ProviderFactory
	lastAccessed   map[string]time.Time
	cleanupTicker  *time.Ticker
	stopCleanup    chan struct{}
	maxIdleTime    time.Duration
	cleanupRunning bool
}

// ProviderKey generates a unique key for provider instances
func ProviderKey(tenantID uint64, providerType interfaces.OAuthProviderType) string {
	return fmt.Sprintf("%d:%s", tenantID, providerType)
}

// NewProviderRegistry creates a new provider registry
func NewProviderRegistry(factory *ProviderFactory, maxIdleTime time.Duration) *ProviderRegistry {
	registry := &ProviderRegistry{
		providers:    make(map[string]interfaces.OAuthAdapter),
		factory:      factory,
		lastAccessed: make(map[string]time.Time),
		stopCleanup:  make(chan struct{}),
		maxIdleTime:  maxIdleTime,
	}

	// Start cleanup routine
	registry.startCleanup()

	return registry
}

// GetProvider retrieves or creates a provider instance
func (pr *ProviderRegistry) GetProvider(ctx context.Context, tenantID uint64, providerType interfaces.OAuthProviderType) (interfaces.OAuthAdapter, error) {
	key := ProviderKey(tenantID, providerType)

	// Check if provider already exists
	pr.mu.RLock()
	provider, exists := pr.providers[key]
	pr.mu.RUnlock()

	if exists {
		// Update last accessed time
		pr.mu.Lock()
		pr.lastAccessed[key] = time.Now()
		pr.mu.Unlock()
		return provider, nil
	}

	// Provider doesn't exist, create it
	return pr.createAndCacheProvider(ctx, tenantID, providerType)
}

// createAndCacheProvider creates a new provider and caches it
func (pr *ProviderRegistry) createAndCacheProvider(ctx context.Context, tenantID uint64, providerType interfaces.OAuthProviderType) (interfaces.OAuthAdapter, error) {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	key := ProviderKey(tenantID, providerType)

	// Double-check if provider was created while waiting for lock
	if provider, exists := pr.providers[key]; exists {
		pr.lastAccessed[key] = time.Now()
		return provider, nil
	}

	// Create provider using factory
	provider, err := pr.factory.GetConfiguredAdapter(ctx, tenantID, providerType)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %v", err)
	}

	// Cache the provider
	pr.providers[key] = provider
	pr.lastAccessed[key] = time.Now()

	return provider, nil
}

// InvalidateProvider removes a provider from the registry
func (pr *ProviderRegistry) InvalidateProvider(tenantID uint64, providerType interfaces.OAuthProviderType) {
	key := ProviderKey(tenantID, providerType)

	pr.mu.Lock()
	defer pr.mu.Unlock()

	delete(pr.providers, key)
	delete(pr.lastAccessed, key)
}

// InvalidateAllForTenant removes all providers for a specific tenant
func (pr *ProviderRegistry) InvalidateAllForTenant(tenantID uint64) {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	tenantPrefix := fmt.Sprintf("%d:", tenantID)
	keysToDelete := make([]string, 0)

	for key := range pr.providers {
		if len(key) > len(tenantPrefix) && key[:len(tenantPrefix)] == tenantPrefix {
			keysToDelete = append(keysToDelete, key)
		}
	}

	for _, key := range keysToDelete {
		delete(pr.providers, key)
		delete(pr.lastAccessed, key)
	}
}

// RefreshProvider forces a refresh of a provider instance
func (pr *ProviderRegistry) RefreshProvider(ctx context.Context, tenantID uint64, providerType interfaces.OAuthProviderType) (interfaces.OAuthAdapter, error) {
	// Invalidate existing provider
	pr.InvalidateProvider(tenantID, providerType)

	// Create new provider
	return pr.GetProvider(ctx, tenantID, providerType)
}

// ListProviders returns information about cached providers
func (pr *ProviderRegistry) ListProviders() []ProviderInfo {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	infos := make([]ProviderInfo, 0, len(pr.providers))
	now := time.Now()

	for key, provider := range pr.providers {
		lastAccessed, exists := pr.lastAccessed[key]
		if !exists {
			lastAccessed = now
		}

		infos = append(infos, ProviderInfo{
			Key:          key,
			ProviderType: provider.GetProviderType(),
			LastAccessed: lastAccessed,
			IdleTime:     now.Sub(lastAccessed),
		})
	}

	return infos
}

// GetStats returns registry statistics
func (pr *ProviderRegistry) GetStats() RegistryStats {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	stats := RegistryStats{
		TotalProviders: len(pr.providers),
		MaxIdleTime:    pr.maxIdleTime,
		CleanupRunning: pr.cleanupRunning,
		LastUpdated:    time.Now(),
	}

	// Count by provider type
	stats.ProvidersByType = make(map[interfaces.OAuthProviderType]int)
	for _, provider := range pr.providers {
		stats.ProvidersByType[provider.GetProviderType()]++
	}

	return stats
}

// startCleanup starts the background cleanup routine
func (pr *ProviderRegistry) startCleanup() {
	if pr.cleanupRunning {
		return
	}

	pr.cleanupRunning = true
	pr.cleanupTicker = time.NewTicker(pr.maxIdleTime / 2) // Cleanup every half of max idle time

	go func() {
		for {
			select {
			case <-pr.cleanupTicker.C:
				pr.performCleanup()
			case <-pr.stopCleanup:
				return
			}
		}
	}()
}

// stopCleanupRoutine stops the background cleanup routine
func (pr *ProviderRegistry) stopCleanupRoutine() {
	if !pr.cleanupRunning {
		return
	}

	pr.cleanupRunning = false
	close(pr.stopCleanup)

	if pr.cleanupTicker != nil {
		pr.cleanupTicker.Stop()
	}
}

// performCleanup removes idle providers from the registry
func (pr *ProviderRegistry) performCleanup() {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	now := time.Now()
	keysToDelete := make([]string, 0)

	for key, lastAccessed := range pr.lastAccessed {
		if now.Sub(lastAccessed) > pr.maxIdleTime {
			keysToDelete = append(keysToDelete, key)
		}
	}

	for _, key := range keysToDelete {
		delete(pr.providers, key)
		delete(pr.lastAccessed, key)
	}
}

// ForceCleanup manually triggers a cleanup operation
func (pr *ProviderRegistry) ForceCleanup() int {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	initialCount := len(pr.providers)
	now := time.Now()
	keysToDelete := make([]string, 0)

	for key, lastAccessed := range pr.lastAccessed {
		if now.Sub(lastAccessed) > pr.maxIdleTime {
			keysToDelete = append(keysToDelete, key)
		}
	}

	for _, key := range keysToDelete {
		delete(pr.providers, key)
		delete(pr.lastAccessed, key)
	}

	return initialCount - len(pr.providers)
}

// SetMaxIdleTime updates the maximum idle time for providers
func (pr *ProviderRegistry) SetMaxIdleTime(maxIdleTime time.Duration) {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	pr.maxIdleTime = maxIdleTime

	// Restart cleanup with new interval
	if pr.cleanupRunning {
		pr.stopCleanupRoutine()
		pr.startCleanup()
	}
}

// GetMaxIdleTime returns the current maximum idle time
func (pr *ProviderRegistry) GetMaxIdleTime() time.Duration {
	pr.mu.RLock()
	defer pr.mu.RUnlock()
	return pr.maxIdleTime
}

// Shutdown gracefully shuts down the provider registry
func (pr *ProviderRegistry) Shutdown() {
	pr.stopCleanupRoutine()

	pr.mu.Lock()
	defer pr.mu.Unlock()

	// Clear all providers
	pr.providers = make(map[string]interfaces.OAuthAdapter)
	pr.lastAccessed = make(map[string]time.Time)
}

// ProviderInfo represents information about a cached provider
type ProviderInfo struct {
	Key          string                        `json:"key"`
	ProviderType interfaces.OAuthProviderType `json:"provider_type"`
	LastAccessed time.Time                    `json:"last_accessed"`
	IdleTime     time.Duration                `json:"idle_time"`
}

// RegistryStats represents statistics about the provider registry
type RegistryStats struct {
	TotalProviders    int                                               `json:"total_providers"`
	ProvidersByType   map[interfaces.OAuthProviderType]int             `json:"providers_by_type"`
	MaxIdleTime       time.Duration                                     `json:"max_idle_time"`
	CleanupRunning    bool                                              `json:"cleanup_running"`
	LastUpdated       time.Time                                         `json:"last_updated"`
}

// Global provider registry instance
var globalProviderRegistry *ProviderRegistry

// InitGlobalProviderRegistry initializes the global provider registry
func InitGlobalProviderRegistry(factory *ProviderFactory, maxIdleTime time.Duration) {
	globalProviderRegistry = NewProviderRegistry(factory, maxIdleTime)
}

// GetGlobalProviderRegistry returns the global provider registry instance
func GetGlobalProviderRegistry() *ProviderRegistry {
	return globalProviderRegistry
}

// SetGlobalProviderRegistry sets the global provider registry instance
func SetGlobalProviderRegistry(registry *ProviderRegistry) {
	globalProviderRegistry = registry
}

// ShutdownGlobalProviderRegistry shuts down the global provider registry
func ShutdownGlobalProviderRegistry() {
	if globalProviderRegistry != nil {
		globalProviderRegistry.Shutdown()
	}
}