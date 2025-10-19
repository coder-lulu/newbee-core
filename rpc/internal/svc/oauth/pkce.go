package oauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// PKCEChallenge represents a PKCE challenge
type PKCEChallenge struct {
	Verifier  string `json:"verifier"`   // code_verifier
	Challenge string `json:"challenge"`  // code_challenge  
	Method    string `json:"method"`     // S256
}

// PKCEManager manages PKCE challenges for OAuth flows
type PKCEManager struct {
	redis      redis.UniversalClient
	expiration time.Duration
}

// NewPKCEManager creates a new PKCE manager
func NewPKCEManager(redis redis.UniversalClient) *PKCEManager {
	return &PKCEManager{
		redis:      redis,
		expiration: 10 * time.Minute, // PKCE challenges expire in 10 minutes
	}
}

// GenerateChallenge generates a new PKCE challenge
func (p *PKCEManager) GenerateChallenge() (*PKCEChallenge, error) {
	verifier, err := generateCodeVerifier()
	if err != nil {
		return nil, fmt.Errorf("failed to generate code verifier: %v", err)
	}

	challenge := sha256Challenge(verifier)

	return &PKCEChallenge{
		Verifier:  verifier,
		Challenge: challenge,
		Method:    "S256",
	}, nil
}

// StoreChallenge stores a PKCE challenge associated with a state
func (p *PKCEManager) StoreChallenge(state string, challenge *PKCEChallenge) error {
	key := fmt.Sprintf("pkce:challenge:%s", state)
	
	data, err := json.Marshal(challenge)
	if err != nil {
		return fmt.Errorf("failed to marshal pkce challenge: %v", err)
	}

	err = p.redis.Set(context.Background(), key, data, p.expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to store pkce challenge: %v", err)
	}

	return nil
}

// GetChallenge retrieves a PKCE challenge by state
func (p *PKCEManager) GetChallenge(state string) (*PKCEChallenge, error) {
	key := fmt.Sprintf("pkce:challenge:%s", state)
	
	data, err := p.redis.Get(context.Background(), key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("pkce challenge not found")
		}
		return nil, fmt.Errorf("failed to get pkce challenge: %v", err)
	}

	var challenge PKCEChallenge
	err = json.Unmarshal([]byte(data), &challenge)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal pkce challenge: %v", err)
	}

	return &challenge, nil
}

// DeleteChallenge deletes a PKCE challenge
func (p *PKCEManager) DeleteChallenge(state string) error {
	key := fmt.Sprintf("pkce:challenge:%s", state)
	
	err := p.redis.Del(context.Background(), key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete pkce challenge: %v", err)
	}

	return nil
}

// ValidateChallenge validates a PKCE challenge
func (p *PKCEManager) ValidateChallenge(state string, codeVerifier string) error {
	challenge, err := p.GetChallenge(state)
	if err != nil {
		return err
	}

	if challenge.Verifier != codeVerifier {
		return fmt.Errorf("invalid code verifier")
	}

	// Delete the challenge after successful validation
	_ = p.DeleteChallenge(state)

	return nil
}

// generateCodeVerifier generates a cryptographically secure code verifier
func generateCodeVerifier() (string, error) {
	// Generate 32 random bytes (256 bits)
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	// Encode as base64url (RFC 7636)
	verifier := base64.RawURLEncoding.EncodeToString(bytes)
	
	return verifier, nil
}

// sha256Challenge creates a SHA256 challenge from a verifier
func sha256Challenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])
	return challenge
}

// Global PKCE manager instance
var globalPKCEManager *PKCEManager

// InitGlobalPKCEManager initializes the global PKCE manager
func InitGlobalPKCEManager(redis redis.UniversalClient) {
	globalPKCEManager = NewPKCEManager(redis)
}

// GetGlobalPKCEManager returns the global PKCE manager instance
func GetGlobalPKCEManager() *PKCEManager {
	return globalPKCEManager
}