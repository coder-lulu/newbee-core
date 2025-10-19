package validation

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/coder-lulu/newbee-core/rpc/internal/interfaces"
)

// ValidationLevel defines the level of validation to perform
type ValidationLevel int

const (
	ValidationLevelBasic     ValidationLevel = iota // Basic field validation
	ValidationLevelExtended                         // Extended validation with connectivity checks
	ValidationLevelFull                             // Full validation including OAuth flow simulation
)

// ValidationResult represents the result of a validation operation
type ValidationResult struct {
	Valid      bool               `json:"valid"`
	Errors     []ValidationError  `json:"errors,omitempty"`
	Warnings   []ValidationError  `json:"warnings,omitempty"`
	Level      ValidationLevel    `json:"level"`
	Provider   string             `json:"provider"`
	ValidatedAt time.Time         `json:"validated_at"`
	Duration   time.Duration      `json:"duration"`
}

// ValidationError represents a validation error or warning
type ValidationError struct {
	Field       string                        `json:"field"`
	Code        string                        `json:"code"`
	Message     string                        `json:"message"`
	Severity    ValidationSeverity            `json:"severity"`
	ProviderType interfaces.OAuthProviderType `json:"provider_type,omitempty"`
}

// ValidationSeverity defines the severity of validation issues
type ValidationSeverity string

const (
	SeverityError   ValidationSeverity = "error"
	SeverityWarning ValidationSeverity = "warning"
	SeverityInfo    ValidationSeverity = "info"
)

// ProviderValidator provides validation for OAuth provider configurations
type ProviderValidator struct {
	httpClient *http.Client
}

// NewProviderValidator creates a new provider validator
func NewProviderValidator() *ProviderValidator {
	return &ProviderValidator{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ValidateProvider validates an OAuth provider configuration
func (pv *ProviderValidator) ValidateProvider(ctx context.Context, config *interfaces.OAuthProviderConfig, level ValidationLevel) *ValidationResult {
	start := time.Now()

	result := &ValidationResult{
		Valid:       true,
		Errors:      make([]ValidationError, 0),
		Warnings:    make([]ValidationError, 0),
		Level:       level,
		Provider:    string(config.Type),
		ValidatedAt: start,
	}

	// Basic validation
	pv.validateBasicFields(config, result)

	if level >= ValidationLevelExtended {
		// Extended validation with connectivity checks
		pv.validateConnectivity(ctx, config, result)
		pv.validateURLFormat(config, result)
		pv.validateScopes(config, result)
	}

	if level >= ValidationLevelFull {
		// Full validation with OAuth flow simulation
		pv.validateOAuthFlow(ctx, config, result)
	}

	// Set overall validity
	result.Valid = len(result.Errors) == 0
	result.Duration = time.Since(start)

	return result
}

// validateBasicFields performs basic field validation
func (pv *ProviderValidator) validateBasicFields(config *interfaces.OAuthProviderConfig, result *ValidationResult) {
	// Check required fields
	if config.Type == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:    "type",
			Code:     "REQUIRED_FIELD",
			Message:  "Provider type is required",
			Severity: SeverityError,
		})
	}

	if config.ClientID == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:    "client_id",
			Code:     "REQUIRED_FIELD",
			Message:  "Client ID is required",
			Severity: SeverityError,
		})
	}

	if config.ClientSecret == "" && config.EncryptedSecret == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:    "client_secret",
			Code:     "REQUIRED_FIELD",
			Message:  "Client secret is required",
			Severity: SeverityError,
		})
	}

	if config.RedirectURL == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:    "redirect_url",
			Code:     "REQUIRED_FIELD",
			Message:  "Redirect URL is required",
			Severity: SeverityError,
		})
	}

	if config.AuthURL == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:    "auth_url",
			Code:     "REQUIRED_FIELD",
			Message:  "Authorization URL is required",
			Severity: SeverityError,
		})
	}

	if config.TokenURL == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:    "token_url",
			Code:     "REQUIRED_FIELD",
			Message:  "Token URL is required",
			Severity: SeverityError,
		})
	}

	if config.UserInfoURL == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:    "user_info_url",
			Code:     "REQUIRED_FIELD",
			Message:  "User info URL is required",
			Severity: SeverityError,
		})
	}

	// Validate redirect URL format
	if config.RedirectURL != "" {
		if !pv.isValidURL(config.RedirectURL) {
			result.Errors = append(result.Errors, ValidationError{
				Field:    "redirect_url",
				Code:     "INVALID_URL",
				Message:  "Redirect URL format is invalid",
				Severity: SeverityError,
			})
		}
	}

	// Check for common security issues
	pv.validateSecurity(config, result)
}

