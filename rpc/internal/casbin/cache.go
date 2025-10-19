package casbin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/coder-lulu/newbee-common/v2/orm/ent/entctx/tenantctx"
	"github.com/zeromicro/go-zero/core/logx"
)

// CacheManager 权限缓存管理器
type CacheManager struct {
	redis  redis.UniversalClient
	logger logx.Logger
	
	// 缓存配置
	permissionTTL time.Duration // 权限检查结果缓存时长
	ruleTTL       time.Duration // 规则数据缓存时长
	userRoleTTL   time.Duration // 用户角色缓存时长
}

// PermissionResult 权限检查结果
type PermissionResult struct {
	Allowed       bool      `json:"allowed"`
	Reason        string    `json:"reason"`
	AppliedRules  []string  `json:"applied_rules"`
	CachedAt      time.Time `json:"cached_at"`
	FromCache     bool      `json:"from_cache"`
}

// NewCacheManager 创建新的缓存管理器
func NewCacheManager(redis redis.UniversalClient, logger logx.Logger) *CacheManager {
	return &CacheManager{
		redis:  redis,
		logger: logger,
		
		// 默认缓存时间配置
		permissionTTL: 5 * time.Minute,  // 权限结果缓存5分钟
		ruleTTL:       30 * time.Minute, // 规则数据缓存30分钟 
		userRoleTTL:   10 * time.Minute, // 用户角色缓存10分钟
	}
}

// GetPermissionFromCache 从缓存获取权限检查结果
func (c *CacheManager) GetPermissionFromCache(ctx context.Context, subject, object, action, serviceName string) (*PermissionResult, error) {
	// 🔥 获取租户ID - 确保多租户隔离安全
	tenantID := tenantctx.GetTenantIDFromCtx(ctx)
	
	// 构建缓存key
	cacheKey := c.buildPermissionCacheKey(tenantID, subject, object, action, serviceName)
	
	// 从Redis获取数据
	data, err := c.redis.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		// 缓存不存在
		return nil, nil
	}
	if err != nil {
		c.logger.Errorf("Get permission cache failed: %v", err)
		return nil, err
	}
	
	// 解析缓存数据
	var result PermissionResult
	err = json.Unmarshal([]byte(data), &result)
	if err != nil {
		c.logger.Errorf("Unmarshal permission cache failed: %v", err)
		return nil, err
	}
	
	// 标记来源为缓存
	result.FromCache = true
	
	return &result, nil
}

// SetPermissionToCache 将权限检查结果存入缓存
func (c *CacheManager) SetPermissionToCache(ctx context.Context, subject, object, action, serviceName string, result *PermissionResult) error {
	// 🔥 获取租户ID - 确保多租户隔离安全
	tenantID := tenantctx.GetTenantIDFromCtx(ctx)
	
	// 构建缓存key
	cacheKey := c.buildPermissionCacheKey(tenantID, subject, object, action, serviceName)
	
	// 设置缓存时间戳
	result.CachedAt = time.Now()
	result.FromCache = false
	
	// 序列化数据
	data, err := json.Marshal(result)
	if err != nil {
		c.logger.Errorf("Marshal permission cache failed: %v", err)
		return err
	}
	
	// 存入Redis
	err = c.redis.Set(ctx, cacheKey, data, c.permissionTTL).Err()
	if err != nil {
		c.logger.Errorf("Set permission cache failed: %v", err)
		return err
	}
	
	return nil
}

// GetUserRolesFromCache 从缓存获取用户角色
func (c *CacheManager) GetUserRolesFromCache(ctx context.Context, userID string) ([]string, error) {
	// 🔥 获取租户ID - 确保多租户隔离安全
	tenantID := tenantctx.GetTenantIDFromCtx(ctx)
	
	// 构建缓存key
	cacheKey := c.buildUserRoleCacheKey(tenantID, userID)
	
	// 从Redis获取数据
	result, err := c.redis.SMembers(ctx, cacheKey).Result()
	if err == redis.Nil || len(result) == 0 {
		return nil, nil
	}
	if err != nil {
		c.logger.Errorf("Get user roles cache failed: %v", err)
		return nil, err
	}
	
	return result, nil
}

