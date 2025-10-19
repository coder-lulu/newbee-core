package interfaces

import (
	"context"
	"time"

	"golang.org/x/oauth2"
)

// OAuthProviderType defines the type of OAuth provider
type OAuthProviderType string

const (
	ProviderTypeGitHub   OAuthProviderType = "github"
	ProviderTypeWechat   OAuthProviderType = "wechat"
	ProviderTypeQQ       OAuthProviderType = "qq"
	ProviderTypeGoogle   OAuthProviderType = "google"
	ProviderTypeFacebook OAuthProviderType = "facebook"
	ProviderTypeFeishu   OAuthProviderType = "feishu"
	ProviderTypeCustom   OAuthProviderType = "custom"
)

// OAuthUserInfo represents standardized user information from OAuth providers
type OAuthUserInfo struct {
	// Standard fields across all providers
	ID          string `json:"id"`           // Provider's unique user ID
	Username    string `json:"username"`     // Username/login handle
	Nickname    string `json:"nickname"`     // Display name
	Email       string `json:"email"`        // Email address
	Avatar      string `json:"avatar"`       // Avatar/profile picture URL
	PhoneNumber string `json:"phone_number"` // Phone number (if available)
	
	// Provider-specific metadata
	ProviderType OAuthProviderType          `json:"provider_type"`
	RawData      map[string]interface{} `json:"raw_data"` // Original response from provider
	
	// Optional fields
	FirstName    string `json:"first_name,omitempty"`
	LastName     string `json:"last_name,omitempty"`
	Location     string `json:"location,omitempty"`
	Company      string `json:"company,omitempty"`
	Website      string `json:"website,omitempty"`
	Bio          string `json:"bio,omitempty"`
	Verified     bool   `json:"verified,omitempty"`
	
	// Timestamps
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// OAuthProviderConfig represents OAuth provider configuration
type OAuthProviderConfig struct {
	Name         string            `json:"name"`
	DisplayName  string            `json:"display_name"`
	Type         OAuthProviderType `json:"type"`
	ClientID     string            `json:"client_id"`
	ClientSecret string            `json:"client_secret"`
	RedirectURL  string            `json:"redirect_url"`
	Scopes       []string          `json:"scopes"`
	
	// OAuth2 endpoints
	AuthURL     string `json:"auth_url"`
	TokenURL    string `json:"token_url"`
	UserInfoURL string `json:"user_info_url"`
	
	// Additional configuration
	ExtraConfig map[string]interface{} `json:"extra_config,omitempty"`
	
	// PKCE support
	SupportPKCE bool `json:"support_pkce"`
	
	// Advanced settings
	AuthStyle oauth2.AuthStyle `json:"auth_style"`
	Enabled   bool              `json:"enabled"`
	
	// Encryption fields
	EncryptedSecret  string `json:"encrypted_secret,omitempty"`
	EncryptionKeyID  string `json:"encryption_key_id,omitempty"`
}

// OAuthSession represents an ongoing OAuth authentication session
type OAuthSession struct {
	SessionID    string            `json:"session_id"`
	State        string            `json:"state"`
	CodeVerifier string            `json:"code_verifier,omitempty"` // For PKCE
	Scopes       []string          `json:"scopes"`
	RedirectURL  string            `json:"redirect_url"`
	ExtraData    map[string]string `json:"extra_data,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	ExpiresAt    time.Time         `json:"expires_at"`
}

// OAuthAdapter defines the interface that all OAuth provider adapters must implement
type OAuthAdapter interface {
	// GetProviderType returns the type of this OAuth provider
	GetProviderType() OAuthProviderType
	
	// GetProviderName returns the human-readable name of this provider
	GetProviderName() string
	
	// ValidateConfig validates the provider configuration
	ValidateConfig(config *OAuthProviderConfig) error
	
	// Configure sets up the adapter with the given configuration
	Configure(config *OAuthProviderConfig) error
	
	// GetAuthorizationURL generates the OAuth authorization URL with optional PKCE
	GetAuthorizationURL(ctx context.Context, session *OAuthSession) (string, error)
	
	// ExchangeCodeForToken exchanges authorization code for access token
	ExchangeCodeForToken(ctx context.Context, code string, session *OAuthSession) (*oauth2.Token, error)
	
	// GetUserInfo retrieves user information using the access token
	GetUserInfo(ctx context.Context, token *oauth2.Token) (*OAuthUserInfo, error)
	
	// RefreshToken refreshes an access token if supported
	RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error)
	
	// ValidateToken validates if a token is still valid
	ValidateToken(ctx context.Context, token *oauth2.Token) (bool, error)
	
	// RevokeToken revokes a token if supported by the provider
	RevokeToken(ctx context.Context, token *oauth2.Token) error
	
	// GetSupportedScopes returns the list of scopes supported by this provider
	GetSupportedScopes() []string
	
	// GetDefaultScopes returns the recommended default scopes for this provider
	GetDefaultScopes() []string
	
	// SupportsFeature checks if the provider supports a specific feature
	SupportsFeature(feature OAuthFeature) bool
}

// OAuthFeature represents features that providers may or may not support
type OAuthFeature string

const (
	FeaturePKCE         OAuthFeature = "pkce"          // PKCE support
	FeatureRefreshToken OAuthFeature = "refresh_token" // Token refresh
	FeatureTokenRevoke  OAuthFeature = "token_revoke"  // Token revocation
	FeatureUserInfo     OAuthFeature = "user_info"     // User info endpoint
	FeatureEmailScope   OAuthFeature = "email_scope"   // Email access
	FeaturePhoneScope   OAuthFeature = "phone_scope"   // Phone access
	FeatureProfileScope OAuthFeature = "profile_scope" // Profile access
)

// OAuthAdapterFactory is responsible for creating OAuth adapters
type OAuthAdapterFactory interface {
	// CreateAdapter creates an OAuth adapter for the specified provider type
	CreateAdapter(providerType OAuthProviderType) (OAuthAdapter, error)
	
	// RegisterAdapter registers a new adapter implementation
	RegisterAdapter(providerType OAuthProviderType, adapterFunc func() OAuthAdapter) error
	
	// GetSupportedProviders returns a list of all supported provider types
	GetSupportedProviders() []OAuthProviderType
	
	// IsProviderSupported checks if a provider type is supported
	IsProviderSupported(providerType OAuthProviderType) bool
}

// OAuthSessionManager manages OAuth authentication sessions
type OAuthSessionManager interface {
	// CreateSession creates a new OAuth session
	CreateSession(ctx context.Context, session *OAuthSession) error
	
	// GetSession retrieves an OAuth session by session ID
	GetSession(ctx context.Context, sessionID string) (*OAuthSession, error)
	
	// GetSessionByState retrieves an OAuth session by state parameter
	GetSessionByState(ctx context.Context, state string) (*OAuthSession, error)
	
	// UpdateSession updates an existing OAuth session
	UpdateSession(ctx context.Context, session *OAuthSession) error
	
	// DeleteSession deletes an OAuth session
	DeleteSession(ctx context.Context, sessionID string) error
	
	// CleanupExpiredSessions removes expired sessions
	CleanupExpiredSessions(ctx context.Context) error
	
	// ValidateState validates the state parameter for CSRF protection
	ValidateState(ctx context.Context, state string) (*OAuthSession, error)
}

// OAuthTokenManager manages OAuth tokens
type OAuthTokenManager interface {
	// StoreToken stores an OAuth token
	StoreToken(ctx context.Context, userID, providerType string, token *oauth2.Token) error
	
	// GetToken retrieves a stored OAuth token
	GetToken(ctx context.Context, userID, providerType string) (*oauth2.Token, error)
	
	// RefreshToken refreshes an expired token
	RefreshToken(ctx context.Context, userID, providerType string) (*oauth2.Token, error)
	
	// DeleteToken deletes a stored token
	DeleteToken(ctx context.Context, userID, providerType string) error
	
	// IsTokenValid checks if a token is valid and not expired
	IsTokenValid(ctx context.Context, token *oauth2.Token) bool
}

// OAuthErrorType represents different types of OAuth errors
type OAuthErrorType string

const (
	ErrorTypeInvalidRequest      OAuthErrorType = "invalid_request"
	ErrorTypeUnauthorizedClient  OAuthErrorType = "unauthorized_client"
	ErrorTypeAccessDenied        OAuthErrorType = "access_denied"
	ErrorTypeUnsupportedResponse OAuthErrorType = "unsupported_response_type"
	ErrorTypeInvalidScope        OAuthErrorType = "invalid_scope"
	ErrorTypeServerError         OAuthErrorType = "server_error"
	ErrorTypeTemporaryUnavailable OAuthErrorType = "temporarily_unavailable"
	ErrorTypeInvalidToken        OAuthErrorType = "invalid_token"
	ErrorTypeTokenExpired        OAuthErrorType = "token_expired"
	ErrorTypeProviderError       OAuthErrorType = "provider_error"
	ErrorTypeConfigurationError  OAuthErrorType = "configuration_error"
	ErrorTypeNetworkError        OAuthErrorType = "network_error"
)

// OAuthError represents an OAuth-related error
type OAuthError struct {
	Type        OAuthErrorType `json:"type"`
	Code        string         `json:"code"`
	Description string         `json:"description"`
	URI         string         `json:"uri,omitempty"`
	Hint        string         `json:"hint,omitempty"`
	Provider    string         `json:"provider,omitempty"`
	Cause       error          `json:"-"`
}

// Error implements the error interface
func (e *OAuthError) Error() string {
	if e.Description != "" {
		return e.Description
	}
	return string(e.Type)
}

// Unwrap returns the underlying error
func (e *OAuthError) Unwrap() error {
	return e.Cause
}