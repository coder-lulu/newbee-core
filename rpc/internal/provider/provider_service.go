package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/internal/cache"
	"github.com/coder-lulu/newbee-core/rpc/internal/encryption"
	"github.com/coder-lulu/newbee-core/rpc/internal/interfaces"
)

// ProviderService provides comprehensive OAuth provider management
type ProviderService struct {
	db                *ent.Client
	cacheManager      *cache.CacheManager
	encryptionService *encryption.ProviderEncryptionService
}

// NewProviderService creates a new provider service
func NewProviderService(db *ent.Client, cacheManager *cache.CacheManager, encryptionService *encryption.ProviderEncryptionService) *ProviderService {
	return &ProviderService{
		db:                db,
		cacheManager:      cacheManager,
		encryptionService: encryptionService,
	}
}

// GetProviderConfig retrieves a provider configuration with decryption
func (ps *ProviderService) GetProviderConfig(ctx context.Context, tenantID uint64, providerType interfaces.OAuthProviderType) (*interfaces.OAuthProviderConfig, error) {
	// Get from cache (may be encrypted)
	config, err := ps.cacheManager.GetProviderConfig(ctx, tenantID, providerType)
	if err != nil {
		return nil, err
	}

	// Decrypt sensitive fields if needed
	if err := ps.encryptionService.DecryptProviderConfig(config); err != nil {
		return nil, fmt.Errorf("failed to decrypt provider config: %v", err)
	}

	return config, nil
}

// GetProviderConfigByID retrieves a provider configuration by ID with decryption
func (ps *ProviderService) GetProviderConfigByID(ctx context.Context, providerID int64) (*interfaces.OAuthProviderConfig, error) {
	// Get from cache (may be encrypted)
	config, err := ps.cacheManager.GetProviderConfigByID(ctx, providerID)
	if err != nil {
		return nil, err
	}

	// Decrypt sensitive fields if needed
	if err := ps.encryptionService.DecryptProviderConfig(config); err != nil {
		return nil, fmt.Errorf("failed to decrypt provider config: %v", err)
	}

	return config, nil
}

// CreateProviderConfig creates a new provider configuration with encryption
func (ps *ProviderService) CreateProviderConfig(ctx context.Context, config *interfaces.OAuthProviderConfig) (*interfaces.OAuthProviderConfig, error) {
	// Make a copy to avoid modifying the original
	configCopy := *config

	// Encrypt sensitive fields
	if err := ps.encryptionService.UpdateProviderConfigWithEncryption(&configCopy); err != nil {
		return nil, fmt.Errorf("failed to encrypt provider config: %v", err)
	}

	// Create in database
	builder := ps.db.OauthProvider.Create().
		SetName(configCopy.Name).
		SetDisplayName(configCopy.DisplayName).
		SetType(string(configCopy.Type)).
		SetClientID(configCopy.ClientID).
		SetRedirectURL(configCopy.RedirectURL).
		SetAuthURL(configCopy.AuthURL).
		SetTokenURL(configCopy.TokenURL).
		SetInfoURL(configCopy.UserInfoURL).
		SetAuthStyle(int(configCopy.AuthStyle)).
		SetEnabled(configCopy.Enabled).
		SetSupportPkce(configCopy.SupportPKCE)

	// Set encrypted secret if available
	if configCopy.EncryptedSecret != "" {
		builder = builder.SetEncryptedSecret(configCopy.EncryptedSecret)
	}
	if configCopy.EncryptionKeyID != "" {
		builder = builder.SetEncryptionKeyID(configCopy.EncryptionKeyID)
	}

	// Set original client secret if no encryption was done
	if configCopy.ClientSecret != "" {
		builder = builder.SetClientSecret(configCopy.ClientSecret)
	}

	// Handle scopes
	if len(configCopy.Scopes) > 0 {
		// For now, join with commas - can be enhanced later
		scopesStr := ""
		for i, scope := range configCopy.Scopes {
			if i > 0 {
				scopesStr += ","
			}
			scopesStr += scope
		}
		builder = builder.SetScopes(scopesStr)
	}

	// Handle extra config
	if configCopy.ExtraConfig != nil {
		builder = builder.SetExtraConfig(configCopy.ExtraConfig)
	}

	provider, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %v", err)
	}

	// Convert back to config and return
	result := ps.cacheManager.ConvertEntToConfig(provider)
	
	// Decrypt for the response
	if err := ps.encryptionService.DecryptProviderConfig(result); err != nil {
		return nil, fmt.Errorf("failed to decrypt created provider config: %v", err)
	}

	return result, nil
}

