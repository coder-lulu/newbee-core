package cache

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2"

	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/ent/oauthprovider"
	"github.com/coder-lulu/newbee-core/rpc/internal/interfaces"
)

// CacheManager manages OAuth provider configuration caching
type CacheManager struct {
	providerCache *ProviderConfigCache
	db           *ent.Client
}

// NewCacheManager creates a new cache manager
func NewCacheManager(db *ent.Client, ttl time.Duration, maxSize int) *CacheManager {
	return &CacheManager{
		providerCache: NewProviderConfigCache(ttl, maxSize),
		db:           db,
	}
}

// GetProviderConfig retrieves a provider configuration with caching
func (cm *CacheManager) GetProviderConfig(ctx context.Context, tenantID uint64, providerType interfaces.OAuthProviderType) (*interfaces.OAuthProviderConfig, error) {
	// Try to get from cache first
	cacheKey := GenerateProviderKey(tenantID, providerType)
	if config, found := cm.providerCache.Get(ctx, cacheKey); found {
		return config, nil
	}

	// Cache miss - fetch from database
	oauthProvider, err := cm.db.OauthProvider.Query().
		Where(
			oauthprovider.TenantIDEQ(tenantID),
			oauthprovider.TypeEQ(string(providerType)),
			oauthprovider.EnabledEQ(true),
		).
		First(ctx)
	
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, &interfaces.OAuthError{
				Type:        interfaces.ErrorTypeConfigurationError,
				Description: fmt.Sprintf("OAuth provider %s not found for tenant %d", providerType, tenantID),
			}
		}
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeConfigurationError,
			Description: fmt.Sprintf("failed to fetch OAuth provider: %v", err),
			Cause:       err,
		}
	}

	// Convert ent model to interfaces config
	config := cm.ConvertEntToConfig(oauthProvider)
	
	// Store in cache
	if err := cm.providerCache.Set(ctx, cacheKey, config); err != nil {
		// Log error but don't fail the request
		// TODO: Add proper logging
	}

	return config, nil
}

// GetProviderConfigByID retrieves a provider configuration by ID with caching
func (cm *CacheManager) GetProviderConfigByID(ctx context.Context, providerID int64) (*interfaces.OAuthProviderConfig, error) {
	// Try to get from cache first
	cacheKey := GenerateProviderIDKey(providerID)
	if config, found := cm.providerCache.Get(ctx, cacheKey); found {
		return config, nil
	}

	// Cache miss - fetch from database
	oauthProvider, err := cm.db.OauthProvider.Get(ctx, uint64(providerID))
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, &interfaces.OAuthError{
				Type:        interfaces.ErrorTypeConfigurationError,
				Description: fmt.Sprintf("OAuth provider with ID %d not found", providerID),
			}
		}
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeConfigurationError,
			Description: fmt.Sprintf("failed to fetch OAuth provider: %v", err),
			Cause:       err,
		}
	}

	// Convert ent model to interfaces config
	config := cm.ConvertEntToConfig(oauthProvider)
	
	// Store in cache
	if err := cm.providerCache.Set(ctx, cacheKey, config); err != nil {
		// Log error but don't fail the request
		// TODO: Add proper logging
	}

	return config, nil
}

// InvalidateProviderConfig removes a provider configuration from cache
func (cm *CacheManager) InvalidateProviderConfig(ctx context.Context, tenantID uint64, providerType interfaces.OAuthProviderType) {
	cacheKey := GenerateProviderKey(tenantID, providerType)
	cm.providerCache.Delete(ctx, cacheKey)
}

// InvalidateProviderConfigByID removes a provider configuration from cache by ID
func (cm *CacheManager) InvalidateProviderConfigByID(ctx context.Context, providerID int64) {
	cacheKey := GenerateProviderIDKey(providerID)
	cm.providerCache.Delete(ctx, cacheKey)
}

// RefreshProviderConfig forces a refresh of provider configuration from database
func (cm *CacheManager) RefreshProviderConfig(ctx context.Context, tenantID uint64, providerType interfaces.OAuthProviderType) (*interfaces.OAuthProviderConfig, error) {
	// Invalidate existing cache entry
	cm.InvalidateProviderConfig(ctx, tenantID, providerType)
	
	// Fetch fresh data
	return cm.GetProviderConfig(ctx, tenantID, providerType)
}

// GetCacheStats returns cache performance statistics
func (cm *CacheManager) GetCacheStats() CacheStats {
	return cm.providerCache.GetStats()
}

// CleanupExpiredEntries removes expired entries from cache
func (cm *CacheManager) CleanupExpiredEntries(ctx context.Context) int {
	return cm.providerCache.CleanupExpired(ctx)
}

// WarmupCache pre-loads frequently used provider configurations
func (cm *CacheManager) WarmupCache(ctx context.Context, tenantIDs []uint64) error {
	for _, tenantID := range tenantIDs {
		// Get all enabled providers for this tenant
		providers, err := cm.db.OauthProvider.Query().
			Where(
				oauthprovider.TenantIDEQ(tenantID),
				oauthprovider.EnabledEQ(true),
			).
			All(ctx)
		
		if err != nil {
			return fmt.Errorf("failed to fetch providers for tenant %d: %v", tenantID, err)
		}

		// Cache each provider configuration
		for _, provider := range providers {
			config := cm.ConvertEntToConfig(provider)
			cacheKey := GenerateProviderKey(tenantID, interfaces.OAuthProviderType(provider.ProviderType))
			
			if err := cm.providerCache.Set(ctx, cacheKey, config); err != nil {
				// Log error but continue with other providers
				// TODO: Add proper logging
			}
			
			// Also cache by ID
			idCacheKey := GenerateProviderIDKey(int64(provider.ID))
			if err := cm.providerCache.Set(ctx, idCacheKey, config); err != nil {
				// Log error but continue
				// TODO: Add proper logging
			}
		}
	}

	return nil
}

// ConvertEntToConfig converts an ent OauthProvider to interfaces.OAuthProviderConfig
func (cm *CacheManager) ConvertEntToConfig(provider *ent.OauthProvider) *interfaces.OAuthProviderConfig {
	config := &interfaces.OAuthProviderConfig{
		Type:            interfaces.OAuthProviderType(provider.Type),
		ClientID:        provider.ClientID,
		ClientSecret:    provider.ClientSecret,
		RedirectURL:     provider.RedirectURL,
		AuthURL:         provider.AuthURL,
		TokenURL:        provider.TokenURL,
		UserInfoURL:     provider.InfoURL,
		Enabled:         provider.Enabled,
		DisplayName:     provider.DisplayName,
		EncryptedSecret: provider.EncryptedSecret,
		EncryptionKeyID: provider.EncryptionKeyID,
	}

	// Handle scopes - parse from string to slice
	if provider.Scopes != "" {
		// Simple split by comma for now - can be enhanced later
		config.Scopes = []string{provider.Scopes}
	}
	
	config.AuthStyle = oauth2.AuthStyle(provider.AuthStyle)
	config.SupportPKCE = provider.SupportPkce

	return config
}

// Global cache manager instance
var globalCacheManager *CacheManager

// InitGlobalCacheManager initializes the global cache manager
func InitGlobalCacheManager(db *ent.Client, ttl time.Duration, maxSize int) {
	globalCacheManager = NewCacheManager(db, ttl, maxSize)
}

// GetGlobalCacheManager returns the global cache manager instance
func GetGlobalCacheManager() *CacheManager {
	return globalCacheManager
}

// SetGlobalCacheManager sets the global cache manager instance
func SetGlobalCacheManager(manager *CacheManager) {
	globalCacheManager = manager
}