// validateConnectivity checks if OAuth endpoints are reachable
func (pv *ProviderValidator) validateConnectivity(ctx context.Context, config *interfaces.OAuthProviderConfig, result *ValidationResult) {
	urls := map[string]string{
		"auth_url":      config.AuthURL,
		"token_url":     config.TokenURL,
		"user_info_url": config.UserInfoURL,
	}

	for field, urlStr := range urls {
		if urlStr == "" {
			continue
		}

		if !pv.checkURLReachability(ctx, urlStr) {
			result.Warnings = append(result.Warnings, ValidationError{
				Field:    field,
				Code:     "UNREACHABLE_URL",
				Message:  fmt.Sprintf("URL %s is not reachable", urlStr),
				Severity: SeverityWarning,
			})
		}
	}
}

// validateURLFormat validates URL formats for OAuth endpoints
func (pv *ProviderValidator) validateURLFormat(config *interfaces.OAuthProviderConfig, result *ValidationResult) {
	urls := map[string]string{
		"auth_url":      config.AuthURL,
		"token_url":     config.TokenURL,
		"user_info_url": config.UserInfoURL,
	}

	for field, urlStr := range urls {
		if urlStr == "" {
			continue
		}

		if !pv.isValidURL(urlStr) {
			result.Errors = append(result.Errors, ValidationError{
				Field:    field,
				Code:     "INVALID_URL",
				Message:  fmt.Sprintf("Invalid URL format: %s", urlStr),
				Severity: SeverityError,
			})
		}

		// Check for HTTPS requirement
		if !strings.HasPrefix(urlStr, "https://") {
			result.Warnings = append(result.Warnings, ValidationError{
				Field:    field,
				Code:     "INSECURE_URL",
				Message:  "URL should use HTTPS for security",
				Severity: SeverityWarning,
			})
		}
	}
}

// validateScopes validates OAuth scopes
func (pv *ProviderValidator) validateScopes(config *interfaces.OAuthProviderConfig, result *ValidationResult) {
	if len(config.Scopes) == 0 {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:    "scopes",
			Code:     "NO_SCOPES",
			Message:  "No OAuth scopes specified",
			Severity: SeverityWarning,
		})
		return
	}

	// Validate scope format based on provider type
	validScopes := pv.getValidScopesForProvider(config.Type)
	if len(validScopes) > 0 {
		for _, scope := range config.Scopes {
			if !pv.containsString(validScopes, scope) {
				result.Warnings = append(result.Warnings, ValidationError{
					Field:    "scopes",
					Code:     "UNKNOWN_SCOPE",
					Message:  fmt.Sprintf("Unknown scope '%s' for provider %s", scope, config.Type),
					Severity: SeverityWarning,
				})
			}
		}
	}
}

// validateOAuthFlow performs OAuth flow simulation
func (pv *ProviderValidator) validateOAuthFlow(ctx context.Context, config *interfaces.OAuthProviderConfig, result *ValidationResult) {
	// Create OAuth2 config
	oauth2Config := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       config.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  config.AuthURL,
			TokenURL: config.TokenURL,
		},
	}

	// Test authorization URL generation
	authURL := oauth2Config.AuthCodeURL("test-state", oauth2.AccessTypeOffline)
	if authURL == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:    "auth_url",
			Code:     "AUTH_URL_GENERATION_FAILED",
			Message:  "Failed to generate authorization URL",
			Severity: SeverityError,
		})
	}

	// Additional OAuth flow validation could be added here
	// For example, testing token exchange with test credentials
}

