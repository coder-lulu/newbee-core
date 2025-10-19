package oauth

import (
	"context"
	"time"

	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/internal/adapters"
	"github.com/coder-lulu/newbee-core/rpc/internal/cache"
	"github.com/coder-lulu/newbee-core/rpc/internal/encryption"
	"github.com/coder-lulu/newbee-core/rpc/internal/health"
	"github.com/coder-lulu/newbee-core/rpc/internal/interfaces"
	"github.com/coder-lulu/newbee-core/rpc/internal/provider"
	"github.com/coder-lulu/newbee-core/rpc/internal/validation"
)

// OAuthManager provides a unified interface for all OAuth-related operations
type OAuthManager struct {
	// Core components
	db                    *ent.Client
	adapterFactory        interfaces.OAuthAdapterFactory
	providerService       *provider.ProviderService
	providerFactory       *provider.ProviderFactory
	providerRegistry      *provider.ProviderRegistry
	
	// Support services
	cacheManager          *cache.CacheManager
	encryptionService     *encryption.ProviderEncryptionService
	validator             *validation.ProviderValidator
	healthChecker         *health.HealthChecker
	systemHealthChecker   *health.SystemHealthChecker
	
	// Configuration
	initialized           bool
}

// OAuthManagerConfig represents configuration for the OAuth manager
type OAuthManagerConfig struct {
	// Cache configuration
	CacheTTL     time.Duration
	CacheMaxSize int
	
	// Registry configuration
	ProviderMaxIdleTime time.Duration
	
	// Health check configuration
	HealthCheckInterval time.Duration
	
	// Encryption configuration
	DefaultEncryptionKey []byte
	EncryptionKeyID      string
}

// DefaultOAuthManagerConfig returns default configuration
func DefaultOAuthManagerConfig() *OAuthManagerConfig {
	return &OAuthManagerConfig{
		CacheTTL:             time.Hour,
		CacheMaxSize:         1000,
		ProviderMaxIdleTime:  30 * time.Minute,
		HealthCheckInterval:  5 * time.Minute,
	}
}

// NewOAuthManager creates a new OAuth manager with all integrated components
func NewOAuthManager(db *ent.Client, config *OAuthManagerConfig) *OAuthManager {
	if config == nil {
		config = DefaultOAuthManagerConfig()
	}

	manager := &OAuthManager{
		db: db,
	}

	// Initialize components in correct order
	manager.initializeComponents(config)

	return manager
}

// initializeComponents initializes all OAuth components
func (om *OAuthManager) initializeComponents(config *OAuthManagerConfig) {
	// 1. Initialize encryption
	encryption.InitGlobalEncryption()
	om.encryptionService = encryption.GetGlobalProviderEncryptionService()
	
	// Add default encryption key if provided
	if len(config.DefaultEncryptionKey) > 0 && config.EncryptionKeyID != "" {
		encManager := encryption.GetGlobalEncryptionManager()
		encManager.AddKey(config.EncryptionKeyID, config.DefaultEncryptionKey, encryption.AlgorithmAES256GCM)
		encManager.SetActiveKey(config.EncryptionKeyID)
	}

	// 2. Initialize cache
	cache.InitGlobalCacheManager(om.db, config.CacheTTL, config.CacheMaxSize)
	om.cacheManager = cache.GetGlobalCacheManager()

	// 3. Initialize adapters
	om.adapterFactory = adapters.GetGlobalAdapterFactory()

	// 4. Initialize provider service
	provider.InitGlobalProviderService(om.db, om.cacheManager, om.encryptionService)
	om.providerService = provider.GetGlobalProviderService()

	// 5. Initialize provider factory
	provider.InitGlobalProviderFactory(om.db, om.adapterFactory, om.providerService, om.encryptionService)
	om.providerFactory = provider.GetGlobalProviderFactory()

	// 6. Initialize provider registry
	provider.InitGlobalProviderRegistry(om.providerFactory, config.ProviderMaxIdleTime)
	om.providerRegistry = provider.GetGlobalProviderRegistry()

	// 7. Initialize validation
	om.validator = validation.GetGlobalProviderValidator()

	// 8. Initialize health checking
	health.InitGlobalHealthChecker(om.providerService, om.validator, config.HealthCheckInterval)
	om.healthChecker = health.GetGlobalHealthChecker()
	om.systemHealthChecker = health.GetGlobalSystemHealthChecker()

	// 9. Start background services
	cache.InitGlobalCleanupService(om.cacheManager, config.CacheTTL/2)
	cache.StartGlobalCleanupService()

	om.initialized = true
}

// GetProvider retrieves a configured OAuth provider adapter
func (om *OAuthManager) GetProvider(ctx context.Context, tenantID uint64, providerType interfaces.OAuthProviderType) (interfaces.OAuthAdapter, error) {
	return om.providerRegistry.GetProvider(ctx, tenantID, providerType)
}

// CreateProvider creates a new OAuth provider
func (om *OAuthManager) CreateProvider(ctx context.Context, registration *provider.ProviderRegistration) (*interfaces.OAuthProviderConfig, error) {
	return om.providerFactory.CreateProvider(ctx, registration)
}

