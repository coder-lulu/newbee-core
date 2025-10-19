package adapters

import (
	"fmt"
	"sync"

	"github.com/coder-lulu/newbee-core/rpc/internal/interfaces"
)

// DefaultOAuthAdapterFactory is the default implementation of OAuthAdapterFactory
type DefaultOAuthAdapterFactory struct {
	mu       sync.RWMutex
	adapters map[interfaces.OAuthProviderType]func() interfaces.OAuthAdapter
}

// NewOAuthAdapterFactory creates a new OAuth adapter factory
func NewOAuthAdapterFactory() interfaces.OAuthAdapterFactory {
	factory := &DefaultOAuthAdapterFactory{
		adapters: make(map[interfaces.OAuthProviderType]func() interfaces.OAuthAdapter),
	}
	
	// Register built-in adapters
	factory.registerBuiltinAdapters()
	
	return factory
}

// CreateAdapter creates an OAuth adapter for the specified provider type
func (f *DefaultOAuthAdapterFactory) CreateAdapter(providerType interfaces.OAuthProviderType) (interfaces.OAuthAdapter, error) {
	f.mu.RLock()
	adapterFunc, exists := f.adapters[providerType]
	f.mu.RUnlock()

	if !exists {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeConfigurationError,
			Description: fmt.Sprintf("unsupported provider type: %s", providerType),
		}
	}

	adapter := adapterFunc()
	if adapter == nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeConfigurationError,
			Description: fmt.Sprintf("failed to create adapter for provider: %s", providerType),
		}
	}

	return adapter, nil
}

// RegisterAdapter registers a new adapter implementation
func (f *DefaultOAuthAdapterFactory) RegisterAdapter(providerType interfaces.OAuthProviderType, adapterFunc func() interfaces.OAuthAdapter) error {
	if adapterFunc == nil {
		return &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeConfigurationError,
			Description: "adapter function cannot be nil",
		}
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.adapters[providerType] = adapterFunc
	return nil
}

// GetSupportedProviders returns a list of all supported provider types
func (f *DefaultOAuthAdapterFactory) GetSupportedProviders() []interfaces.OAuthProviderType {
	f.mu.RLock()
	defer f.mu.RUnlock()

	providers := make([]interfaces.OAuthProviderType, 0, len(f.adapters))
	for providerType := range f.adapters {
		providers = append(providers, providerType)
	}

	return providers
}

// IsProviderSupported checks if a provider type is supported
func (f *DefaultOAuthAdapterFactory) IsProviderSupported(providerType interfaces.OAuthProviderType) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	_, exists := f.adapters[providerType]
	return exists
}

// registerBuiltinAdapters registers the built-in OAuth adapters
func (f *DefaultOAuthAdapterFactory) registerBuiltinAdapters() {
	// Register GitHub adapter
	f.adapters[interfaces.ProviderTypeGitHub] = func() interfaces.OAuthAdapter {
		return NewGitHubAdapter()
	}

	// Register WeChat adapter
	f.adapters[interfaces.ProviderTypeWechat] = func() interfaces.OAuthAdapter {
		return NewWechatAdapter()
	}

	// Register QQ adapter
	f.adapters[interfaces.ProviderTypeQQ] = func() interfaces.OAuthAdapter {
		return NewQQAdapter()
	}

	// Register Google adapter
	f.adapters[interfaces.ProviderTypeGoogle] = func() interfaces.OAuthAdapter {
		return NewGoogleAdapter()
	}

	// Register Facebook adapter
	f.adapters[interfaces.ProviderTypeFacebook] = func() interfaces.OAuthAdapter {
		return NewFacebookAdapter()
	}

	// Register Feishu adapter
	f.adapters[interfaces.ProviderTypeFeishu] = func() interfaces.OAuthAdapter {
		return NewFeishuAdapter()
	}

	// Register Custom adapter (for generic OAuth2 providers)
	f.adapters[interfaces.ProviderTypeCustom] = func() interfaces.OAuthAdapter {
		return NewCustomAdapter()
	}
}

