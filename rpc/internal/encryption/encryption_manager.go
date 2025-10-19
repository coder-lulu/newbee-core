package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/coder-lulu/newbee-core/rpc/internal/interfaces"
)

// EncryptionAlgorithm defines the encryption algorithm type
type EncryptionAlgorithm string

const (
	AlgorithmAES256GCM EncryptionAlgorithm = "AES-256-GCM"
	AlgorithmAES128GCM EncryptionAlgorithm = "AES-128-GCM"
)

// EncryptionKey represents an encryption key with metadata
type EncryptionKey struct {
	ID        string              `json:"id"`
	Key       []byte              `json:"-"` // Never include in JSON
	Algorithm EncryptionAlgorithm `json:"algorithm"`
	CreatedAt time.Time           `json:"created_at"`
	ExpiresAt *time.Time          `json:"expires_at,omitempty"`
	Active    bool                `json:"active"`
}

// IsExpired checks if the encryption key has expired
func (k *EncryptionKey) IsExpired() bool {
	return k.ExpiresAt != nil && time.Now().After(*k.ExpiresAt)
}

// EncryptionManager manages encryption keys and operations
type EncryptionManager struct {
	mu         sync.RWMutex
	keys       map[string]*EncryptionKey
	activeKey  *EncryptionKey
	defaultAlg EncryptionAlgorithm
}

// NewEncryptionManager creates a new encryption manager
func NewEncryptionManager() *EncryptionManager {
	return &EncryptionManager{
		keys:       make(map[string]*EncryptionKey),
		defaultAlg: AlgorithmAES256GCM,
	}
}

// AddKey adds an encryption key to the manager
func (em *EncryptionManager) AddKey(id string, key []byte, algorithm EncryptionAlgorithm) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	if len(key) == 0 {
		return errors.New("encryption key cannot be empty")
	}

	// Validate key length for algorithm
	if err := em.validateKeyLength(key, algorithm); err != nil {
		return err
	}

	encKey := &EncryptionKey{
		ID:        id,
		Key:       make([]byte, len(key)),
		Algorithm: algorithm,
		CreatedAt: time.Now(),
		Active:    true,
	}
	copy(encKey.Key, key)

	em.keys[id] = encKey

	// Set as active key if this is the first key or no active key exists
	if em.activeKey == nil {
		em.activeKey = encKey
	}

	return nil
}

// SetActiveKey sets the active encryption key
func (em *EncryptionManager) SetActiveKey(keyID string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	key, exists := em.keys[keyID]
	if !exists {
		return fmt.Errorf("encryption key %s not found", keyID)
	}

	if key.IsExpired() {
		return fmt.Errorf("encryption key %s has expired", keyID)
	}

	if !key.Active {
		return fmt.Errorf("encryption key %s is not active", keyID)
	}

	em.activeKey = key
	return nil
}

// GetActiveKey returns the current active encryption key
func (em *EncryptionManager) GetActiveKey() *EncryptionKey {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return em.activeKey
}

// GetKey returns an encryption key by ID
func (em *EncryptionManager) GetKey(keyID string) (*EncryptionKey, error) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	key, exists := em.keys[keyID]
	if !exists {
		return nil, fmt.Errorf("encryption key %s not found", keyID)
	}

	return key, nil
}

// EncryptData encrypts data using the active encryption key
func (em *EncryptionManager) EncryptData(plaintext []byte) (*EncryptionResult, error) {
	em.mu.RLock()
	activeKey := em.activeKey
	em.mu.RUnlock()

	if activeKey == nil {
		return nil, errors.New("no active encryption key")
	}

	if activeKey.IsExpired() {
		return nil, errors.New("active encryption key has expired")
	}

	ciphertext, err := em.encryptWithKey(plaintext, activeKey)
	if err != nil {
		return nil, err
	}

	return &EncryptionResult{
		Ciphertext:    ciphertext,
		KeyID:         activeKey.ID,
		Algorithm:     activeKey.Algorithm,
		EncryptedAt:   time.Now(),
	}, nil
}

