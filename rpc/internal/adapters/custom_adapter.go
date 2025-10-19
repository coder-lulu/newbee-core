package adapters

import (
	"context"

	"golang.org/x/oauth2"

	"github.com/coder-lulu/newbee-core/rpc/internal/interfaces"
)

// CustomAdapter implements OAuth adapter for generic OAuth2 providers
type CustomAdapter struct {
	*BaseOAuthAdapter
}

// NewCustomAdapter creates a new Custom OAuth adapter
func NewCustomAdapter() interfaces.OAuthAdapter {
	return &CustomAdapter{
		BaseOAuthAdapter: NewBaseOAuthAdapter(),
	}
}

// GetProviderType returns the provider type
func (c *CustomAdapter) GetProviderType() interfaces.OAuthProviderType {
	return interfaces.ProviderTypeCustom
}

// GetProviderName returns the human-readable name of this provider
func (c *CustomAdapter) GetProviderName() string {
	return "Custom OAuth2"
}

// GetDefaultScopes returns the recommended default scopes for this provider
func (c *CustomAdapter) GetDefaultScopes() []string {
	return []string{}
}

// SupportsFeature checks if the provider supports a specific feature
func (c *CustomAdapter) SupportsFeature(feature interfaces.OAuthFeature) bool {
	switch feature {
	case interfaces.FeaturePKCE:
		return true
	case interfaces.FeatureRefreshToken:
		return true
	case interfaces.FeatureTokenRevoke:
		return false
	case interfaces.FeatureUserInfo:
		return true
	default:
		return false
	}
}

// ValidateToken validates if a token is still valid
func (c *CustomAdapter) ValidateToken(ctx context.Context, token *oauth2.Token) (bool, error) {
	return c.BaseOAuthAdapter.ValidateToken(ctx, token)
}

// GetAuthorizationURL generates the OAuth authorization URL with optional PKCE
func (c *CustomAdapter) GetAuthorizationURL(ctx context.Context, session *interfaces.OAuthSession) (string, error) {
	return c.BaseOAuthAdapter.GetAuthorizationURL(ctx, session)
}

// ExchangeCodeForToken exchanges authorization code for access token
func (c *CustomAdapter) ExchangeCodeForToken(ctx context.Context, code string, session *interfaces.OAuthSession) (*oauth2.Token, error) {
	return c.BaseOAuthAdapter.ExchangeCodeForToken(ctx, code, session)
}

// Configure sets up the adapter with provider configuration
func (c *CustomAdapter) Configure(config *interfaces.OAuthProviderConfig) error {
	if config.Type != interfaces.ProviderTypeCustom {
		return &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeConfigurationError,
			Description: "invalid provider type for Custom adapter",
		}
	}

	return c.BaseOAuthAdapter.Configure(config)
}

// GetUserInfo retrieves user information using the access token
func (c *CustomAdapter) GetUserInfo(ctx context.Context, token *oauth2.Token) (*interfaces.OAuthUserInfo, error) {
	// TODO: Implement generic OAuth2 user info retrieval
	return nil, &interfaces.OAuthError{
		Type:        interfaces.ErrorTypeUnsupportedResponse,
		Description: "Custom adapter requires specific implementation for user info endpoint",
	}
}

// GetSupportedScopes returns the scopes supported by the custom provider
func (c *CustomAdapter) GetSupportedScopes() []string {
	// Custom providers can define their own scopes
	return []string{}
}

// ValidateConfig validates the custom provider configuration
func (c *CustomAdapter) ValidateConfig(config *interfaces.OAuthProviderConfig) error {
	if config.AuthURL == "" {
		return &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeConfigurationError,
			Description: "auth_url is required for custom provider",
		}
	}

	if config.TokenURL == "" {
		return &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeConfigurationError,
			Description: "token_url is required for custom provider",
		}
	}

	if config.UserInfoURL == "" {
		return &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeConfigurationError,
			Description: "user_info_url is required for custom provider",
		}
	}

	return nil
}

// RefreshToken refreshes an expired access token using the refresh token
func (c *CustomAdapter) RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	return c.BaseOAuthAdapter.RefreshToken(ctx, refreshToken)
}

// RevokeToken revokes an access token
func (c *CustomAdapter) RevokeToken(ctx context.Context, token *oauth2.Token) error {
	return c.BaseOAuthAdapter.RevokeToken(ctx, token)
}