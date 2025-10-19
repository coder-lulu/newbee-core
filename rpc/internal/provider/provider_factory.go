package provider

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/internal/encryption"
	"github.com/coder-lulu/newbee-core/rpc/internal/interfaces"
)

// ProviderFactory manages OAuth provider creation and configuration
type ProviderFactory struct {
	mu                sync.RWMutex
	db                *ent.Client
	adapterFactory    interfaces.OAuthAdapterFactory
	providerService   *ProviderService
	encryptionService *encryption.ProviderEncryptionService
	registeredTypes   map[interfaces.OAuthProviderType]*ProviderTemplate
}

// ProviderTemplate defines a template for creating OAuth providers
type ProviderTemplate struct {
	Type            interfaces.OAuthProviderType   `json:"type"`
	DisplayName     string                         `json:"display_name"`
	Description     string                         `json:"description"`
	DefaultConfig   *interfaces.OAuthProviderConfig `json:"default_config"`
	RequiredFields  []string                       `json:"required_fields"`
	OptionalFields  []string                       `json:"optional_fields"`
	SupportedScopes []string                       `json:"supported_scopes"`
	DefaultScopes   []string                       `json:"default_scopes"`
	CreatedAt       time.Time                      `json:"created_at"`
}

// ProviderRegistration represents a provider registration request
type ProviderRegistration struct {
	TenantID     uint64                         `json:"tenant_id"`
	Type         interfaces.OAuthProviderType   `json:"type"`
	Name         string                         `json:"name"`
	DisplayName  string                         `json:"display_name"`
	ClientID     string                         `json:"client_id"`
	ClientSecret string                         `json:"client_secret"`
	RedirectURL  string                         `json:"redirect_url"`
	Scopes       []string                       `json:"scopes,omitempty"`
	ExtraConfig  map[string]interface{}         `json:"extra_config,omitempty"`
	Enabled      bool                           `json:"enabled"`
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory(db *ent.Client, adapterFactory interfaces.OAuthAdapterFactory, providerService *ProviderService, encryptionService *encryption.ProviderEncryptionService) *ProviderFactory {
	factory := &ProviderFactory{
		db:                db,
		adapterFactory:    adapterFactory,
		providerService:   providerService,
		encryptionService: encryptionService,
		registeredTypes:   make(map[interfaces.OAuthProviderType]*ProviderTemplate),
	}

	// Register built-in provider templates
	factory.registerBuiltinTemplates()

	return factory
}

// RegisterProviderTemplate registers a new provider template
func (pf *ProviderFactory) RegisterProviderTemplate(template *ProviderTemplate) error {
	pf.mu.Lock()
	defer pf.mu.Unlock()

	if template == nil {
		return fmt.Errorf("template cannot be nil")
	}

	if template.Type == "" {
		return fmt.Errorf("provider type cannot be empty")
	}

	// Validate that the adapter exists
	if !pf.adapterFactory.IsProviderSupported(template.Type) {
		return fmt.Errorf("adapter for provider type %s is not registered", template.Type)
	}

	template.CreatedAt = time.Now()
	pf.registeredTypes[template.Type] = template

	return nil
}

// GetProviderTemplate returns a provider template by type
func (pf *ProviderFactory) GetProviderTemplate(providerType interfaces.OAuthProviderType) (*ProviderTemplate, error) {
	pf.mu.RLock()
	defer pf.mu.RUnlock()

	template, exists := pf.registeredTypes[providerType]
	if !exists {
		return nil, fmt.Errorf("provider template for type %s not found", providerType)
	}

	return template, nil
}

// CreateProvider creates a new OAuth provider from a registration
func (pf *ProviderFactory) CreateProvider(ctx context.Context, registration *ProviderRegistration) (*interfaces.OAuthProviderConfig, error) {
	if err := pf.validateRegistration(registration); err != nil {
		return nil, fmt.Errorf("registration validation failed: %v", err)
	}

	// Get provider template
	template, err := pf.GetProviderTemplate(registration.Type)
	if err != nil {
		return nil, err
	}

	// Build configuration from template and registration
	config := pf.buildConfigFromTemplate(template, registration)

	// Validate the configuration
	if err := pf.providerService.ValidateProviderConfig(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %v", err)
	}

	// Create the provider
	return pf.providerService.CreateProviderConfig(ctx, config)
}

// CreateProviderFromAdapter creates a provider using an existing adapter
func (pf *ProviderFactory) CreateProviderFromAdapter(ctx context.Context, tenantID uint64, providerType interfaces.OAuthProviderType, config *interfaces.OAuthProviderConfig) (interfaces.OAuthAdapter, error) {
	// Create the adapter
	adapter, err := pf.adapterFactory.CreateAdapter(providerType)
	if err != nil {
		return nil, fmt.Errorf("failed to create adapter: %v", err)
	}

	// Configure the adapter
	if err := adapter.Configure(config); err != nil {
		return nil, fmt.Errorf("failed to configure adapter: %v", err)
	}

	return adapter, nil
}

// GetConfiguredAdapter retrieves and configures an adapter for a provider
func (pf *ProviderFactory) GetConfiguredAdapter(ctx context.Context, tenantID uint64, providerType interfaces.OAuthProviderType) (interfaces.OAuthAdapter, error) {
	// Get provider configuration
	config, err := pf.providerService.GetProviderConfig(ctx, tenantID, providerType)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider config: %v", err)
	}

	if !config.Enabled {
		return nil, fmt.Errorf("provider %s is disabled for tenant %d", providerType, tenantID)
	}

	// Create and configure adapter
	return pf.CreateProviderFromAdapter(ctx, tenantID, providerType, config)
}

