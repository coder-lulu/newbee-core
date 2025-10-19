package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"golang.org/x/oauth2"

	"github.com/coder-lulu/newbee-core/rpc/internal/interfaces"
)

// QQAdapter implements OAuth adapter for QQ
type QQAdapter struct {
	*BaseOAuthAdapter
}

// QQOpenIDResponse represents the OpenID response from QQ
type QQOpenIDResponse struct {
	ClientID string `json:"client_id"`
	OpenID   string `json:"openid"`
}

// QQUserInfo represents user information from QQ API
type QQUserInfo struct {
	Ret             int    `json:"ret"`
	Msg             string `json:"msg"`
	Nickname        string `json:"nickname"`
	FigureURL       string `json:"figureurl"`
	FigureURL1      string `json:"figureurl_1"`
	FigureURL2      string `json:"figureurl_2"`
	FigureURLQQ1    string `json:"figureurl_qq_1"`
	FigureURLQQ2    string `json:"figureurl_qq_2"`
	Gender          string `json:"gender"`
	IsYellowVip     string `json:"is_yellow_vip"`
	Vip             string `json:"vip"`
	YellowVipLevel  string `json:"yellow_vip_level"`
	Level           string `json:"level"`
	IsYellowYearVip string `json:"is_yellow_year_vip"`
}

// NewQQAdapter creates a new QQ OAuth adapter
func NewQQAdapter() interfaces.OAuthAdapter {
	return &QQAdapter{
		BaseOAuthAdapter: NewBaseOAuthAdapter(),
	}
}

// GetProviderType returns the provider type
func (q *QQAdapter) GetProviderType() interfaces.OAuthProviderType {
	return interfaces.ProviderTypeQQ
}

// GetProviderName returns the provider name
func (q *QQAdapter) GetProviderName() string {
	return "QQ"
}

// Configure sets up the QQ adapter with the given configuration
func (q *QQAdapter) Configure(config *interfaces.OAuthProviderConfig) error {
	// Set QQ-specific defaults
	if config.AuthURL == "" {
		config.AuthURL = "https://graph.qq.com/oauth2.0/authorize"
	}
	if config.TokenURL == "" {
		config.TokenURL = "https://graph.qq.com/oauth2.0/token"
	}
	if config.UserInfoURL == "" {
		config.UserInfoURL = "https://graph.qq.com/user/get_user_info"
	}
	if len(config.Scopes) == 0 {
		config.Scopes = []string{"get_user_info"}
	}

	// QQ doesn't support PKCE
	config.SupportPKCE = false

	return q.BaseOAuthAdapter.Configure(config)
}

// GetUserInfo retrieves user information using the access token
func (q *QQAdapter) GetUserInfo(ctx context.Context, token *oauth2.Token) (*interfaces.OAuthUserInfo, error) {
	if q.config == nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeConfigurationError,
			Description: "adapter not configured",
		}
	}

	// First, get OpenID
	openID, err := q.getOpenID(ctx, token)
	if err != nil {
		return nil, err
	}

	// Then get user info using OpenID
	userInfo, err := q.fetchUserProfile(ctx, token, openID)
	if err != nil {
		return nil, err
	}

	// Check for API error
	if userInfo.Ret != 0 {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeProviderError,
			Description: fmt.Sprintf("QQ API error: %s (code: %d)", userInfo.Msg, userInfo.Ret),
		}
	}

	// Choose best available avatar
	avatar := userInfo.FigureURLQQ2
	if avatar == "" {
		avatar = userInfo.FigureURLQQ1
	}
	if avatar == "" {
		avatar = userInfo.FigureURL2
	}
	if avatar == "" {
		avatar = userInfo.FigureURL1
	}
	if avatar == "" {
		avatar = userInfo.FigureURL
	}

	// Create standardized user info
	result := &interfaces.OAuthUserInfo{
		ID:           openID,
		Username:     openID, // QQ uses OpenID as unique identifier
		Nickname:     userInfo.Nickname,
		Email:        "", // QQ doesn't provide email
		Avatar:       avatar,
		Verified:     false, // QQ doesn't provide verification status
		CreatedAt:    time.Time{}, // QQ doesn't provide registration time
		UpdatedAt:    time.Now(),
		ProviderType: interfaces.ProviderTypeQQ,
		RawData: map[string]interface{}{
			"qq":     userInfo,
			"openid": openID,
		},
	}

	return result, nil
}