// GetAdapterConfiguration returns the default configuration for a provider type
func GetAdapterConfiguration(providerType interfaces.OAuthProviderType) *interfaces.OAuthProviderConfig {
	switch providerType {
	case interfaces.ProviderTypeGitHub:
		return &interfaces.OAuthProviderConfig{
			Type:        interfaces.ProviderTypeGitHub,
			DisplayName: "GitHub",
			AuthURL:     "https://github.com/login/oauth/authorize",
			TokenURL:    "https://github.com/login/oauth/access_token",
			UserInfoURL: "https://api.github.com/user",
			Scopes:      []string{"user:email"},
			AuthStyle:   2, // AuthStyleInParams
			SupportPKCE: true,
			Enabled:     true,
		}

	case interfaces.ProviderTypeGoogle:
		return &interfaces.OAuthProviderConfig{
			Type:        interfaces.ProviderTypeGoogle,
			DisplayName: "Google",
			AuthURL:     "https://accounts.google.com/o/oauth2/auth",
			TokenURL:    "https://oauth2.googleapis.com/token",
			UserInfoURL: "https://www.googleapis.com/oauth2/v2/userinfo",
			Scopes:      []string{"openid", "profile", "email"},
			AuthStyle:   2, // AuthStyleInParams
			SupportPKCE: true,
			Enabled:     true,
		}

	case interfaces.ProviderTypeFacebook:
		return &interfaces.OAuthProviderConfig{
			Type:        interfaces.ProviderTypeFacebook,
			DisplayName: "Facebook",
			AuthURL:     "https://www.facebook.com/v18.0/dialog/oauth",
			TokenURL:    "https://graph.facebook.com/v18.0/oauth/access_token",
			UserInfoURL: "https://graph.facebook.com/v18.0/me",
			Scopes:      []string{"email", "public_profile"},
			AuthStyle:   2, // AuthStyleInParams
			SupportPKCE: true,
			Enabled:     true,
		}

	case interfaces.ProviderTypeWechat:
		return &interfaces.OAuthProviderConfig{
			Type:        interfaces.ProviderTypeWechat,
			DisplayName: "微信",
			AuthURL:     "https://open.weixin.qq.com/connect/qrconnect",
			TokenURL:    "https://api.weixin.qq.com/sns/oauth2/access_token",
			UserInfoURL: "https://api.weixin.qq.com/sns/userinfo",
			Scopes:      []string{"snsapi_login"},
			AuthStyle:   1, // AuthStyleInHeader
			SupportPKCE: false,
			Enabled:     true,
		}

	case interfaces.ProviderTypeQQ:
		return &interfaces.OAuthProviderConfig{
			Type:        interfaces.ProviderTypeQQ,
			DisplayName: "QQ",
			AuthURL:     "https://graph.qq.com/oauth2.0/authorize",
			TokenURL:    "https://graph.qq.com/oauth2.0/token",
			UserInfoURL: "https://graph.qq.com/user/get_user_info",
			Scopes:      []string{"get_user_info"},
			AuthStyle:   1, // AuthStyleInHeader
			SupportPKCE: false,
			Enabled:     true,
		}

	case interfaces.ProviderTypeFeishu:
		return &interfaces.OAuthProviderConfig{
			Type:        interfaces.ProviderTypeFeishu,
			DisplayName: "飞书",
			AuthURL:     "https://open.feishu.cn/open-apis/authen/v1/index",
			TokenURL:    "https://open.feishu.cn/open-apis/authen/v1/access_token",
			UserInfoURL: "https://open.feishu.cn/open-apis/authen/v1/user_info",
			Scopes:      []string{},
			AuthStyle:   2, // AuthStyleInParams
			SupportPKCE: false,
			Enabled:     true,
		}

	default:
		return &interfaces.OAuthProviderConfig{
			Type:        interfaces.ProviderTypeCustom,
			DisplayName: "Custom OAuth2",
			AuthStyle:   2, // AuthStyleInParams
			SupportPKCE: true,
			Enabled:     true,
		}
	}
}

// ValidateProviderConfig validates a provider configuration against its type
func ValidateProviderConfig(config *interfaces.OAuthProviderConfig) error {
	if config == nil {
		return &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeConfigurationError,
			Description: "configuration cannot be nil",
		}
	}

	// Get default config for comparison
	defaultConfig := GetAdapterConfiguration(config.Type)
	
	// Validate required URLs are set
	if config.AuthURL == "" {
		config.AuthURL = defaultConfig.AuthURL
	}
	if config.TokenURL == "" {
		config.TokenURL = defaultConfig.TokenURL
	}
	if config.UserInfoURL == "" {
		config.UserInfoURL = defaultConfig.UserInfoURL
	}

	// Validate scopes
	if len(config.Scopes) == 0 && len(defaultConfig.Scopes) > 0 {
		config.Scopes = defaultConfig.Scopes
	}

	// Set display name if not provided
	if config.DisplayName == "" {
		config.DisplayName = defaultConfig.DisplayName
	}

	return nil
}

// Global adapter factory instance
var globalFactory interfaces.OAuthAdapterFactory

// GetGlobalAdapterFactory returns the global adapter factory instance
func GetGlobalAdapterFactory() interfaces.OAuthAdapterFactory {
	if globalFactory == nil {
		globalFactory = NewOAuthAdapterFactory()
	}
	return globalFactory
}

// SetGlobalAdapterFactory sets the global adapter factory instance
func SetGlobalAdapterFactory(factory interfaces.OAuthAdapterFactory) {
	globalFactory = factory
}