// ListRegisteredTypes returns all registered provider types
func (pf *ProviderFactory) ListRegisteredTypes() []interfaces.OAuthProviderType {
	pf.mu.RLock()
	defer pf.mu.RUnlock()

	types := make([]interfaces.OAuthProviderType, 0, len(pf.registeredTypes))
	for providerType := range pf.registeredTypes {
		types = append(types, providerType)
	}

	return types
}

// ListProviderTemplates returns all registered provider templates
func (pf *ProviderFactory) ListProviderTemplates() []*ProviderTemplate {
	pf.mu.RLock()
	defer pf.mu.RUnlock()

	templates := make([]*ProviderTemplate, 0, len(pf.registeredTypes))
	for _, template := range pf.registeredTypes {
		templates = append(templates, template)
	}

	return templates
}

// IsProviderTypeSupported checks if a provider type is supported
func (pf *ProviderFactory) IsProviderTypeSupported(providerType interfaces.OAuthProviderType) bool {
	pf.mu.RLock()
	defer pf.mu.RUnlock()

	_, exists := pf.registeredTypes[providerType]
	return exists
}

// validateRegistration validates a provider registration
func (pf *ProviderFactory) validateRegistration(registration *ProviderRegistration) error {
	if registration == nil {
		return fmt.Errorf("registration cannot be nil")
	}

	if registration.TenantID == 0 {
		return fmt.Errorf("tenant ID is required")
	}

	if registration.Type == "" {
		return fmt.Errorf("provider type is required")
	}

	if registration.Name == "" {
		return fmt.Errorf("provider name is required")
	}

	if registration.ClientID == "" {
		return fmt.Errorf("client ID is required")
	}

	if registration.ClientSecret == "" {
		return fmt.Errorf("client secret is required")
	}

	if registration.RedirectURL == "" {
		return fmt.Errorf("redirect URL is required")
	}

	// Check if provider type is supported
	if !pf.IsProviderTypeSupported(registration.Type) {
		return fmt.Errorf("provider type %s is not supported", registration.Type)
	}

	// Validate required fields based on template
	template, err := pf.GetProviderTemplate(registration.Type)
	if err != nil {
		return err
	}

	return pf.validateRequiredFields(registration, template)
}

// validateRequiredFields validates required fields based on template
func (pf *ProviderFactory) validateRequiredFields(registration *ProviderRegistration, template *ProviderTemplate) error {
	// This would validate specific required fields for each provider type
	// For now, basic validation is done in validateRegistration
	return nil
}

// buildConfigFromTemplate builds a provider config from template and registration
func (pf *ProviderFactory) buildConfigFromTemplate(template *ProviderTemplate, registration *ProviderRegistration) *interfaces.OAuthProviderConfig {
	config := &interfaces.OAuthProviderConfig{
		Name:         registration.Name,
		DisplayName:  registration.DisplayName,
		Type:         registration.Type,
		ClientID:     registration.ClientID,
		ClientSecret: registration.ClientSecret,
		RedirectURL:  registration.RedirectURL,
		Enabled:      registration.Enabled,
		ExtraConfig:  registration.ExtraConfig,
	}

	// Use default display name if not provided
	if config.DisplayName == "" {
		config.DisplayName = template.DisplayName
	}

	// Use provided scopes or default to template scopes
	if len(registration.Scopes) > 0 {
		config.Scopes = registration.Scopes
	} else {
		config.Scopes = template.DefaultScopes
	}

	// Copy default configuration from template
	if template.DefaultConfig != nil {
		if config.AuthURL == "" {
			config.AuthURL = template.DefaultConfig.AuthURL
		}
		if config.TokenURL == "" {
			config.TokenURL = template.DefaultConfig.TokenURL
		}
		if config.UserInfoURL == "" {
			config.UserInfoURL = template.DefaultConfig.UserInfoURL
		}
		config.AuthStyle = template.DefaultConfig.AuthStyle
		config.SupportPKCE = template.DefaultConfig.SupportPKCE
	}

	return config
}