// SetUserRolesToCache 将用户角色存入缓存
func (c *CacheManager) SetUserRolesToCache(ctx context.Context, userID string, roles []string) error {
	if len(roles) == 0 {
		return nil // 不缓存空角色
	}
	
	// 🔥 获取租户ID - 确保多租户隔离安全
	tenantID := tenantctx.GetTenantIDFromCtx(ctx)
	
	// 构建缓存key
	cacheKey := c.buildUserRoleCacheKey(tenantID, userID)
	
	// 使用Set数据结构存储角色列表
	pipe := c.redis.Pipeline()
	
	// 清除旧数据
	pipe.Del(ctx, cacheKey)
	
	// 添加新角色
	for _, role := range roles {
		pipe.SAdd(ctx, cacheKey, role)
	}
	
	// 设置过期时间
	pipe.Expire(ctx, cacheKey, c.userRoleTTL)
	
	// 执行管道
	_, err := pipe.Exec(ctx)
	if err != nil {
		c.logger.Errorf("Set user roles cache failed: %v", err)
		return err
	}
	
	return nil
}

// InvalidateUserCache 使用户相关缓存失效
func (c *CacheManager) InvalidateUserCache(ctx context.Context, userID string) error {
	// 🔥 获取租户ID - 确保多租户隔离安全
	tenantID := tenantctx.GetTenantIDFromCtx(ctx)
	
	// 删除用户角色缓存
	userRoleKey := c.buildUserRoleCacheKey(tenantID, userID)
	
	// 删除用户相关的权限检查缓存（通过模式匹配）
	permissionPattern := c.buildPermissionCachePattern(tenantID, userID)
	
	// 获取匹配的key
	keys, err := c.redis.Keys(ctx, permissionPattern).Result()
	if err != nil {
		c.logger.Errorf("Get permission cache keys failed: %v", err)
		return err
	}
	
	// 删除所有匹配的key
	if len(keys) > 0 {
		keys = append(keys, userRoleKey) // 添加用户角色key
		err = c.redis.Del(ctx, keys...).Err()
		if err != nil {
			c.logger.Errorf("Delete user cache failed: %v", err)
			return err
		}
		
		c.logger.Infof("Invalidated %d cache keys for user: %s", len(keys), userID)
	}
	
	return nil
}

// InvalidateTenantCache 使指定租户的所有缓存失效
func (c *CacheManager) InvalidateTenantCache(ctx context.Context) error {
	// 🔥 获取租户ID - 确保多租户隔离安全
	tenantID := tenantctx.GetTenantIDFromCtx(ctx)
	
	// 构建租户缓存模式
	pattern := fmt.Sprintf("casbin:tenant:%d:*", tenantID)
	
	// 获取匹配的key
	keys, err := c.redis.Keys(ctx, pattern).Result()
	if err != nil {
		c.logger.Errorf("Get tenant cache keys failed: %v", err)
		return err
	}
	
	// 删除所有匹配的key
	if len(keys) > 0 {
		err = c.redis.Del(ctx, keys...).Err()
		if err != nil {
			c.logger.Errorf("Delete tenant cache failed: %v", err)
			return err
		}
		
		c.logger.Infof("Invalidated %d cache keys for tenant: %d", len(keys), tenantID)
	}
	
	return nil
}

// GetCacheStats 获取缓存统计信息
func (c *CacheManager) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	// 🔥 获取租户ID - 确保多租户隔离安全
	tenantID := tenantctx.GetTenantIDFromCtx(ctx)
	
	stats := make(map[string]interface{})
	
	// 统计不同类型缓存的数量
	patterns := map[string]string{
		"permissions": fmt.Sprintf("casbin:tenant:%d:perm:*", tenantID),
		"user_roles":  fmt.Sprintf("casbin:tenant:%d:roles:*", tenantID),
	}
	
	for cacheType, pattern := range patterns {
		keys, err := c.redis.Keys(ctx, pattern).Result()
		if err != nil {
			c.logger.Errorf("Get cache stats for %s failed: %v", cacheType, err)
			continue
		}
		stats[cacheType] = len(keys)
	}
	
	stats["tenant_id"] = tenantID
	stats["generated_at"] = time.Now()
	
	return stats, nil
}

// 构建权限检查缓存key
func (c *CacheManager) buildPermissionCacheKey(tenantID uint64, subject, object, action, serviceName string) string {
	return fmt.Sprintf("casbin:tenant:%d:perm:%s:%s:%s:%s", tenantID, serviceName, subject, object, action)
}

// 构建权限检查缓存模式（用于批量删除）
func (c *CacheManager) buildPermissionCachePattern(tenantID uint64, subject string) string {
	return fmt.Sprintf("casbin:tenant:%d:perm:*:%s:*", tenantID, subject)
}

// 构建用户角色缓存key
func (c *CacheManager) buildUserRoleCacheKey(tenantID uint64, userID string) string {
	return fmt.Sprintf("casbin:tenant:%d:roles:%s", tenantID, userID)
}