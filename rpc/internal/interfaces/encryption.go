package interfaces

import (
	"context"
	"time"
)

// AuditRecord represents an audit log record for encryption operations
type AuditRecord struct {
	// Timestamp when the operation occurred
	Timestamp time.Time `json:"timestamp"`
	
	// Operation type (e.g., "encrypt", "decrypt", "generate_key", "rotate_key")
	Operation string `json:"operation"`
	
	// User ID who performed the operation
	UserID string `json:"user_id,omitempty"`
	
	// Key ID involved in the operation
	KeyID string `json:"key_id,omitempty"`
	
	// Source IP address
	SourceIP string `json:"source_ip,omitempty"`
	
	// Whether the operation was successful
	Success bool `json:"success"`
	
	// Error message if operation failed
	ErrorMessage string `json:"error_message,omitempty"`
	
	// Additional metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// EncryptionConfig represents configuration for encryption operations
type EncryptionConfig struct {
	// Default encryption algorithm
	DefaultAlgorithm string `json:"default_algorithm"`
	
	// Key rotation interval
	KeyRotationInterval time.Duration `json:"key_rotation_interval"`
	
	// Maximum key age before rotation
	MaxKeyAge time.Duration `json:"max_key_age"`
	
	// Enable audit logging
	EnableAuditLogging bool `json:"enable_audit_logging"`
	
	// Audit log retention period
	AuditRetentionPeriod time.Duration `json:"audit_retention_period"`
	
	// Key storage backend configuration
	KeyStorageConfig map[string]interface{} `json:"key_storage_config"`
	
	// Encryption strength settings
	EncryptionStrength EncryptionStrength `json:"encryption_strength"`
	
	// Compliance requirements
	ComplianceRequirements []string `json:"compliance_requirements"`
}

// EncryptionStrength defines encryption strength levels
type EncryptionStrength string

const (
	EncryptionStrengthLow    EncryptionStrength = "low"
	EncryptionStrengthMedium EncryptionStrength = "medium"
	EncryptionStrengthHigh   EncryptionStrength = "high"
)

// PolicyManager manages encryption policies
type PolicyManager interface {
	// GetPolicy retrieves encryption policy for given data classification
	GetPolicy(ctx context.Context, classification DataClassification) (*EncryptionPolicy, error)
	
	// SetPolicy sets encryption policy for given data classification
	SetPolicy(ctx context.Context, classification DataClassification, policy *EncryptionPolicy) error
	
	// DeletePolicy removes encryption policy for given data classification
	DeletePolicy(ctx context.Context, classification DataClassification) error
	
	// ListPolicies lists all encryption policies
	ListPolicies(ctx context.Context) (map[DataClassification]*EncryptionPolicy, error)
	
	// ValidatePolicy validates an encryption policy
	ValidatePolicy(ctx context.Context, policy *EncryptionPolicy) error
	
	// GetDefaultPolicy returns the default encryption policy
	GetDefaultPolicy(ctx context.Context) (*EncryptionPolicy, error)
}

// DataClassification represents data sensitivity levels
type DataClassification string

const (
	DataClassificationPublic       DataClassification = "public"
	DataClassificationInternal     DataClassification = "internal"
	DataClassificationConfidential DataClassification = "confidential"
	DataClassificationRestricted   DataClassification = "restricted"
	DataClassificationTopSecret    DataClassification = "top_secret"
)

// EncryptionPolicy defines encryption requirements for different data classifications
type EncryptionPolicy struct {
	// Data classification this policy applies to
	Classification DataClassification `json:"classification"`
	
	// Required encryption algorithm
	Algorithm string `json:"algorithm"`
	
	// Minimum key size in bits
	MinKeySize int `json:"min_key_size"`
	
	// Key rotation frequency
	KeyRotationFrequency time.Duration `json:"key_rotation_frequency"`
	
	// Whether encryption is required
	EncryptionRequired bool `json:"encryption_required"`
	
	// Whether key escrow is required
	KeyEscrowRequired bool `json:"key_escrow_required"`
	
	// Compliance standards this policy satisfies
	ComplianceStandards []string `json:"compliance_standards"`
	
	// Additional policy attributes
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	
	// Policy creation and update timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EncryptionManager manages encryption operations and key lifecycle
type EncryptionManager interface {
	// Key management
	GenerateKey(ctx context.Context, keyType KeyType, algorithm Algorithm) (*EncryptionKey, error)
	GetKey(ctx context.Context, keyID string) (*EncryptionKey, error)
	DeleteKey(ctx context.Context, keyID string) error
	RotateKey(ctx context.Context, keyID string) (*EncryptionKey, error)
	ListKeys(ctx context.Context, filters map[string]interface{}) ([]*EncryptionKey, error)
	
	// Encryption operations
	Encrypt(ctx context.Context, data []byte, keyID string) (*EncryptionResult, error)
	Decrypt(ctx context.Context, request *DecryptionRequest) ([]byte, error)
	
	// Bulk operations
	BulkEncrypt(ctx context.Context, requests []*EncryptionRequest) ([]*EncryptionResult, error)
	BulkDecrypt(ctx context.Context, requests []*DecryptionRequest) ([][]byte, error)
	
	// Key derivation
	DeriveKey(ctx context.Context, masterKeyID string, context []byte, keyLength int) (*EncryptionKey, error)
	
	// Audit and monitoring
	AuditKeyUsage(ctx context.Context, keyID string, startTime, endTime time.Time) ([]AuditRecord, error)
	GetMetrics(ctx context.Context) (*EncryptionMetrics, error)
	
	// Health and status
	HealthCheck(ctx context.Context) error
	GetStatus(ctx context.Context) (*ManagerStatus, error)
}

// EncryptionKey represents an encryption key with metadata
type EncryptionKey struct {
	// Key identifier
	ID string `json:"id"`
	
	// Key type and algorithm
	Type      KeyType   `json:"type"`
	Algorithm Algorithm `json:"algorithm"`
	
	// Key material (encrypted when stored)
	KeyMaterial []byte `json:"key_material,omitempty"`
	
	// Key size in bits
	KeySize int `json:"key_size"`
	
	// Key status
	Status KeyStatus `json:"status"`
	
	// Usage permissions
	Usage []KeyUsage `json:"usage"`
	
	// Key metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	
	// Lifecycle timestamps
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	RotatedAt  *time.Time `json:"rotated_at,omitempty"`
	
	// Key derivation information
	DerivedFrom string `json:"derived_from,omitempty"`
	
	// Compliance and policy
	Classification DataClassification `json:"classification"`
	PolicyID       string             `json:"policy_id,omitempty"`
}

// KeyType defines the type of encryption key
type KeyType string

const (
	AESKey    KeyType = "aes"
	RSAKey    KeyType = "rsa"
	ECDSAKey  KeyType = "ecdsa"
	HMACKey   KeyType = "hmac"
	CustomKey KeyType = "custom"
)

// Algorithm defines encryption algorithms
type Algorithm string

const (
	AES256GCM    Algorithm = "aes-256-gcm"
	AES128GCM    Algorithm = "aes-128-gcm"
	RSA2048      Algorithm = "rsa-2048"
	RSA4096      Algorithm = "rsa-4096"
	ECDSA256     Algorithm = "ecdsa-p256"
	ECDSA384     Algorithm = "ecdsa-p384"
	HMACSHA256   Algorithm = "hmac-sha256"
	ChaCha20Poly Algorithm = "chacha20-poly1305"
)

// KeyStatus defines the status of an encryption key
type KeyStatus string

const (
	KeyStatusActive     KeyStatus = "active"
	KeyStatusInactive   KeyStatus = "inactive"
	KeyStatusRotating   KeyStatus = "rotating"
	KeyStatusDeprecated KeyStatus = "deprecated"
	KeyStatusRevoked    KeyStatus = "revoked"
)

// KeyUsage defines how a key can be used
type KeyUsage string

const (
	KeyUsageEncrypt KeyUsage = "encrypt"
	KeyUsageDecrypt KeyUsage = "decrypt"
	KeyUsageSign    KeyUsage = "sign"
	KeyUsageVerify  KeyUsage = "verify"
	KeyUsageDerive  KeyUsage = "derive"
)

// EncryptionResult represents the result of an encryption operation
type EncryptionResult struct {
	// Encrypted data
	EncryptedData []byte `json:"encrypted_data"`
	
	// Key ID used for encryption
	KeyID string `json:"key_id"`
	
	// Algorithm used
	Algorithm Algorithm `json:"algorithm"`
	
	// Initialization vector or nonce
	Nonce []byte `json:"nonce,omitempty"`
	
	// Authentication tag (for AEAD algorithms)
	AuthTag []byte `json:"auth_tag,omitempty"`
	
	// Additional authenticated data
	AAD []byte `json:"aad,omitempty"`
	
	// Encryption timestamp
	Timestamp time.Time `json:"timestamp"`
}

// DecryptionRequest represents a request to decrypt data
type DecryptionRequest struct {
	// Encrypted data to decrypt
	EncryptedData []byte `json:"encrypted_data"`
	
	// Key ID to use for decryption
	KeyID string `json:"key_id"`
	
	// Algorithm used for encryption
	Algorithm Algorithm `json:"algorithm"`
	
	// Initialization vector or nonce
	Nonce []byte `json:"nonce,omitempty"`
	
	// Authentication tag (for AEAD algorithms)
	AuthTag []byte `json:"auth_tag,omitempty"`
	
	// Additional authenticated data
	AAD []byte `json:"aad,omitempty"`
}

// EncryptionRequest represents a request to encrypt data
type EncryptionRequest struct {
	// Data to encrypt
	Data []byte `json:"data"`
	
	// Key ID to use for encryption
	KeyID string `json:"key_id"`
	
	// Additional authenticated data
	AAD []byte `json:"aad,omitempty"`
}

// EncryptionMetrics contains metrics about encryption operations
type EncryptionMetrics struct {
	// Operation counts
	TotalEncryptions int64 `json:"total_encryptions"`
	TotalDecryptions int64 `json:"total_decryptions"`
	FailedOperations int64 `json:"failed_operations"`
	
	// Key metrics
	ActiveKeys     int64 `json:"active_keys"`
	DeprecatedKeys int64 `json:"deprecated_keys"`
	RevokedKeys    int64 `json:"revoked_keys"`
	
	// Performance metrics
	AverageEncryptionTime time.Duration `json:"average_encryption_time"`
	AverageDecryptionTime time.Duration `json:"average_decryption_time"`
	
	// Error rates
	EncryptionErrorRate float64 `json:"encryption_error_rate"`
	DecryptionErrorRate float64 `json:"decryption_error_rate"`
	
	// Key rotation metrics
	KeyRotations       int64     `json:"key_rotations"`
	LastKeyRotation    time.Time `json:"last_key_rotation"`
	OverdueRotations   int64     `json:"overdue_rotations"`
	
	// Compliance metrics
	ComplianceViolations int64 `json:"compliance_violations"`
	
	// Collection timestamp
	CollectedAt time.Time `json:"collected_at"`
}

// ManagerStatus represents the status of the encryption manager
type ManagerStatus struct {
	// Overall health status
	Healthy bool `json:"healthy"`
	
	// Component statuses
	KeyStoreStatus    string `json:"key_store_status"`
	AuditLogStatus    string `json:"audit_log_status"`
	PolicyStatus      string `json:"policy_status"`
	
	// Resource usage
	MemoryUsage int64 `json:"memory_usage"`
	CPUUsage    float64 `json:"cpu_usage"`
	
	// Last health check
	LastHealthCheck time.Time `json:"last_health_check"`
	
	// Error information
	Errors []string `json:"errors,omitempty"`
	
	// Version information
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
}

// KeyRotationConfig 密钥轮换配置
type KeyRotationConfig struct {
	Enabled         bool          `json:"enabled"`
	RotationPeriod  time.Duration `json:"rotation_period"`
	MaxUsageCount   int64         `json:"max_usage_count"`
	MaxAge          time.Duration `json:"max_age"`
	AutoRotate      bool          `json:"auto_rotate"`
	NotifyBeforeExp time.Duration `json:"notify_before_exp"`
}