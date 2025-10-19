package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// StateInfo represents OAuth state information
type StateInfo struct {
	UserID     uint64                 `json:"user_id"`
	ProviderID uint64                 `json:"provider_id"`
	Extra      map[string]interface{} `json:"extra"`
	CreatedAt  time.Time             `json:"created_at"`
	ExpiresAt  time.Time             `json:"expires_at"`
}

// StateManager manages OAuth state parameters
type StateManager struct {
	redis      redis.UniversalClient
	secretKey  []byte
	expiration time.Duration
}

// NewStateManager creates a new state manager
func NewStateManager(redis redis.UniversalClient, secretKey []byte) *StateManager {
	return &StateManager{
		redis:      redis,
		secretKey:  secretKey,
		expiration: 10 * time.Minute, // State expires in 10 minutes
	}
}

// GenerateState generates and stores a state parameter
func (s *StateManager) GenerateState(userID uint64, providerID uint64, extra map[string]interface{}) (string, error) {
	// Generate a cryptographically secure random state
	state, err := generateSecureState()
	if err != nil {
		return "", fmt.Errorf("failed to generate state: %v", err)
	}

	// Create state info
	now := time.Now()
	stateInfo := &StateInfo{
		UserID:     userID,
		ProviderID: providerID,
		Extra:      extra,
		CreatedAt:  now,
		ExpiresAt:  now.Add(s.expiration),
	}

	// Store state info in Redis
	key := fmt.Sprintf("oauth:state:%s", state)
	data, err := json.Marshal(stateInfo)
	if err != nil {
		return "", fmt.Errorf("failed to marshal state info: %v", err)
	}

	err = s.redis.Set(context.Background(), key, data, s.expiration).Err()
	if err != nil {
		return "", fmt.Errorf("failed to store state: %v", err)
	}

	return state, nil
}

// ValidateState validates and retrieves state information
func (s *StateManager) ValidateState(state string) (*StateInfo, error) {
	if state == "" {
		return nil, fmt.Errorf("empty state parameter")
	}

	key := fmt.Sprintf("oauth:state:%s", state)
	
	data, err := s.redis.Get(context.Background(), key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("invalid or expired state parameter")
		}
		return nil, fmt.Errorf("failed to get state: %v", err)
	}

	var stateInfo StateInfo
	err = json.Unmarshal([]byte(data), &stateInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal state info: %v", err)
	}

	// Check if state has expired
	if time.Now().After(stateInfo.ExpiresAt) {
		// Clean up expired state
		_ = s.DeleteState(state)
		return nil, fmt.Errorf("state parameter has expired")
	}

	return &stateInfo, nil
}

// DeleteState deletes a state parameter
func (s *StateManager) DeleteState(state string) error {
	key := fmt.Sprintf("oauth:state:%s", state)
	
	err := s.redis.Del(context.Background(), key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete state: %v", err)
	}

	return nil
}

// ConsumeState validates and deletes a state parameter (one-time use)
func (s *StateManager) ConsumeState(state string) (*StateInfo, error) {
	stateInfo, err := s.ValidateState(state)
	if err != nil {
		return nil, err
	}

	// Delete the state after successful validation
	_ = s.DeleteState(state)

	return stateInfo, nil
}

// CleanExpiredStates removes expired state parameters
func (s *StateManager) CleanExpiredStates() error {
	// Use Redis SCAN to find all state keys
	ctx := context.Background()
	pattern := "oauth:state:*"
	
	iter := s.redis.Scan(ctx, 0, pattern, 100).Iterator()
	expiredKeys := make([]string, 0)
	
	for iter.Next(ctx) {
		key := iter.Val()
		
		// Get the state data
		data, err := s.redis.Get(ctx, key).Result()
		if err != nil {
			continue // Skip if error
		}
		
		var stateInfo StateInfo
		err = json.Unmarshal([]byte(data), &stateInfo)
		if err != nil {
			continue // Skip if error
		}
		
		// Check if expired
		if time.Now().After(stateInfo.ExpiresAt) {
			expiredKeys = append(expiredKeys, key)
		}
	}
	
	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan state keys: %v", err)
	}
	
	// Delete expired keys in batches
	if len(expiredKeys) > 0 {
		err := s.redis.Del(ctx, expiredKeys...).Err()
		if err != nil {
			return fmt.Errorf("failed to delete expired states: %v", err)
		}
	}
	
	return nil
}

// GetStateStats returns statistics about stored states
func (s *StateManager) GetStateStats() (map[string]interface{}, error) {
	ctx := context.Background()
	pattern := "oauth:state:*"
	
	var totalStates int64
	var expiredStates int64
	
	iter := s.redis.Scan(ctx, 0, pattern, 100).Iterator()
	
	for iter.Next(ctx) {
		key := iter.Val()
		totalStates++
		
		// Get the state data
		data, err := s.redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}
		
		var stateInfo StateInfo
		err = json.Unmarshal([]byte(data), &stateInfo)
		if err != nil {
			continue
		}
		
		// Check if expired
		if time.Now().After(stateInfo.ExpiresAt) {
			expiredStates++
		}
	}
	
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan state keys: %v", err)
	}
	
	return map[string]interface{}{
		"total_states":   totalStates,
		"expired_states": expiredStates,
		"active_states":  totalStates - expiredStates,
	}, nil
}

// generateSecureState generates a cryptographically secure state parameter
func generateSecureState() (string, error) {
	// Generate 32 random bytes (256 bits)
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	// Encode as base64url for URL safety
	state := base64.RawURLEncoding.EncodeToString(bytes)
	
	return state, nil
}

// Global state manager instance
var globalStateManager *StateManager

// InitGlobalStateManager initializes the global state manager
func InitGlobalStateManager(redis redis.UniversalClient, secretKey []byte) {
	globalStateManager = NewStateManager(redis, secretKey)
}

// GetGlobalStateManager returns the global state manager instance
func GetGlobalStateManager() *StateManager {
	return globalStateManager
}