// UpdateProviderConfig updates an existing provider configuration with encryption
func (ps *ProviderService) UpdateProviderConfig(ctx context.Context, providerID int64, config *interfaces.OAuthProviderConfig) (*interfaces.OAuthProviderConfig, error) {
	// Make a copy to avoid modifying the original
	configCopy := *config

	// Encrypt sensitive fields if they changed
	if configCopy.ClientSecret != "" {
		if err := ps.encryptionService.UpdateProviderConfigWithEncryption(&configCopy); err != nil {
			return nil, fmt.Errorf("failed to encrypt provider config: %v", err)
		}
	}

	// Update in database
	builder := ps.db.OauthProvider.UpdateOneID(uint64(providerID)).
		SetDisplayName(configCopy.DisplayName).
		SetRedirectURL(configCopy.RedirectURL).
		SetAuthURL(configCopy.AuthURL).
		SetTokenURL(configCopy.TokenURL).
		SetInfoURL(configCopy.UserInfoURL).
		SetAuthStyle(int(configCopy.AuthStyle)).
		SetEnabled(configCopy.Enabled).
		SetSupportPkce(configCopy.SupportPKCE)

	// Update encrypted secret if available
	if configCopy.EncryptedSecret != "" {
		builder = builder.SetEncryptedSecret(configCopy.EncryptedSecret)
	}
	if configCopy.EncryptionKeyID != "" {
		builder = builder.SetEncryptionKeyID(configCopy.EncryptionKeyID)
	}

	// Update original client secret if no encryption was done
	if configCopy.ClientSecret != "" && configCopy.EncryptedSecret == "" {
		builder = builder.SetClientSecret(configCopy.ClientSecret)
	}

	// Handle scopes
	if len(configCopy.Scopes) > 0 {
		scopesStr := ""
		for i, scope := range configCopy.Scopes {
			if i > 0 {
				scopesStr += ","
			}
			scopesStr += scope
		}
		builder = builder.SetScopes(scopesStr)
	}

	// Handle extra config
	if configCopy.ExtraConfig != nil {
		builder = builder.SetExtraConfig(configCopy.ExtraConfig)
	}

	provider, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update provider: %v", err)
	}

	// Invalidate cache
	ps.cacheManager.InvalidateProviderConfigByID(ctx, providerID)

	// Convert back to config
	result := ps.cacheManager.ConvertEntToConfig(provider)
	
	// Decrypt for the response
	if err := ps.encryptionService.DecryptProviderConfig(result); err != nil {
		return nil, fmt.Errorf("failed to decrypt updated provider config: %v", err)
	}

	return result, nil
}

// DeleteProviderConfig deletes a provider configuration
func (ps *ProviderService) DeleteProviderConfig(ctx context.Context, providerID int64) error {
	err := ps.db.OauthProvider.DeleteOneID(uint64(providerID)).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete provider: %v", err)
	}

	// Invalidate cache
	ps.cacheManager.InvalidateProviderConfigByID(ctx, providerID)

	return nil
}

