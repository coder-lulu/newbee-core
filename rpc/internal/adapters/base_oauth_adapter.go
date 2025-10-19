package adapters

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/coder-lulu/newbee-core/rpc/internal/interfaces"
)

// BaseOAuthAdapter provides common functionality for OAuth adapters
type BaseOAuthAdapter struct {
	config       *interfaces.OAuthProviderConfig
	oauth2Config *oauth2.Config
	httpClient   *http.Client
}

// NewBaseOAuthAdapter creates a new base OAuth adapter
func NewBaseOAuthAdapter() *BaseOAuthAdapter {
	return &BaseOAuthAdapter{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Configure sets up the base adapter with the given configuration
func (b *BaseOAuthAdapter) Configure(config *interfaces.OAuthProviderConfig) error {
	if err := b.ValidateConfig(config); err != nil {
		return err
	}

	b.config = config
	b.oauth2Config = &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       config.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:   config.AuthURL,
			TokenURL:  config.TokenURL,
			AuthStyle: config.AuthStyle,
		},
	}

	return nil
}

// ValidateConfig validates the base OAuth provider configuration
func (b *BaseOAuthAdapter) ValidateConfig(config *interfaces.OAuthProviderConfig) error {
	if config == nil {
		return &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeConfigurationError,
			Description: "configuration cannot be nil",
		}
	}

	if config.ClientID == "" {
		return &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeConfigurationError,
			Description: "client_id is required",
		}
	}

	if config.ClientSecret == "" {
		return &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeConfigurationError,
			Description: "client_secret is required",
		}
	}

	if config.RedirectURL == "" {
		return &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeConfigurationError,
			Description: "redirect_url is required",
		}
	}

	if config.AuthURL == "" {
		return &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeConfigurationError,
			Description: "auth_url is required",
		}
	}

	if config.TokenURL == "" {
		return &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeConfigurationError,
			Description: "token_url is required",
		}
	}

	// Validate redirect URL format
	if _, err := url.Parse(config.RedirectURL); err != nil {
		return &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeConfigurationError,
			Description: fmt.Sprintf("invalid redirect_url format: %v", err),
		}
	}

	return nil
}

// GetAuthorizationURL generates the OAuth authorization URL with optional PKCE
func (b *BaseOAuthAdapter) GetAuthorizationURL(ctx context.Context, session *interfaces.OAuthSession) (string, error) {
	if b.oauth2Config == nil {
		return "", &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeConfigurationError,
			Description: "adapter not configured",
		}
	}

	opts := make([]oauth2.AuthCodeOption, 0)

	// Add PKCE if supported
	if b.config.SupportPKCE && session.CodeVerifier != "" {
		codeChallenge := b.generateCodeChallenge(session.CodeVerifier)
		opts = append(opts,
			oauth2.SetAuthURLParam("code_challenge", codeChallenge),
			oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		)
	}

	// Add any extra parameters from config
	for key, value := range b.config.ExtraConfig {
		if strValue, ok := value.(string); ok {
			opts = append(opts, oauth2.SetAuthURLParam(key, strValue))
		}
	}

	authURL := b.oauth2Config.AuthCodeURL(session.State, opts...)
	return authURL, nil
}

// ExchangeCodeForToken exchanges authorization code for access token
func (b *BaseOAuthAdapter) ExchangeCodeForToken(ctx context.Context, code string, session *interfaces.OAuthSession) (*oauth2.Token, error) {
	if b.oauth2Config == nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeConfigurationError,
			Description: "adapter not configured",
		}
	}

	opts := make([]oauth2.AuthCodeOption, 0)

	// Add PKCE code verifier if used
	if b.config.SupportPKCE && session.CodeVerifier != "" {
		opts = append(opts, oauth2.SetAuthURLParam("code_verifier", session.CodeVerifier))
	}

	// Exchange code for token
	token, err := b.oauth2Config.Exchange(ctx, code, opts...)
	if err != nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeProviderError,
			Description: fmt.Sprintf("failed to exchange code for token: %v", err),
			Cause:       err,
		}
	}

	return token, nil
}