// validateSecurity checks for common security issues
func (pv *ProviderValidator) validateSecurity(config *interfaces.OAuthProviderConfig, result *ValidationResult) {
	// Check for weak client secrets
	if config.ClientSecret != "" && len(config.ClientSecret) < 16 {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:    "client_secret",
			Code:     "WEAK_SECRET",
			Message:  "Client secret appears to be weak (less than 16 characters)",
			Severity: SeverityWarning,
		})
	}

	// Check for localhost redirect URLs in production-like configs
	if config.RedirectURL != "" && strings.Contains(config.RedirectURL, "localhost") {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:    "redirect_url",
			Code:     "LOCALHOST_REDIRECT",
			Message:  "Localhost redirect URL may not be suitable for production",
			Severity: SeverityWarning,
		})
	}

	// Check for missing PKCE support for public clients
	if !config.SupportPKCE {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:    "support_pkce",
			Code:     "NO_PKCE_SUPPORT",
			Message:  "PKCE support is recommended for enhanced security",
			Severity: SeverityWarning,
		})
	}
}

// isValidURL checks if a string is a valid URL
func (pv *ProviderValidator) isValidURL(urlStr string) bool {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	return parsedURL.Scheme != "" && parsedURL.Host != ""
}

// checkURLReachability checks if a URL is reachable
func (pv *ProviderValidator) checkURLReachability(ctx context.Context, urlStr string) bool {
	req, err := http.NewRequestWithContext(ctx, "HEAD", urlStr, nil)
	if err != nil {
		return false
	}

	resp, err := pv.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Consider 2xx and 3xx status codes as reachable
	return resp.StatusCode < 400
}

// getValidScopesForProvider returns valid scopes for a provider type
func (pv *ProviderValidator) getValidScopesForProvider(providerType interfaces.OAuthProviderType) []string {
	switch providerType {
	case interfaces.ProviderTypeGitHub:
		return []string{"user", "user:email", "read:user", "repo", "public_repo", "notifications", "gist"}
	case interfaces.ProviderTypeGoogle:
		return []string{"openid", "profile", "email", "https://www.googleapis.com/auth/userinfo.profile"}
	case interfaces.ProviderTypeFacebook:
		return []string{"email", "public_profile", "user_friends"}
	case interfaces.ProviderTypeWechat:
		return []string{"snsapi_login", "snsapi_userinfo"}
	case interfaces.ProviderTypeQQ:
		return []string{"get_user_info", "list_album", "upload_pic"}
	default:
		return []string{} // Unknown provider, allow any scopes
	}
}

// containsString checks if a slice contains a string
func (pv *ProviderValidator) containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// ValidateProviderName validates a provider name
func (pv *ProviderValidator) ValidateProviderName(name string) *ValidationResult {
	result := &ValidationResult{
		Valid:       true,
		Errors:      make([]ValidationError, 0),
		Warnings:    make([]ValidationError, 0),
		Level:       ValidationLevelBasic,
		ValidatedAt: time.Now(),
	}

	if name == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:    "name",
			Code:     "REQUIRED_FIELD",
			Message:  "Provider name is required",
			Severity: SeverityError,
		})
	}

	// Validate name format
	nameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !nameRegex.MatchString(name) {
		result.Errors = append(result.Errors, ValidationError{
			Field:    "name",
			Code:     "INVALID_FORMAT",
			Message:  "Provider name can only contain letters, numbers, underscores, and hyphens",
			Severity: SeverityError,
		})
	}

	// Check name length
	if len(name) > 50 {
		result.Errors = append(result.Errors, ValidationError{
			Field:    "name",
			Code:     "TOO_LONG",
			Message:  "Provider name must be 50 characters or less",
			Severity: SeverityError,
		})
	}

	result.Valid = len(result.Errors) == 0
	return result
}

// Global provider validator instance
var globalProviderValidator *ProviderValidator

// GetGlobalProviderValidator returns the global provider validator instance
func GetGlobalProviderValidator() *ProviderValidator {
	if globalProviderValidator == nil {
		globalProviderValidator = NewProviderValidator()
	}
	return globalProviderValidator
}