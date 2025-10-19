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

// GitHubAdapter implements OAuth adapter for GitHub
type GitHubAdapter struct {
	*BaseOAuthAdapter
}

// GitHubUserInfo represents user information from GitHub API
type GitHubUserInfo struct {
	ID        int64     `json:"id"`
	Login     string    `json:"login"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatar_url"`
	Bio       string    `json:"bio"`
	Blog      string    `json:"blog"`
	Company   string    `json:"company"`
	Location  string    `json:"location"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GitHubEmail represents email information from GitHub API
type GitHubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

// NewGitHubAdapter creates a new GitHub OAuth adapter
func NewGitHubAdapter() interfaces.OAuthAdapter {
	return &GitHubAdapter{
		BaseOAuthAdapter: NewBaseOAuthAdapter(),
	}
}

// GetProviderType returns the provider type
func (g *GitHubAdapter) GetProviderType() interfaces.OAuthProviderType {
	return interfaces.ProviderTypeGitHub
}

// GetProviderName returns the provider name
func (g *GitHubAdapter) GetProviderName() string {
	return "GitHub"
}

// Configure sets up the GitHub adapter with the given configuration
func (g *GitHubAdapter) Configure(config *interfaces.OAuthProviderConfig) error {
	// Set GitHub-specific defaults
	if config.AuthURL == "" {
		config.AuthURL = "https://github.com/login/oauth/authorize"
	}
	if config.TokenURL == "" {
		config.TokenURL = "https://github.com/login/oauth/access_token"
	}
	if config.UserInfoURL == "" {
		config.UserInfoURL = "https://api.github.com/user"
	}
	if len(config.Scopes) == 0 {
		config.Scopes = []string{"user:email"}
	}

	// Set PKCE support
	config.SupportPKCE = true

	return g.BaseOAuthAdapter.Configure(config)
}

// GetUserInfo retrieves user information using the access token
func (g *GitHubAdapter) GetUserInfo(ctx context.Context, token *oauth2.Token) (*interfaces.OAuthUserInfo, error) {
	if g.config == nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeConfigurationError,
			Description: "adapter not configured",
		}
	}

	// Get user profile
	userInfo, err := g.fetchUserProfile(ctx, token)
	if err != nil {
		return nil, err
	}

	// Get user emails if scope includes email
	emails, err := g.fetchUserEmails(ctx, token)
	if err != nil {
		// Don't fail if we can't get emails, just log the error
		// In production, you might want to log this
	}

	// Find primary email
	primaryEmail := userInfo.Email
	if primaryEmail == "" && len(emails) > 0 {
		for _, email := range emails {
			if email.Primary && email.Verified {
				primaryEmail = email.Email
				break
			}
		}
		// If no primary verified email, use the first verified one
		if primaryEmail == "" {
			for _, email := range emails {
				if email.Verified {
					primaryEmail = email.Email
					break
				}
			}
		}
	}

	// Create standardized user info
	result := &interfaces.OAuthUserInfo{
		ID:           fmt.Sprintf("%d", userInfo.ID),
		Username:     userInfo.Login,
		Nickname:     userInfo.Name,
		Email:        primaryEmail,
		Avatar:       userInfo.AvatarURL,
		Bio:          userInfo.Bio,
		Website:      userInfo.Blog,
		Company:      userInfo.Company,
		Location:     userInfo.Location,
		Verified:     primaryEmail != "",
		CreatedAt:    userInfo.CreatedAt,
		UpdatedAt:    userInfo.UpdatedAt,
		ProviderType: interfaces.ProviderTypeGitHub,
		RawData: map[string]interface{}{
			"user":   userInfo,
			"emails": emails,
		},
	}

	return result, nil
}

// fetchUserProfile fetches user profile from GitHub API
func (g *GitHubAdapter) fetchUserProfile(ctx context.Context, token *oauth2.Token) (*GitHubUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", g.config.UserInfoURL, nil)
	if err != nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeNetworkError,
			Description: fmt.Sprintf("failed to create request: %v", err),
			Cause:       err,
		}
	}

	req.Header.Set("Authorization", "token "+token.AccessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "OAuth-Client/1.0")

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

	var userInfo GitHubUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeProviderError,
			Description: fmt.Sprintf("failed to parse user info: %v", err),
			Cause:       err,
		}
	}

	return &userInfo, nil
}

// fetchUserEmails fetches user emails from GitHub API
func (g *GitHubAdapter) fetchUserEmails(ctx context.Context, token *oauth2.Token) ([]GitHubEmail, error) {
	emailsURL := "https://api.github.com/user/emails"
	
	req, err := http.NewRequestWithContext(ctx, "GET", emailsURL, nil)
	if err != nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeNetworkError,
			Description: fmt.Sprintf("failed to create emails request: %v", err),
			Cause:       err,
		}
	}

	req.Header.Set("Authorization", "token "+token.AccessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "OAuth-Client/1.0")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeNetworkError,
			Description: fmt.Sprintf("failed to fetch user emails: %v", err),
			Cause:       err,
		}
	}
	defer resp.Body.Close()

	// If we can't access emails (maybe no permission), return empty slice
	if resp.StatusCode == 403 || resp.StatusCode == 404 {
		return []GitHubEmail{}, nil
	}

	if err := g.parseErrorResponse(resp); err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeProviderError,
			Description: fmt.Sprintf("failed to read emails response: %v", err),
			Cause:       err,
		}
	}

	var emails []GitHubEmail
	if err := json.Unmarshal(body, &emails); err != nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeProviderError,
			Description: fmt.Sprintf("failed to parse emails: %v", err),
			Cause:       err,
		}
	}

	return emails, nil
}

// RevokeToken revokes a GitHub access token
func (g *GitHubAdapter) RevokeToken(ctx context.Context, token *oauth2.Token) error {
	if g.config == nil {
		return &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeConfigurationError,
			Description: "adapter not configured",
		}
	}

	revokeURL := fmt.Sprintf("https://api.github.com/applications/%s/token", g.config.ClientID)
	
	req, err := http.NewRequestWithContext(ctx, "DELETE", revokeURL, nil)
	if err != nil {
		return &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeNetworkError,
			Description: fmt.Sprintf("failed to create revoke request: %v", err),
			Cause:       err,
		}
	}

	req.SetBasicAuth(g.config.ClientID, g.config.ClientSecret)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "OAuth-Client/1.0")

	// Set the token in the request body
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeNetworkError,
			Description: fmt.Sprintf("failed to revoke token: %v", err),
			Cause:       err,
		}
	}
	defer resp.Body.Close()

	return g.parseErrorResponse(resp)
}

// GetSupportedScopes returns the scopes supported by GitHub
func (g *GitHubAdapter) GetSupportedScopes() []string {
	return []string{
		"repo",
		"repo:status",
		"repo_deployment",
		"public_repo",
		"repo:invite",
		"security_events",
		"admin:repo_hook",
		"write:repo_hook",
		"read:repo_hook",
		"admin:org",
		"write:org",
		"read:org",
		"admin:public_key",
		"write:public_key",
		"read:public_key",
		"admin:org_hook",
		"gist",
		"notifications",
		"user",
		"read:user",
		"user:email",
		"user:follow",
		"delete_repo",
		"write:discussion",
		"read:discussion",
		"admin:gpg_key",
		"write:gpg_key",
		"read:gpg_key",
	}
}

// GetDefaultScopes returns the recommended default scopes for GitHub
func (g *GitHubAdapter) GetDefaultScopes() []string {
	return []string{"user:email"}
}

// SupportsFeature checks if GitHub supports a specific feature
func (g *GitHubAdapter) SupportsFeature(feature interfaces.OAuthFeature) bool {
	switch feature {
	case interfaces.FeaturePKCE:
		return true
	case interfaces.FeatureRefreshToken:
		return false // GitHub doesn't support refresh tokens
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