// ListProviderConfigs lists provider configurations for a tenant
func (ps *ProviderService) ListProviderConfigs(ctx context.Context, tenantID uint64) ([]*interfaces.OAuthProviderConfig, error) {
	providers, err := ps.db.OauthProvider.Query().
		Where(
			// Using proper ent predicate imports would be needed here
			// For now, using basic query - this should be fixed in production
		).
		All(ctx)
	
	if err != nil {
		return nil, fmt.Errorf("failed to list providers: %v", err)
	}

	configs := make([]*interfaces.OAuthProviderConfig, 0, len(providers))
	for _, provider := range providers {
		config := ps.cacheManager.ConvertEntToConfig(provider)
		
		// Decrypt sensitive fields
		if err := ps.encryptionService.DecryptProviderConfig(config); err != nil {
			// Log error but don't fail the entire operation
			// TODO: Add proper logging
			continue
		}
		
		configs = append(configs, config)
	}

	return configs, nil
}

// ValidateProviderConfig validates a provider configuration
func (ps *ProviderService) ValidateProviderConfig(config *interfaces.OAuthProviderConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.Type == "" {
		return fmt.Errorf("provider type is required")
	}

	if config.ClientID == "" {
		return fmt.Errorf("client ID is required")
	}

	if config.ClientSecret == "" && config.EncryptedSecret == "" {
		return fmt.Errorf("client secret is required")
	}

	if config.RedirectURL == "" {
		return fmt.Errorf("redirect URL is required")
	}

	if config.AuthURL == "" {
		return fmt.Errorf("auth URL is required")
	}

	if config.TokenURL == "" {
		return fmt.Errorf("token URL is required")
	}

	if config.UserInfoURL == "" {
		return fmt.Errorf("user info URL is required")
	}

	return nil
}

// RotateProviderSecret rotates the client secret for a provider
func (ps *ProviderService) RotateProviderSecret(ctx context.Context, providerID int64, newSecret string) error {
	if newSecret == "" {
		return fmt.Errorf("new secret cannot be empty")
	}

	// Encrypt the new secret
	encryptedSecret, keyID, err := ps.encryptionService.EncryptProviderSecret(newSecret)
	if err != nil {
		return fmt.Errorf("failed to encrypt new secret: %v", err)
	}

	// Update in database
	_, err = ps.db.OauthProvider.UpdateOneID(uint64(providerID)).
		SetClientSecret(""). // Clear plain text
		SetEncryptedSecret(encryptedSecret).
		SetEncryptionKeyID(keyID).
		Save(ctx)
	
	if err != nil {
		return fmt.Errorf("failed to update provider secret: %v", err)
	}

	// Invalidate cache
	ps.cacheManager.InvalidateProviderConfigByID(ctx, providerID)

	return nil
}

// GetProviderStats returns statistics about providers
func (ps *ProviderService) GetProviderStats(ctx context.Context, tenantID uint64) (*ProviderStats, error) {
	// This would typically aggregate from the database
	// For now, return basic stats
	return &ProviderStats{
		TenantID:       tenantID,
		TotalProviders: 0,
		EnabledProviders: 0,
		CacheStats:     ps.cacheManager.GetCacheStats(),
		LastUpdated:    time.Now(),
	}, nil
}

// ProviderStats represents provider statistics
type ProviderStats struct {
	TenantID         uint64                `json:"tenant_id"`
	TotalProviders   int                   `json:"total_providers"`
	EnabledProviders int                   `json:"enabled_providers"`
	CacheStats       cache.CacheStats      `json:"cache_stats"`
	LastUpdated      time.Time             `json:"last_updated"`
}

// ConvertEntToConfig is a public method to convert ent entities
func (ps *ProviderService) ConvertEntToConfig(provider *ent.OauthProvider) *interfaces.OAuthProviderConfig {
	return ps.cacheManager.ConvertEntToConfig(provider)
}

// Global provider service instance
var globalProviderService *ProviderService

// InitGlobalProviderService initializes the global provider service
func InitGlobalProviderService(db *ent.Client, cacheManager *cache.CacheManager, encryptionService *encryption.ProviderEncryptionService) {
	globalProviderService = NewProviderService(db, cacheManager, encryptionService)
}

// GetGlobalProviderService returns the global provider service instance
func GetGlobalProviderService() *ProviderService {
	return globalProviderService
}

// SetGlobalProviderService sets the global provider service instance
func SetGlobalProviderService(service *ProviderService) {
	globalProviderService = service
}