// registerBuiltinTemplates registers the built-in provider templates
func (pf *ProviderFactory) registerBuiltinTemplates() {
	templates := []*ProviderTemplate{
		{
			Type:        interfaces.ProviderTypeGitHub,
			DisplayName: "GitHub",
			Description: "GitHub OAuth provider for developer authentication",
			DefaultConfig: &interfaces.OAuthProviderConfig{
				AuthURL:     "https://github.com/login/oauth/authorize",
				TokenURL:    "https://github.com/login/oauth/access_token",
				UserInfoURL: "https://api.github.com/user",
				AuthStyle:   2, // AuthStyleInParams
				SupportPKCE: true,
			},
			RequiredFields:  []string{"client_id", "client_secret", "redirect_url"},
			OptionalFields:  []string{"scopes"},
			SupportedScopes: []string{"user", "user:email", "read:user", "repo", "public_repo"},
			DefaultScopes:   []string{"user:email"},
		},
		{
			Type:        interfaces.ProviderTypeGoogle,
			DisplayName: "Google",
			Description: "Google OAuth provider for Gmail and Google services",
			DefaultConfig: &interfaces.OAuthProviderConfig{
				AuthURL:     "https://accounts.google.com/o/oauth2/auth",
				TokenURL:    "https://oauth2.googleapis.com/token",
				UserInfoURL: "https://www.googleapis.com/oauth2/v2/userinfo",
				AuthStyle:   2, // AuthStyleInParams
				SupportPKCE: true,
			},
			RequiredFields:  []string{"client_id", "client_secret", "redirect_url"},
			OptionalFields:  []string{"scopes"},
			SupportedScopes: []string{"openid", "profile", "email"},
			DefaultScopes:   []string{"openid", "profile", "email"},
		},
		{
			Type:        interfaces.ProviderTypeFacebook,
			DisplayName: "Facebook",
			Description: "Facebook OAuth provider for social authentication",
			DefaultConfig: &interfaces.OAuthProviderConfig{
				AuthURL:     "https://www.facebook.com/v18.0/dialog/oauth",
				TokenURL:    "https://graph.facebook.com/v18.0/oauth/access_token",
				UserInfoURL: "https://graph.facebook.com/v18.0/me",
				AuthStyle:   2, // AuthStyleInParams
				SupportPKCE: true,
			},
			RequiredFields:  []string{"client_id", "client_secret", "redirect_url"},
			OptionalFields:  []string{"scopes"},
			SupportedScopes: []string{"email", "public_profile"},
			DefaultScopes:   []string{"email", "public_profile"},
		},
		{
			Type:        interfaces.ProviderTypeWechat,
			DisplayName: "微信",
			Description: "WeChat OAuth provider for Chinese users",
			DefaultConfig: &interfaces.OAuthProviderConfig{
				AuthURL:     "https://open.weixin.qq.com/connect/qrconnect",
				TokenURL:    "https://api.weixin.qq.com/sns/oauth2/access_token",
				UserInfoURL: "https://api.weixin.qq.com/sns/userinfo",
				AuthStyle:   1, // AuthStyleInHeader
				SupportPKCE: false,
			},
			RequiredFields:  []string{"client_id", "client_secret", "redirect_url"},
			OptionalFields:  []string{"scopes"},
			SupportedScopes: []string{"snsapi_login"},
			DefaultScopes:   []string{"snsapi_login"},
		},
		{
			Type:        interfaces.ProviderTypeQQ,
			DisplayName: "QQ",
			Description: "QQ OAuth provider for Chinese users",
			DefaultConfig: &interfaces.OAuthProviderConfig{
				AuthURL:     "https://graph.qq.com/oauth2.0/authorize",
				TokenURL:    "https://graph.qq.com/oauth2.0/token",
				UserInfoURL: "https://graph.qq.com/user/get_user_info",
				AuthStyle:   1, // AuthStyleInHeader
				SupportPKCE: false,
			},
			RequiredFields:  []string{"client_id", "client_secret", "redirect_url"},
			OptionalFields:  []string{"scopes"},
			SupportedScopes: []string{"get_user_info"},
			DefaultScopes:   []string{"get_user_info"},
		},
	}

	for _, template := range templates {
		if err := pf.RegisterProviderTemplate(template); err != nil {
			// Log error but continue with other templates
			// TODO: Add proper logging
		}
	}
}

// Global provider factory instance
var globalProviderFactory *ProviderFactory

// InitGlobalProviderFactory initializes the global provider factory
func InitGlobalProviderFactory(db *ent.Client, adapterFactory interfaces.OAuthAdapterFactory, providerService *ProviderService, encryptionService *encryption.ProviderEncryptionService) {
	globalProviderFactory = NewProviderFactory(db, adapterFactory, providerService, encryptionService)
}

// GetGlobalProviderFactory returns the global provider factory instance
func GetGlobalProviderFactory() *ProviderFactory {
	return globalProviderFactory
}

// SetGlobalProviderFactory sets the global provider factory instance
func SetGlobalProviderFactory(factory *ProviderFactory) {
	globalProviderFactory = factory
}