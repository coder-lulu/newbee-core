package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2"

	"github.com/coder-lulu/newbee-core/rpc/internal/interfaces"
)

// GoogleAdapter implements OAuth adapter for Google
type GoogleAdapter struct {
	*BaseOAuthAdapter
}

// GoogleUserInfo represents user information from Google API
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

// NewGoogleAdapter creates a new Google OAuth adapter
func NewGoogleAdapter() interfaces.OAuthAdapter {
	return &GoogleAdapter{
		BaseOAuthAdapter: NewBaseOAuthAdapter(),
	}
}

// GetProviderType returns the provider type
func (g *GoogleAdapter) GetProviderType() interfaces.OAuthProviderType {
	return interfaces.ProviderTypeGoogle
}

// GetProviderName returns the provider name
func (g *GoogleAdapter) GetProviderName() string {
	return "Google"
}

// Configure sets up the Google adapter with the given configuration
func (g *GoogleAdapter) Configure(config *interfaces.OAuthProviderConfig) error {
	// Set Google-specific defaults
	if config.AuthURL == "" {
		config.AuthURL = "https://accounts.google.com/o/oauth2/auth"
	}
	if config.TokenURL == "" {
		config.TokenURL = "https://oauth2.googleapis.com/token"
	}
	if config.UserInfoURL == "" {
		config.UserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	}
	if len(config.Scopes) == 0 {
		config.Scopes = []string{"openid", "profile", "email"}
	}

	// Google supports PKCE
	config.SupportPKCE = true

	return g.BaseOAuthAdapter.Configure(config)
}

// GetUserInfo retrieves user information using the access token
func (g *GoogleAdapter) GetUserInfo(ctx context.Context, token *oauth2.Token) (*interfaces.OAuthUserInfo, error) {
	if g.config == nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeConfigurationError,
			Description: "adapter not configured",
		}
	}

	// Fetch user profile
	userInfo, err := g.fetchUserProfile(ctx, token)
	if err != nil {
		return nil, err
	}

	// Create standardized user info
	result := &interfaces.OAuthUserInfo{
		ID:           userInfo.ID,
		Username:     userInfo.Email, // Google uses email as username
		Nickname:     userInfo.Name,
		Email:        userInfo.Email,
		Avatar:       userInfo.Picture,
		Verified:     userInfo.VerifiedEmail,
		CreatedAt:    time.Time{}, // Google doesn't provide registration time
		UpdatedAt:    time.Now(),
		ProviderType: interfaces.ProviderTypeGoogle,
		RawData: map[string]interface{}{
			"google": userInfo,
		},
	}

	return result, nil
}

// fetchUserProfile fetches user profile from Google API
func (g *GoogleAdapter) fetchUserProfile(ctx context.Context, token *oauth2.Token) (*GoogleUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", g.config.UserInfoURL, nil)
	if err != nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeNetworkError,
			Description: fmt.Sprintf("failed to create request: %v", err),
			Cause:       err,
		}
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeNetworkError,
			Description: fmt.Sprintf("failed to fetch user info: %v", err),
			Cause:       err,
		}
	}
	defer resp.Body.Close()

	if err := g.parseErrorResponse(resp); err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeProviderError,
			Description: fmt.Sprintf("failed to read response: %v", err),
			Cause:       err,
		}
	}

	var userInfo GoogleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeProviderError,
			Description: fmt.Sprintf("failed to parse user info: %v", err),
			Cause:       err,
		}
	}

	return &userInfo, nil
}

// GetSupportedScopes returns the scopes supported by Google
func (g *GoogleAdapter) GetSupportedScopes() []string {
	return []string{
		"openid",
		"profile",
		"email",
		"https://www.googleapis.com/auth/userinfo.profile",
		"https://www.googleapis.com/auth/userinfo.email",
	}
}

// GetDefaultScopes returns the recommended default scopes for Google
func (g *GoogleAdapter) GetDefaultScopes() []string {
	return []string{"openid", "profile", "email"}
}

// SupportsFeature checks if Google supports a specific feature
func (g *GoogleAdapter) SupportsFeature(feature interfaces.OAuthFeature) bool {
	switch feature {
	case interfaces.FeaturePKCE:
		return true
	case interfaces.FeatureRefreshToken:
		return true
	case interfaces.FeatureTokenRevoke:
		return true
	case interfaces.FeatureUserInfo:
		return true
	case interfaces.FeatureEmailScope:
		return true
	case interfaces.FeatureProfileScope:
		return true
	default:
		return g.BaseOAuthAdapter.SupportsFeature(feature)
	}
}