package adapters

import (
	"context"

	"golang.org/x/oauth2"

	"github.com/coder-lulu/newbee-core/rpc/internal/interfaces"
)

// FeishuAdapter implements OAuth adapter for Feishu
type FeishuAdapter struct {
	*BaseOAuthAdapter
}

// NewFeishuAdapter creates a new Feishu OAuth adapter
func NewFeishuAdapter() interfaces.OAuthAdapter {
	return &FeishuAdapter{
		BaseOAuthAdapter: NewBaseOAuthAdapter(),
	}
}

// GetProviderType returns the provider type
func (f *FeishuAdapter) GetProviderType() interfaces.OAuthProviderType {
	return interfaces.ProviderTypeFeishu
}

// GetProviderName returns the human-readable name of this provider
func (f *FeishuAdapter) GetProviderName() string {
	return "飞书"
}

// GetDefaultScopes returns the recommended default scopes for this provider
func (f *FeishuAdapter) GetDefaultScopes() []string {
	return []string{}
}

// SupportsFeature checks if the provider supports a specific feature
func (f *FeishuAdapter) SupportsFeature(feature interfaces.OAuthFeature) bool {
	switch feature {
	case interfaces.FeaturePKCE:
		return false
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
func (f *FeishuAdapter) ValidateToken(ctx context.Context, token *oauth2.Token) (bool, error) {
	return f.BaseOAuthAdapter.ValidateToken(ctx, token)
}

// GetAuthorizationURL generates the OAuth authorization URL with optional PKCE
func (f *FeishuAdapter) GetAuthorizationURL(ctx context.Context, session *interfaces.OAuthSession) (string, error) {
	return f.BaseOAuthAdapter.GetAuthorizationURL(ctx, session)
}

// ExchangeCodeForToken exchanges authorization code for access token
func (f *FeishuAdapter) ExchangeCodeForToken(ctx context.Context, code string, session *interfaces.OAuthSession) (*oauth2.Token, error) {
	return f.BaseOAuthAdapter.ExchangeCodeForToken(ctx, code, session)
}

// Configure sets up the adapter with provider configuration
func (f *FeishuAdapter) Configure(config *interfaces.OAuthProviderConfig) error {
	if config.Type != interfaces.ProviderTypeFeishu {
		return &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeConfigurationError,
			Description: "invalid provider type for Feishu adapter",
		}
	}

	return f.BaseOAuthAdapter.Configure(config)
}

// GetUserInfo retrieves user information using the access token
func (f *FeishuAdapter) GetUserInfo(ctx context.Context, token *oauth2.Token) (*interfaces.OAuthUserInfo, error) {
	// TODO: Implement Feishu user info retrieval
	return nil, &interfaces.OAuthError{
		Type:        interfaces.ErrorTypeUnsupportedResponse,
		Description: "Feishu adapter not fully implemented yet",
	}
}

// GetSupportedScopes returns the scopes supported by Feishu
func (f *FeishuAdapter) GetSupportedScopes() []string {
	return []string{}
}

// ValidateConfig validates the Feishu-specific configuration
func (f *FeishuAdapter) ValidateConfig(config *interfaces.OAuthProviderConfig) error {
	if config.AuthURL == "" {
		return &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeConfigurationError,
			Description: "auth_url is required for Feishu",
		}
	}

	if config.TokenURL == "" {
		return &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeConfigurationError,
			Description: "token_url is required for Feishu",
		}
	}

	if config.UserInfoURL == "" {
		return &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeConfigurationError,
			Description: "user_info_url is required for Feishu",
		}
	}

	return nil
}

// RefreshToken refreshes an expired access token using the refresh token
func (f *FeishuAdapter) RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	return f.BaseOAuthAdapter.RefreshToken(ctx, refreshToken)
}

// RevokeToken revokes an access token
func (f *FeishuAdapter) RevokeToken(ctx context.Context, token *oauth2.Token) error {
	return f.BaseOAuthAdapter.RevokeToken(ctx, token)
}