// RefreshToken refreshes an access token if supported
func (b *BaseOAuthAdapter) RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	if b.oauth2Config == nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeConfigurationError,
			Description: "adapter not configured",
		}
	}

	tokenSource := b.oauth2Config.TokenSource(ctx, &oauth2.Token{
		RefreshToken: refreshToken,
	})

	token, err := tokenSource.Token()
	if err != nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeProviderError,
			Description: fmt.Sprintf("failed to refresh token: %v", err),
			Cause:       err,
		}
	}

	return token, nil
}

// ValidateToken validates if a token is still valid
func (b *BaseOAuthAdapter) ValidateToken(ctx context.Context, token *oauth2.Token) (bool, error) {
	if token == nil {
		return false, nil
	}

	// Check if token is expired
	if token.Expiry.Before(time.Now()) {
		return false, nil
	}

	return true, nil
}

// RevokeToken provides a default implementation that returns not supported
func (b *BaseOAuthAdapter) RevokeToken(ctx context.Context, token *oauth2.Token) error {
	return &interfaces.OAuthError{
		Type:        interfaces.ErrorTypeUnsupportedResponse,
		Description: "token revocation not supported by this provider",
	}
}

// SupportsFeature provides default feature support detection
func (b *BaseOAuthAdapter) SupportsFeature(feature interfaces.OAuthFeature) bool {
	switch feature {
	case interfaces.FeaturePKCE:
		return b.config.SupportPKCE
	case interfaces.FeatureRefreshToken:
		return true // Most providers support refresh tokens
	case interfaces.FeatureUserInfo:
		return b.config.UserInfoURL != ""
	default:
		return false
	}
}

// generateCodeVerifier generates a PKCE code verifier
func (b *BaseOAuthAdapter) generateCodeVerifier() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// generateCodeChallenge generates a PKCE code challenge from the verifier
func (b *BaseOAuthAdapter) generateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// GeneratePKCEPair generates both code verifier and challenge for PKCE
func (b *BaseOAuthAdapter) GeneratePKCEPair() (verifier, challenge string, err error) {
	verifier, err = b.generateCodeVerifier()
	if err != nil {
		return "", "", err
	}
	challenge = b.generateCodeChallenge(verifier)
	return verifier, challenge, nil
}

// makeHTTPRequest is a helper method for making HTTP requests to OAuth providers
func (b *BaseOAuthAdapter) makeHTTPRequest(ctx context.Context, method, url string, headers map[string]string, body interface{}) (*http.Response, error) {

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeNetworkError,
			Description: fmt.Sprintf("failed to create HTTP request: %v", err),
			Cause:       err,
		}
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Set default headers
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "OAuth-Client/1.0")
	}

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeNetworkError,
			Description: fmt.Sprintf("HTTP request failed: %v", err),
			Cause:       err,
		}
	}

	return resp, nil
}

// parseErrorResponse parses error responses from OAuth providers
func (b *BaseOAuthAdapter) parseErrorResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	errorType := interfaces.ErrorTypeProviderError
	description := fmt.Sprintf("HTTP %d", resp.StatusCode)

	switch resp.StatusCode {
	case 400:
		errorType = interfaces.ErrorTypeInvalidRequest
		description = "Bad Request"
	case 401:
		errorType = interfaces.ErrorTypeUnauthorizedClient
		description = "Unauthorized"
	case 403:
		errorType = interfaces.ErrorTypeAccessDenied
		description = "Access Denied"
	case 500:
		errorType = interfaces.ErrorTypeServerError
		description = "Internal Server Error"
	case 503:
		errorType = interfaces.ErrorTypeTemporaryUnavailable
		description = "Service Temporarily Unavailable"
	}

	return &interfaces.OAuthError{
		Type:        errorType,
		Code:        fmt.Sprintf("%d", resp.StatusCode),
		Description: description,
		Provider:    string(b.config.Type),
	}
}

// normalizeScopes normalizes scope strings (handles comma vs space separation)
func (b *BaseOAuthAdapter) normalizeScopes(scopes []string) []string {
	var normalized []string
	for _, scope := range scopes {
		// Split by comma or space
		parts := strings.FieldsFunc(scope, func(r rune) bool {
			return r == ',' || r == ' '
		})
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				normalized = append(normalized, trimmed)
			}
		}
	}
	return normalized
}