// getOpenID retrieves the OpenID from QQ's me endpoint
func (q *QQAdapter) getOpenID(ctx context.Context, token *oauth2.Token) (string, error) {
	openIDURL := fmt.Sprintf("https://graph.qq.com/oauth2.0/me?access_token=%s", token.AccessToken)

	req, err := http.NewRequestWithContext(ctx, "GET", openIDURL, nil)
	if err != nil {
		return "", &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeNetworkError,
			Description: fmt.Sprintf("failed to create OpenID request: %v", err),
			Cause:       err,
		}
	}

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return "", &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeNetworkError,
			Description: fmt.Sprintf("failed to fetch OpenID: %v", err),
			Cause:       err,
		}
	}
	defer resp.Body.Close()

	if err := q.parseErrorResponse(resp); err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeProviderError,
			Description: fmt.Sprintf("failed to read OpenID response: %v", err),
			Cause:       err,
		}
	}

	// QQ returns JSONP format, need to extract JSON
	bodyStr := string(body)
	re := regexp.MustCompile(`callback\(\s*(\{.*\})\s*\);`)
	matches := re.FindStringSubmatch(bodyStr)
	if len(matches) < 2 {
		return "", &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeProviderError,
			Description: "failed to parse OpenID response format",
		}
	}

	var openIDResp QQOpenIDResponse
	if err := json.Unmarshal([]byte(matches[1]), &openIDResp); err != nil {
		return "", &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeProviderError,
			Description: fmt.Sprintf("failed to parse OpenID JSON: %v", err),
			Cause:       err,
		}
	}

	if openIDResp.OpenID == "" {
		return "", &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeProviderError,
			Description: "OpenID not found in response",
		}
	}

	return openIDResp.OpenID, nil
}

// fetchUserProfile fetches user profile from QQ API
func (q *QQAdapter) fetchUserProfile(ctx context.Context, token *oauth2.Token, openID string) (*QQUserInfo, error) {
	userInfoURL := fmt.Sprintf("%s?access_token=%s&oauth_consumer_key=%s&openid=%s",
		q.config.UserInfoURL, token.AccessToken, q.config.ClientID, openID)

	req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
	if err != nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeNetworkError,
			Description: fmt.Sprintf("failed to create user info request: %v", err),
			Cause:       err,
		}
	}

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeNetworkError,
			Description: fmt.Sprintf("failed to fetch user info: %v", err),
			Cause:       err,
		}
	}
	defer resp.Body.Close()

	if err := q.parseErrorResponse(resp); err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeProviderError,
			Description: fmt.Sprintf("failed to read user info response: %v", err),
			Cause:       err,
		}
	}

	var userInfo QQUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeProviderError,
			Description: fmt.Sprintf("failed to parse user info: %v", err),
			Cause:       err,
		}
	}

	return &userInfo, nil
}

// GetSupportedScopes returns the scopes supported by QQ
func (q *QQAdapter) GetSupportedScopes() []string {
	return []string{
		"get_user_info",
		"list_album",
		"upload_pic",
		"do_like",
	}
}

// GetDefaultScopes returns the recommended default scopes for QQ
func (q *QQAdapter) GetDefaultScopes() []string {
	return []string{"get_user_info"}
}

// SupportsFeature checks if QQ supports a specific feature
func (q *QQAdapter) SupportsFeature(feature interfaces.OAuthFeature) bool {
	switch feature {
	case interfaces.FeaturePKCE:
		return false
	case interfaces.FeatureRefreshToken:
		return true
	case interfaces.FeatureTokenRevoke:
		return false
	case interfaces.FeatureUserInfo:
		return true
	case interfaces.FeatureEmailScope:
		return false // QQ doesn't provide email directly
	case interfaces.FeatureProfileScope:
		return true
	default:
		return q.BaseOAuthAdapter.SupportsFeature(feature)
	}
}