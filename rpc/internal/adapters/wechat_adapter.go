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

// WechatAdapter implements OAuth adapter for WeChat
type WechatAdapter struct {
	*BaseOAuthAdapter
}

// WechatUserInfo represents user information from WeChat API
type WechatUserInfo struct {
	OpenID     string `json:"openid"`
	UnionID    string `json:"unionid"`
	Nickname   string `json:"nickname"`
	Sex        int    `json:"sex"`
	Province   string `json:"province"`
	City       string `json:"city"`
	Country    string `json:"country"`
	HeadImgURL string `json:"headimgurl"`
	Language   string `json:"language"`
}

// NewWechatAdapter creates a new WeChat OAuth adapter
func NewWechatAdapter() interfaces.OAuthAdapter {
	return &WechatAdapter{
		BaseOAuthAdapter: NewBaseOAuthAdapter(),
	}
}

// GetProviderType returns the provider type
func (w *WechatAdapter) GetProviderType() interfaces.OAuthProviderType {
	return interfaces.ProviderTypeWechat
}

// GetProviderName returns the provider name
func (w *WechatAdapter) GetProviderName() string {
	return "微信"
}

// Configure sets up the WeChat adapter with the given configuration
func (w *WechatAdapter) Configure(config *interfaces.OAuthProviderConfig) error {
	// Set WeChat-specific defaults
	if config.AuthURL == "" {
		config.AuthURL = "https://open.weixin.qq.com/connect/qrconnect"
	}
	if config.TokenURL == "" {
		config.TokenURL = "https://api.weixin.qq.com/sns/oauth2/access_token"
	}
	if config.UserInfoURL == "" {
		config.UserInfoURL = "https://api.weixin.qq.com/sns/userinfo"
	}
	if len(config.Scopes) == 0 {
		config.Scopes = []string{"snsapi_login"}
	}

	// WeChat doesn't support PKCE
	config.SupportPKCE = false

	return w.BaseOAuthAdapter.Configure(config)
}

// GetUserInfo retrieves user information using the access token
func (w *WechatAdapter) GetUserInfo(ctx context.Context, token *oauth2.Token) (*interfaces.OAuthUserInfo, error) {
	if w.config == nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeConfigurationError,
			Description: "adapter not configured",
		}
	}

	// WeChat requires both access_token and openid for user info
	openid := token.Extra("openid")
	if openid == nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeProviderError,
			Description: "openid not found in token response",
		}
	}

	// Construct user info URL with parameters
	userInfoURL := fmt.Sprintf("%s?access_token=%s&openid=%s&lang=zh_CN",
		w.config.UserInfoURL, token.AccessToken, openid)

	req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
	if err != nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeNetworkError,
			Description: fmt.Sprintf("failed to create request: %v", err),
			Cause:       err,
		}
	}

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeNetworkError,
			Description: fmt.Sprintf("failed to fetch user info: %v", err),
			Cause:       err,
		}
	}
	defer resp.Body.Close()

	if err := w.parseErrorResponse(resp); err != nil {
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

	var userInfo WechatUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, &interfaces.OAuthError{
			Type:        interfaces.ErrorTypeProviderError,
			Description: fmt.Sprintf("failed to parse user info: %v", err),
			Cause:       err,
		}
	}

	// WeChat doesn't provide email, construct location from province/city
	location := ""
	if userInfo.Province != "" && userInfo.City != "" {
		location = fmt.Sprintf("%s, %s", userInfo.City, userInfo.Province)
	} else if userInfo.City != "" {
		location = userInfo.City
	} else if userInfo.Province != "" {
		location = userInfo.Province
	}

	// Create standardized user info
	result := &interfaces.OAuthUserInfo{
		ID:           userInfo.OpenID,
		Username:     userInfo.OpenID, // WeChat uses OpenID as unique identifier
		Nickname:     userInfo.Nickname,
		Email:        "", // WeChat doesn't provide email
		Avatar:       userInfo.HeadImgURL,
		Location:     location,
		Verified:     false, // WeChat doesn't provide verification status
		CreatedAt:    time.Time{}, // WeChat doesn't provide registration time
		UpdatedAt:    time.Now(),
		ProviderType: interfaces.ProviderTypeWechat,
		RawData: map[string]interface{}{
			"wechat": userInfo,
		},
	}

	return result, nil
}

// GetSupportedScopes returns the scopes supported by WeChat
func (w *WechatAdapter) GetSupportedScopes() []string {
	return []string{
		"snsapi_login",
		"snsapi_userinfo",
	}
}

// GetDefaultScopes returns the recommended default scopes for WeChat
func (w *WechatAdapter) GetDefaultScopes() []string {
	return []string{"snsapi_login"}
}

// SupportsFeature checks if WeChat supports a specific feature
func (w *WechatAdapter) SupportsFeature(feature interfaces.OAuthFeature) bool {
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
		return false // WeChat doesn't provide email
	case interfaces.FeatureProfileScope:
		return true
	default:
		return w.BaseOAuthAdapter.SupportsFeature(feature)
	}
}