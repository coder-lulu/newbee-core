package oauthprovider

import (
	"context"
	"fmt"

	"github.com/coder-lulu/newbee-core/rpc/ent/oauthprovider"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

// MigrateEncryption 数据迁移: 加密所有现有的明文client_secret
// 用于从旧版本升级到加密版本时的一次性数据迁移
func MigrateEncryption(ctx context.Context, svcCtx *svc.ServiceContext) error {
	logger := logx.WithContext(ctx)
	
	// 查询所有未加密的Provider (client_secret不为空但encrypted_secret为空)
	providers, err := svcCtx.DB.OauthProvider.Query().
		Where(
			oauthprovider.ClientSecretNEQ(""),
			oauthprovider.EncryptedSecretEQ(""),
		).
		All(ctx)
	
	if err != nil {
		logger.Errorw("Failed to query providers for migration", logx.Field("error", err))
		return fmt.Errorf("failed to query providers: %w", err)
	}
	
	if len(providers) == 0 {
		logger.Infow("No providers need encryption migration")
		return nil
	}
	
	logger.Infow("Starting encryption migration", logx.Field("count", len(providers)))
	
	successCount := 0
	failCount := 0
	
	for _, p := range providers {
		// 加密client_secret
		encrypted, keyID, err := svcCtx.EncryptionService.EncryptProviderSecret(p.ClientSecret)
		if err != nil {
			logger.Errorw("Failed to encrypt provider secret",
				logx.Field("provider_id", p.ID),
				logx.Field("provider_name", p.Name),
				logx.Field("error", err),
			)
			failCount++
			continue
		}
		
		// 更新数据库: 存储加密值并清除明文
		err = svcCtx.DB.OauthProvider.UpdateOneID(p.ID).
			SetEncryptedSecret(encrypted).
			SetEncryptionKeyID(keyID).
			SetClientSecret(""). // 清除明文
			Exec(ctx)
		
		if err != nil {
			logger.Errorw("Failed to update provider with encrypted secret",
				logx.Field("provider_id", p.ID),
				logx.Field("provider_name", p.Name),
				logx.Field("error", err),
			)
			failCount++
			continue
		}
		
		logger.Infow("Successfully migrated provider",
			logx.Field("provider_id", p.ID),
			logx.Field("provider_name", p.Name),
		)
		successCount++
	}
	
	logger.Infow("Encryption migration completed",
		logx.Field("total", len(providers)),
		logx.Field("success", successCount),
		logx.Field("failed", failCount),
	)
	
	if failCount > 0 {
		return fmt.Errorf("migration completed with %d failures out of %d providers", failCount, len(providers))
	}
	
	return nil
}

// ValidateEncryption 验证所有Provider的加密状态
// 检查是否还有未加密的敏感数据
func ValidateEncryption(ctx context.Context, svcCtx *svc.ServiceContext) (map[string]interface{}, error) {
	logger := logx.WithContext(ctx)
	
	// 统计加密状态
	totalCount, err := svcCtx.DB.OauthProvider.Query().Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count total providers: %w", err)
	}
	
	encryptedCount, err := svcCtx.DB.OauthProvider.Query().
		Where(oauthprovider.EncryptedSecretNEQ("")).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count encrypted providers: %w", err)
	}
	
	unencryptedCount, err := svcCtx.DB.OauthProvider.Query().
		Where(
			oauthprovider.ClientSecretNEQ(""),
			oauthprovider.EncryptedSecretEQ(""),
		).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count unencrypted providers: %w", err)
	}
	
	result := map[string]interface{}{
		"total_providers":      totalCount,
		"encrypted_providers":  encryptedCount,
		"unencrypted_providers": unencryptedCount,
		"encryption_rate":      float64(encryptedCount) / float64(totalCount) * 100,
		"all_encrypted":        unencryptedCount == 0,
	}
	
	logger.Infow("Encryption validation result", logx.Field("stats", result))
	
	return result, nil
}

// ReEncryptProvider 重新加密单个Provider (用于密钥轮换)
func ReEncryptProvider(ctx context.Context, svcCtx *svc.ServiceContext, providerID uint64) error {
	logger := logx.WithContext(ctx)
	
	// 获取Provider
	p, err := svcCtx.DB.OauthProvider.Get(ctx, providerID)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}
	
	// 解密旧密钥
	var clientSecret string
	if p.EncryptedSecret != "" && p.EncryptionKeyID != "" {
		decrypted, err := svcCtx.EncryptionService.DecryptProviderSecret(p.EncryptedSecret, p.EncryptionKeyID)
		if err != nil {
			return fmt.Errorf("failed to decrypt old secret: %w", err)
		}
		clientSecret = decrypted
	} else if p.ClientSecret != "" {
		clientSecret = p.ClientSecret
	} else {
		return fmt.Errorf("no secret found for provider %d", providerID)
	}
	
	// 使用新密钥重新加密
	encrypted, keyID, err := svcCtx.EncryptionService.EncryptProviderSecret(clientSecret)
	if err != nil {
		return fmt.Errorf("failed to re-encrypt secret: %w", err)
	}
	
	// 更新数据库
	err = svcCtx.DB.OauthProvider.UpdateOneID(providerID).
		SetEncryptedSecret(encrypted).
		SetEncryptionKeyID(keyID).
		SetClientSecret("").
		Exec(ctx)
	
	if err != nil {
		return fmt.Errorf("failed to update provider: %w", err)
	}
	
	logger.Infow("Successfully re-encrypted provider",
		logx.Field("provider_id", providerID),
		logx.Field("provider_name", p.Name),
		logx.Field("new_key_id", keyID),
	)
	
	return nil
}

// RotateAllKeys 批量密钥轮换 - 重新加密所有已加密的Provider
// 用于定期密钥轮换以增强安全性
func RotateAllKeys(ctx context.Context, svcCtx *svc.ServiceContext) error {
	logger := logx.WithContext(ctx)
	
	// 查询所有已加密的Provider
	providers, err := svcCtx.DB.OauthProvider.Query().
		Where(oauthprovider.EncryptedSecretNEQ("")).
		All(ctx)
	
	if err != nil {
		logger.Errorw("Failed to query providers for key rotation", logx.Field("error", err))
		return fmt.Errorf("failed to query providers: %w", err)
	}
	
	if len(providers) == 0 {
		logger.Infow("No providers need key rotation")
		return nil
	}
	
	logger.Infow("Starting batch key rotation", logx.Field("count", len(providers)))
	
	successCount := 0
	failCount := 0
	
	for _, p := range providers {
		err := ReEncryptProvider(ctx, svcCtx, p.ID)
		if err != nil {
			logger.Errorw("Failed to rotate key for provider",
				logx.Field("provider_id", p.ID),
				logx.Field("provider_name", p.Name),
				logx.Field("error", err),
			)
			failCount++
			continue
		}
		successCount++
	}
	
	logger.Infow("Batch key rotation completed",
		logx.Field("total", len(providers)),
		logx.Field("success", successCount),
		logx.Field("failed", failCount),
	)
	
	if failCount > 0 {
		return fmt.Errorf("key rotation completed with %d failures out of %d providers", failCount, len(providers))
	}
	
	return nil
}
