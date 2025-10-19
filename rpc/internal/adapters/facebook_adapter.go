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

// FacebookAdapter implements OAuth adapter for Facebook
type FacebookAdapter struct {
	*BaseOAuthAdapter
}

// FacebookUserInfo represents user information from Facebook API
type FacebookUserInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Picture  FacebookPictureData `json:"picture"`
	Location FacebookLocationData `json:"location"`
}

// FacebookPictureData represents the picture data structure
type FacebookPictureData struct {
	Data FacebookPictureURL `json:"data"`
}

// FacebookPictureURL represents the picture URL structure
type FacebookPictureURL struct {
	Height       int    `json:"height"`
	IsSilhouette bool   `json:"is_silhouette"`
	URL          string `json:"url"`
	Width        int    `json:"width"`
}

// FacebookLocationData represents location information
type FacebookLocationData struct {
	Name string `json:"name"`
}

// NewFacebookAdapter creates a new Facebook OAuth adapter
func NewFacebookAdapter() interfaces.OAuthAdapter {
	return &FacebookAdapter{
		BaseOAuthAdapter: NewBaseOAuthAdapter(),
	}
}

// GetProviderType returns the provider type
func (f *FacebookAdapter) GetProviderType() interfaces.OAuthProviderType {
	return interfaces.ProviderTypeFacebook
}

// GetProviderName returns the provider name
func (f *FacebookAdapter) GetProviderName() string {
	return "Facebook"
}

// Configure sets up the Facebook adapter with the given configuration
func (f *FacebookAdapter) Configure(config *interfaces.OAuthProviderConfig) error {
	// Set Facebook-specific defaults
	if config.AuthURL == "" {
		config.AuthURL = "https://www.facebook.com/v18.0/dialog/oauth"
	}
	if config.TokenURL == "" {
		config.TokenURL = "https://graph.facebook.com/v18.0/oauth/access_token"
	}
	if config.UserInfoURL == "" {
		config.UserInfoURL = "https://graph.facebook.com/v18.0/me"
	}
	if len(config.Scopes) == 0 {
		config.Scopes = []string{"email", "public_profile"}
	}

	// Facebook supports PKCE
	config.SupportPKCE = true

	return f.BaseOAuthAdapter.Configure(config)
}

// GetUserInfo retrieves user information using the access token
func (f *FacebookAdapter) GetUserInfo(ctx context.Context, token *oauth2.Token) (*interfaces.OAuthUserInfo, error) {
	if f.config == nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeConfigurationError,
			Description: "adapter not configured",
		}
	}

	// Fetch user profile
	userInfo, err := f.fetchUserProfile(ctx, token)
	if err != nil {
		return nil, err
	}

	// Extract avatar URL
	avatarURL := ""
	if userInfo.Picture.Data.URL != "" && !userInfo.Picture.Data.IsSilhouette {
		avatarURL = userInfo.Picture.Data.URL
	}

	// Create standardized user info
	result := &interfaces.OAuthUserInfo{
		ID:           userInfo.ID,
		Username:     userInfo.ID, // Facebook uses ID as username
		Nickname:     userInfo.Name,
		Email:        userInfo.Email,
		Avatar:       avatarURL,
		Location:     userInfo.Location.Name,
		Verified:     userInfo.Email != "", // Consider user verified if email is provided
		CreatedAt:    time.Time{}, // Facebook doesn't provide registration time in basic info
		UpdatedAt:    time.Now(),
		ProviderType: interfaces.ProviderTypeFacebook,
		RawData: map[string]interface{}{
			"facebook": userInfo,
		},
	}

	return result, nil
}

// fetchUserProfile fetches user profile from Facebook API
func (f *FacebookAdapter) fetchUserProfile(ctx context.Context, token *oauth2.Token) (*FacebookUserInfo, error) {
	// Request specific fields to get comprehensive user info
	fieldsParam := "id,name,email,picture{url,width,height,is_silhouette},location"
	userInfoURL := fmt.Sprintf("%s?fields=%s&access_token=%s", 
		f.config.UserInfoURL, fieldsParam, token.AccessToken)

	req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
	if err != nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeNetworkError,
			Description: fmt.Sprintf("failed to create request: %v", err),
			Cause:       err,
		}
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeNetworkError,
			Description: fmt.Sprintf("failed to fetch user info: %v", err),
			Cause:       err,
		}
	}
	defer resp.Body.Close()

	if err := f.parseErrorResponse(resp); err != nil {
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

	var userInfo FacebookUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeProviderError,
			Description: fmt.Sprintf("failed to parse user info: %v", err),
			Cause:       err,
		}
	}

	return &userInfo, nil
}

// GetSupportedScopes returns the scopes supported by Facebook
func (f *FacebookAdapter) GetSupportedScopes() []string {
	return []string{
		"public_profile",
		"email",
		"user_birthday",
		"user_friends",
		"user_hometown",
		"user_location",
	}
}

// GetDefaultScopes returns the recommended default scopes for Facebook
func (f *FacebookAdapter) GetDefaultScopes() []string {
	return []string{"email", "public_profile"}
}

// SupportsFeature checks if Facebook supports a specific feature
func (f *FacebookAdapter) SupportsFeature(feature interfaces.OAuthFeature) bool {
	switch feature {
	case interfaces.FeaturePKCE:
		return true
	case interfaces.FeatureRefreshToken:
		return false // Facebook uses long-lived tokens instead
	case interfaces.FeatureTokenRevoke:
		return true
	case interfaces.FeatureUserInfo:
		return true
	case interfaces.FeatureEmailScope:
		return true
	case interfaces.FeatureProfileScope:
		return true
	default:
		return f.BaseOAuthAdapter.SupportsFeature(feature)
	}
}