// DecryptData decrypts data using the specified key
func (em *EncryptionManager) DecryptData(result *EncryptionResult) ([]byte, error) {
	key, err := em.GetKey(result.KeyID)
	if err != nil {
		return nil, err
	}

	return em.decryptWithKey(result.Ciphertext, key)
}

// EncryptString encrypts a string and returns base64-encoded result
func (em *EncryptionManager) EncryptString(plaintext string) (string, string, error) {
	result, err := em.EncryptData([]byte(plaintext))
	if err != nil {
		return "", "", err
	}

	encoded := base64.StdEncoding.EncodeToString(result.Ciphertext)
	return encoded, result.KeyID, nil
}

// DecryptString decrypts a base64-encoded string
func (em *EncryptionManager) DecryptString(ciphertext, keyID string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %v", err)
	}

	key, err := em.GetKey(keyID)
	if err != nil {
		return "", err
	}

	plaintext, err := em.decryptWithKey(decoded, key)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// RotateKeys rotates encryption keys by setting a new active key
func (em *EncryptionManager) RotateKeys(newKeyID string, newKey []byte) error {
	// Add the new key
	if err := em.AddKey(newKeyID, newKey, em.defaultAlg); err != nil {
		return fmt.Errorf("failed to add new key: %v", err)
	}

	// Set as active
	if err := em.SetActiveKey(newKeyID); err != nil {
		return fmt.Errorf("failed to set new key as active: %v", err)
	}

	return nil
}

// CleanupExpiredKeys removes expired keys from memory
func (em *EncryptionManager) CleanupExpiredKeys() int {
	em.mu.Lock()
	defer em.mu.Unlock()

	cleaned := 0
	for id, key := range em.keys {
		if key.IsExpired() && key != em.activeKey {
			delete(em.keys, id)
			cleaned++
		}
	}

	return cleaned
}

// ListKeys returns information about all keys (without the actual key data)
func (em *EncryptionManager) ListKeys() []*EncryptionKey {
	em.mu.RLock()
	defer em.mu.RUnlock()

	keys := make([]*EncryptionKey, 0, len(em.keys))
	for _, key := range em.keys {
		// Create a copy without the actual key data
		keyCopy := &EncryptionKey{
			ID:        key.ID,
			Algorithm: key.Algorithm,
			CreatedAt: key.CreatedAt,
			ExpiresAt: key.ExpiresAt,
			Active:    key.Active,
		}
		keys = append(keys, keyCopy)
	}

	return keys
}