// GetProviderConfig retrieves a provider configuration
func (om *OAuthManager) GetProviderConfig(ctx context.Context, tenantID uint64, providerType interfaces.OAuthProviderType) (*interfaces.OAuthProviderConfig, error) {
	return om.providerService.GetProviderConfig(ctx, tenantID, providerType)
}

// UpdateProviderConfig updates a provider configuration
func (om *OAuthManager) UpdateProviderConfig(ctx context.Context, providerID int64, config *interfaces.OAuthProviderConfig) (*interfaces.OAuthProviderConfig, error) {
	// Update the configuration
	result, err := om.providerService.UpdateProviderConfig(ctx, providerID, config)
	if err != nil {
		return nil, err
	}

	// Note: We would need tenant ID to properly invalidate the provider
	// For now, we rely on cache TTL to refresh the configuration

	return result, nil
}

// DeleteProvider deletes a provider
func (om *OAuthManager) DeleteProvider(ctx context.Context, providerID int64) error {
	return om.providerService.DeleteProviderConfig(ctx, providerID)
}

// ValidateProvider validates a provider configuration
func (om *OAuthManager) ValidateProvider(ctx context.Context, config *interfaces.OAuthProviderConfig, level validation.ValidationLevel) *validation.ValidationResult {
	return om.validator.ValidateProvider(ctx, config, level)
}

// CheckProviderHealth performs a health check on a provider
func (om *OAuthManager) CheckProviderHealth(ctx context.Context, tenantID uint64, providerType interfaces.OAuthProviderType) *health.ProviderHealthCheck {
	return om.healthChecker.CheckProviderHealth(ctx, tenantID, providerType)
}

// GetHealthSummary returns a health summary of all providers
func (om *OAuthManager) GetHealthSummary() *health.HealthSummary {
	return om.healthChecker.GetHealthSummary()
}

// CheckSystemHealth performs a comprehensive system health check
func (om *OAuthManager) CheckSystemHealth(ctx context.Context) map[string]*health.HealthCheck {
	return om.systemHealthChecker.CheckSystemHealth(ctx)
}

// GetCacheStats returns cache performance statistics
func (om *OAuthManager) GetCacheStats() cache.CacheStats {
	return om.cacheManager.GetCacheStats()
}

// GetRegistryStats returns provider registry statistics
func (om *OAuthManager) GetRegistryStats() provider.RegistryStats {
	return om.providerRegistry.GetStats()
}

// ListProviderTypes returns all supported provider types
func (om *OAuthManager) ListProviderTypes() []interfaces.OAuthProviderType {
	return om.providerFactory.ListRegisteredTypes()
}

// ListProviderTemplates returns all available provider templates
func (om *OAuthManager) ListProviderTemplates() []*provider.ProviderTemplate {
	return om.providerFactory.ListProviderTemplates()
}

// RefreshProvider forces a refresh of a cached provider
func (om *OAuthManager) RefreshProvider(ctx context.Context, tenantID uint64, providerType interfaces.OAuthProviderType) (interfaces.OAuthAdapter, error) {
	return om.providerRegistry.RefreshProvider(ctx, tenantID, providerType)
}

// WarmupCache pre-loads provider configurations for specified tenants
func (om *OAuthManager) WarmupCache(ctx context.Context, tenantIDs []uint64) error {
	return om.cacheManager.WarmupCache(ctx, tenantIDs)
}

// StartPeriodicHealthChecks starts periodic health checking for specified tenants
func (om *OAuthManager) StartPeriodicHealthChecks(ctx context.Context, tenantIDs []uint64) {
	om.healthChecker.StartPeriodicHealthChecks(ctx, tenantIDs)
}

// StopPeriodicHealthChecks stops periodic health checking
func (om *OAuthManager) StopPeriodicHealthChecks() {
	om.healthChecker.StopPeriodicHealthChecks()
}

// RotateProviderSecret rotates the client secret for a provider
func (om *OAuthManager) RotateProviderSecret(ctx context.Context, providerID int64, newSecret string) error {
	return om.providerService.RotateProviderSecret(ctx, providerID, newSecret)
}

// IsInitialized returns whether the OAuth manager is properly initialized
func (om *OAuthManager) IsInitialized() bool {
	return om.initialized
}

// Shutdown gracefully shuts down the OAuth manager and all its components
func (om *OAuthManager) Shutdown() {
	if !om.initialized {
		return
	}

	// Stop background services
	om.healthChecker.StopPeriodicHealthChecks()
	cache.StopGlobalCleanupService()
	
	// Shutdown components
	om.providerRegistry.Shutdown()
	
	om.initialized = false
}

// Global OAuth manager instance
var globalOAuthManager *OAuthManager

// InitGlobalOAuthManager initializes the global OAuth manager
func InitGlobalOAuthManager(db *ent.Client, config *OAuthManagerConfig) {
	globalOAuthManager = NewOAuthManager(db, config)
}

// GetGlobalOAuthManager returns the global OAuth manager instance
func GetGlobalOAuthManager() *OAuthManager {
	return globalOAuthManager
}

// SetGlobalOAuthManager sets the global OAuth manager instance
func SetGlobalOAuthManager(manager *OAuthManager) {
	globalOAuthManager = manager
}