// validateKeyLength validates that the key length matches the algorithm requirements
func (em *EncryptionManager) validateKeyLength(key []byte, algorithm EncryptionAlgorithm) error {
	switch algorithm {
	case AlgorithmAES256GCM:
		if len(key) != 32 {
			return fmt.Errorf("AES-256-GCM requires 32-byte key, got %d bytes", len(key))
		}
	case AlgorithmAES128GCM:
		if len(key) != 16 {
			return fmt.Errorf("AES-128-GCM requires 16-byte key, got %d bytes", len(key))
		}
	default:
		return fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
	return nil
}

// encryptWithKey encrypts data using the specified key
func (em *EncryptionManager) encryptWithKey(plaintext []byte, key *EncryptionKey) ([]byte, error) {
	switch key.Algorithm {
	case AlgorithmAES256GCM, AlgorithmAES128GCM:
		return em.encryptAESGCM(plaintext, key.Key)
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", key.Algorithm)
	}
}

// decryptWithKey decrypts data using the specified key
func (em *EncryptionManager) decryptWithKey(ciphertext []byte, key *EncryptionKey) ([]byte, error) {
	switch key.Algorithm {
	case AlgorithmAES256GCM, AlgorithmAES128GCM:
		return em.decryptAESGCM(ciphertext, key.Key)
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", key.Algorithm)
	}
}

// encryptAESGCM encrypts data using AES-GCM
func (em *EncryptionManager) encryptAESGCM(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decryptAESGCM decrypts data using AES-GCM
func (em *EncryptionManager) decryptAESGCM(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// EncryptionResult represents the result of an encryption operation
type EncryptionResult struct {
	Ciphertext  []byte              `json:"ciphertext"`
	KeyID       string              `json:"key_id"`
	Algorithm   EncryptionAlgorithm `json:"algorithm"`
	EncryptedAt time.Time           `json:"encrypted_at"`
}

// ProviderEncryptionService provides encryption services specifically for OAuth providers
type ProviderEncryptionService struct {
	encManager *EncryptionManager
}

// NewProviderEncryptionService creates a new provider encryption service
func NewProviderEncryptionService(manager *EncryptionManager) *ProviderEncryptionService {
	return &ProviderEncryptionService{
		encManager: manager,
	}
}

// EncryptProviderSecret encrypts a provider's client secret
func (pes *ProviderEncryptionService) EncryptProviderSecret(secret string) (string, string, error) {
	if secret == "" {
		return "", "", errors.New("secret cannot be empty")
	}

	return pes.encManager.EncryptString(secret)
}

// DecryptProviderSecret decrypts a provider's client secret
func (pes *ProviderEncryptionService) DecryptProviderSecret(encryptedSecret, keyID string) (string, error) {
	if encryptedSecret == "" {
		return "", errors.New("encrypted secret cannot be empty")
	}
	
	if keyID == "" {
		return "", errors.New("key ID cannot be empty")
	}

	return pes.encManager.DecryptString(encryptedSecret, keyID)
}

// UpdateProviderConfigWithEncryption updates provider config with encryption
func (pes *ProviderEncryptionService) UpdateProviderConfigWithEncryption(config *interfaces.OAuthProviderConfig) error {
	if config.ClientSecret == "" {
		return nil // Nothing to encrypt
	}

	encryptedSecret, keyID, err := pes.EncryptProviderSecret(config.ClientSecret)
	if err != nil {
		return fmt.Errorf("failed to encrypt provider secret: %v", err)
	}

	// Store the encrypted secret and clear the plain text
	config.EncryptedSecret = encryptedSecret
	config.EncryptionKeyID = keyID
	config.ClientSecret = "" // Clear plain text secret

	return nil
}

// DecryptProviderConfig decrypts sensitive fields in provider config
func (pes *ProviderEncryptionService) DecryptProviderConfig(config *interfaces.OAuthProviderConfig) error {
	if config.EncryptedSecret == "" || config.EncryptionKeyID == "" {
		return nil // Nothing to decrypt
	}

	decryptedSecret, err := pes.DecryptProviderSecret(config.EncryptedSecret, config.EncryptionKeyID)
	if err != nil {
		return fmt.Errorf("failed to decrypt provider secret: %v", err)
	}

	config.ClientSecret = decryptedSecret
	return nil
}

// Global encryption manager and service instances
var (
	globalEncryptionManager       *EncryptionManager
	globalProviderEncryptionService *ProviderEncryptionService
)

// InitGlobalEncryption initializes the global encryption services
func InitGlobalEncryption() {
	globalEncryptionManager = NewEncryptionManager()
	globalProviderEncryptionService = NewProviderEncryptionService(globalEncryptionManager)
}

// GetGlobalEncryptionManager returns the global encryption manager
func GetGlobalEncryptionManager() *EncryptionManager {
	if globalEncryptionManager == nil {
		InitGlobalEncryption()
	}
	return globalEncryptionManager
}

// GetGlobalProviderEncryptionService returns the global provider encryption service
func GetGlobalProviderEncryptionService() *ProviderEncryptionService {
	if globalProviderEncryptionService == nil {
		InitGlobalEncryption()
	}
	return globalProviderEncryptionService
}