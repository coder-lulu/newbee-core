package casbin

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/redis/go-redis/v9"

	"github.com/coder-lulu/newbee-common/v2/middleware/keys"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/entctx/tenantctx"
	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/zeromicro/go-zero/core/logx"
)

// EnforcerManager Casbin执行器管理器
type EnforcerManager struct {
	db           *ent.Client
	redis        redis.UniversalClient
	enforcers    sync.Map // map[uint64]*casbin.SyncedEnforcer - 按租户ID缓存执行器
	modelText    string   // Casbin模型定义
	mu           sync.RWMutex
	logger       logx.Logger
	cacheManager *CacheManager // 缓存管理器
}

const superAdminRoleCode = "superadmin"

// NewEnforcerManager 创建新的执行器管理器
func NewEnforcerManager(db *ent.Client, redisClient redis.UniversalClient, logger logx.Logger) *EnforcerManager {
	return &EnforcerManager{
		db:           db,
		redis:        redisClient,
		enforcers:    sync.Map{},
		modelText:    getDefaultModel(),
		logger:       logger,
		cacheManager: NewCacheManager(redisClient, logger),
	}
}

// GetEnforcer 获取指定租户的执行器
func (em *EnforcerManager) GetEnforcer(ctx context.Context) (*casbin.SyncedEnforcer, error) {
	// 🔥 获取租户ID - 确保多租户隔离安全
	tenantID := tenantctx.GetTenantIDFromCtx(ctx)

	// 检查是否已存在执行器
	if enforcer, ok := em.enforcers.Load(tenantID); ok {
		return enforcer.(*casbin.SyncedEnforcer), nil
	}

	em.mu.Lock()
	defer em.mu.Unlock()

	// 双重检查
	if enforcer, ok := em.enforcers.Load(tenantID); ok {
		return enforcer.(*casbin.SyncedEnforcer), nil
	}

	// 创建新的执行器
	enforcer, err := em.createEnforcer(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to create enforcer for tenant %d: %w", tenantID, err)
	}

	em.enforcers.Store(tenantID, enforcer)
	em.logger.Infof("Created new Casbin enforcer for tenant: %d", tenantID)

	return enforcer, nil
}

// createEnforcer 为指定租户创建新的执行器
func (em *EnforcerManager) createEnforcer(ctx context.Context, tenantID uint64) (*casbin.SyncedEnforcer, error) {
	// 创建模型
	m, err := model.NewModelFromString(em.modelText)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin model: %w", err)
	}

	// 创建适配器
	adapter := NewEntAdapter(em.db, ctx)

	// 创建执行器
	e, err := casbin.NewSyncedEnforcer(m, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin enforcer: %w", err)
	}

	// 加载策略
	err = e.LoadPolicy()
	if err != nil {
		return nil, fmt.Errorf("failed to load policies: %w", err)
	}

	// 启用日志
	e.EnableLog(true)

	return e, nil
}

// CheckPermission 检查权限（带缓存）
// 🔥 使用 RBAC with Domains 模型，租户ID作为domain参数
func (em *EnforcerManager) CheckPermission(ctx context.Context, subject, object, action, serviceName string) (*PermissionResult, error) {
	// 🔥 获取租户ID作为domain
	tenantID := tenantctx.GetTenantIDFromCtx(ctx)
	domain := strconv.FormatUint(tenantID, 10)

	// 首先尝试从缓存获取
	cachedResult, err := em.cacheManager.GetPermissionFromCache(ctx, subject, object, action, serviceName)
	if err != nil {
		em.logger.Errorf("Get permission from cache failed: %v", err)
		// 继续执行，不因缓存错误而中断
	}
	if cachedResult != nil {
		return cachedResult, nil
	}

	// 缓存未命中，执行实际检查
	enforcer, err := em.GetEnforcer(ctx)
	if err != nil {
		return nil, err
	}

	// 🔥 传入domain参数，确保租户隔离
	allowed, err := enforcer.Enforce(subject, domain, object, action)
	if err != nil {
		return nil, err
	}

	// 构建结果
	result := &PermissionResult{
		Allowed:      allowed,
		Reason:       fmt.Sprintf("checked by casbin enforcer for domain %s", domain),
		AppliedRules: []string{"direct"},
		FromCache:    false,
	}

	// 存入缓存
	cacheErr := em.cacheManager.SetPermissionToCache(ctx, subject, object, action, serviceName, result)
	if cacheErr != nil {
		em.logger.Errorf("Set permission to cache failed: %v", cacheErr)
		// 不影响主要逻辑
	}

	return result, nil
}

// CheckPermissionWithRoles 检查权限（包含角色，带缓存）
// 🔥 使用 RBAC with Domains 模型，租户ID作为domain参数
func (em *EnforcerManager) CheckPermissionWithRoles(ctx context.Context, subject, object, action, serviceName string) (*PermissionResult, error) {
	// 🔥 获取租户ID作为domain
	tenantID := tenantctx.GetTenantIDFromCtx(ctx)
	domain := strconv.FormatUint(tenantID, 10)

	// 超级管理员直接放行，确保 impersonation 不被数据权限拦截
	cm := keys.NewContextManager()
	roleCodes := cm.GetRoleCodes(ctx)
	isSuperAdmin := false
	for _, code := range strings.Split(roleCodes, ",") {
		if strings.TrimSpace(code) == superAdminRoleCode {
			isSuperAdmin = true
			break
		}
	}

	if isSuperAdmin {
		originalTenant := cm.GetOriginalTenantID(ctx)
		em.logger.Infof("Superadmin bypass applied: subject=%s, activeTenant=%s, originalTenant=%s, domain=%s, object=%s, action=%s",
			subject, cm.GetTenantID(ctx), originalTenant, domain, object, action)
		return &PermissionResult{
			Allowed:      true,
			Reason:       fmt.Sprintf("superadmin bypass for domain %s", domain),
			AppliedRules: []string{"superadmin_bypass"},
			FromCache:    false,
		}, nil
	}

	// 首先尝试从缓存获取
	cachedResult, err := em.cacheManager.GetPermissionFromCache(ctx, subject, object, action, serviceName)
	if err != nil {
		em.logger.Errorf("Get permission from cache failed: %v", err)
	}
	if cachedResult != nil {
		return cachedResult, nil
	}

	// 缓存未命中，执行实际检查
	enforcer, err := em.GetEnforcer(ctx)
	if err != nil {
		return nil, err
	}

	// 首先检查直接权限 - 🔥 传入domain
	if allowed, _ := enforcer.Enforce(subject, domain, object, action); allowed {
		result := &PermissionResult{
			Allowed:      true,
			Reason:       fmt.Sprintf("access granted via direct permission in domain %s", domain),
			AppliedRules: []string{"direct"},
			FromCache:    false,
		}

		// 存入缓存
		em.cacheManager.SetPermissionToCache(ctx, subject, object, action, serviceName, result)
		return result, nil
	}

	// 获取用户角色（优先从缓存）- 🔥 角色也在domain内
	roles, err := em.GetUserRolesWithCache(ctx, subject)
	if err != nil {
		return nil, err
	}

	// 检查每个角色的权限 - 🔥 传入domain
	var appliedRoles []string
	for _, role := range roles {
		if allowed, _ := enforcer.Enforce(role, domain, object, action); allowed {
			appliedRoles = append(appliedRoles, role)
		}
	}

	// 构建结果
	allowed := len(appliedRoles) > 0
	var reason string
	if allowed {
		reason = fmt.Sprintf("access granted via roles: %v in domain %s", appliedRoles, domain)
	} else {
		reason = fmt.Sprintf("access denied: no matching permissions or roles in domain %s", domain)
	}

	result := &PermissionResult{
		Allowed:      allowed,
		Reason:       reason,
		AppliedRules: appliedRoles,
		FromCache:    false,
	}

	// 存入缓存
	em.cacheManager.SetPermissionToCache(ctx, subject, object, action, serviceName, result)

	return result, nil
}

// GetUserRoles 获取用户角色
// 🔥 使用 RBAC with Domains 模型，获取指定租户域内的角色
func (em *EnforcerManager) GetUserRoles(ctx context.Context, user string) ([]string, error) {
	// 🔥 获取租户ID作为domain
	tenantID := tenantctx.GetTenantIDFromCtx(ctx)
	domain := strconv.FormatUint(tenantID, 10)

	enforcer, err := em.GetEnforcer(ctx)
	if err != nil {
		return nil, err
	}

	// 🔥 传入domain，获取该租户域内的角色
	return enforcer.GetRolesForUser(user, domain)
}

// GetUserRolesWithCache 获取用户角色（带缓存）
func (em *EnforcerManager) GetUserRolesWithCache(ctx context.Context, user string) ([]string, error) {
	// 先从缓存获取
	cachedRoles, err := em.cacheManager.GetUserRolesFromCache(ctx, user)
	if err != nil {
		em.logger.Errorf("Get user roles from cache failed: %v", err)
	}
	if cachedRoles != nil {
		return cachedRoles, nil
	}

	// 缓存未命中，从执行器获取
	roles, err := em.GetUserRoles(ctx, user)
	if err != nil {
		return nil, err
	}

	// 存入缓存
	cacheErr := em.cacheManager.SetUserRolesToCache(ctx, user, roles)
	if cacheErr != nil {
		em.logger.Errorf("Set user roles to cache failed: %v", cacheErr)
	}

	return roles, nil
}

// GetRoleUsers 获取角色用户
// 🔥 使用 RBAC with Domains 模型，获取指定租户域内的用户
func (em *EnforcerManager) GetRoleUsers(ctx context.Context, role string) ([]string, error) {
	// 🔥 获取租户ID作为domain
	tenantID := tenantctx.GetTenantIDFromCtx(ctx)
	domain := strconv.FormatUint(tenantID, 10)

	enforcer, err := em.GetEnforcer(ctx)
	if err != nil {
		return nil, err
	}

	// 🔥 传入domain，获取该租户域内的用户
	return enforcer.GetUsersForRole(role, domain)
}

// ReloadPolicy 重新加载指定租户的策略
func (em *EnforcerManager) ReloadPolicy(ctx context.Context) error {
	tenantID := tenantctx.GetTenantIDFromCtx(ctx)

	// 清除缓存
	err := em.cacheManager.InvalidateTenantCache(ctx)
	if err != nil {
		em.logger.Errorf("Failed to invalidate tenant cache: %v", err)
	}

	if enforcer, ok := em.enforcers.Load(tenantID); ok {
		err := enforcer.(*casbin.SyncedEnforcer).LoadPolicy()
		if err != nil {
			return fmt.Errorf("failed to reload policy for tenant %d: %w", tenantID, err)
		}
		em.logger.Infof("Reloaded policy for tenant: %d", tenantID)
	}

	return nil
}

// ClearCache 清除指定租户的缓存
func (em *EnforcerManager) ClearCache(ctx context.Context) {
	tenantID := tenantctx.GetTenantIDFromCtx(ctx)

	em.enforcers.Delete(tenantID)
	em.logger.Infof("Cleared enforcer cache for tenant: %d", tenantID)
}

// ClearAllCache 清除所有租户的缓存
func (em *EnforcerManager) ClearAllCache() {
	em.enforcers.Range(func(key, value interface{}) bool {
		em.enforcers.Delete(key)
		return true
	})
	em.logger.Info("Cleared all enforcer caches")
}

// getDefaultModel 获取默认的Casbin模型定义
// 🔥 使用 RBAC with Domains 模型，确保租户ID在规则层面显式隔离
func getDefaultModel() string {
	return `
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act, eft

[role_definition]
g = _, _, _

[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))

[matchers]
m = (g(r.sub, r.dom, p.sub) || (r.sub == p.sub)) && r.dom == p.dom && keyMatch2(r.obj, p.obj) && keyMatch2(r.act, p.act)
`
}

// GetStats 获取执行器统计信息
func (em *EnforcerManager) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	var tenantCount int
	em.enforcers.Range(func(key, value interface{}) bool {
		tenantCount++
		return true
	})

	stats["active_tenants"] = tenantCount
	stats["created_at"] = time.Now()

	return stats
}

// =================================
// 策略管理方法 - 支持 CRUD 同步
// =================================

// AddPolicy 添加权限策略
// 🔥 使用 RBAC with Domains 模型，将租户ID作为domain
func (em *EnforcerManager) AddPolicy(ctx context.Context, tenantID uint64, sub, obj, act string) (bool, error) {
	// 🔥 构建domain
	domain := strconv.FormatUint(tenantID, 10)

	enforcer, err := em.getEnforcerForTenant(ctx, tenantID)
	if err != nil {
		return false, err
	}

	// 🔥 传入domain，策略格式：sub, domain, obj, act
	added, err := enforcer.AddPolicy(sub, domain, obj, act)
	if err != nil {
		return false, err
	}

	// 清理相关缓存
	em.invalidatePermissionCache(ctx, sub)

	em.logger.Infof("Added policy: tenant=%d, domain=%s, sub=%s, obj=%s, act=%s, added=%t",
		tenantID, domain, sub, obj, act, added)
	return added, nil
}

// RemovePolicy 移除权限策略
// 🔥 使用 RBAC with Domains 模型，将租户ID作为domain
func (em *EnforcerManager) RemovePolicy(ctx context.Context, tenantID uint64, sub, obj, act string) (bool, error) {
	// 🔥 构建domain
	domain := strconv.FormatUint(tenantID, 10)

	enforcer, err := em.getEnforcerForTenant(ctx, tenantID)
	if err != nil {
		return false, err
	}

	// 🔥 传入domain
	removed, err := enforcer.RemovePolicy(sub, domain, obj, act)
	if err != nil {
		return false, err
	}

	// 清理相关缓存
	em.invalidatePermissionCache(ctx, sub)

	em.logger.Infof("Removed policy: tenant=%d, domain=%s, sub=%s, obj=%s, act=%s, removed=%t",
		tenantID, domain, sub, obj, act, removed)
	return removed, nil
}

// AddGroupingPolicy 添加角色继承策略
// 🔥 使用 RBAC with Domains 模型，将租户ID作为domain
func (em *EnforcerManager) AddGroupingPolicy(ctx context.Context, tenantID uint64, user, role string) (bool, error) {
	// 🔥 构建domain
	domain := strconv.FormatUint(tenantID, 10)

	enforcer, err := em.getEnforcerForTenant(ctx, tenantID)
	if err != nil {
		return false, err
	}

	// 🔥 传入domain，角色继承格式：user, role, domain
	added, err := enforcer.AddGroupingPolicy(user, role, domain)
	if err != nil {
		return false, err
	}

	// 清理相关缓存
	em.invalidateUserRoleCache(ctx, user)
	em.invalidatePermissionCache(ctx, user)

	em.logger.Infof("Added grouping policy: tenant=%d, domain=%s, user=%s, role=%s, added=%t",
		tenantID, domain, user, role, added)
	return added, nil
}

// RemoveGroupingPolicy 移除角色继承策略
// 🔥 使用 RBAC with Domains 模型，将租户ID作为domain
func (em *EnforcerManager) RemoveGroupingPolicy(ctx context.Context, tenantID uint64, user, role string) (bool, error) {
	// 🔥 构建domain
	domain := strconv.FormatUint(tenantID, 10)

	enforcer, err := em.getEnforcerForTenant(ctx, tenantID)
	if err != nil {
		return false, err
	}

	// 🔥 传入domain
	removed, err := enforcer.RemoveGroupingPolicy(user, role, domain)
	if err != nil {
		return false, err
	}

	// 清理相关缓存
	em.invalidateUserRoleCache(ctx, user)
	em.invalidatePermissionCache(ctx, user)

	em.logger.Infof("Removed grouping policy: tenant=%d, domain=%s, user=%s, role=%s, removed=%t",
		tenantID, domain, user, role, removed)
	return removed, nil
}

// AddNamedGroupingPolicy 添加命名角色继承策略
// 🔥 使用 RBAC with Domains 模型，将租户ID作为domain
func (em *EnforcerManager) AddNamedGroupingPolicy(ctx context.Context, tenantID uint64, ptype, user, role string) (bool, error) {
	// 🔥 构建domain
	domain := strconv.FormatUint(tenantID, 10)

	enforcer, err := em.getEnforcerForTenant(ctx, tenantID)
	if err != nil {
		return false, err
	}

	// 🔥 传入domain
	added, err := enforcer.AddNamedGroupingPolicy(ptype, user, role, domain)
	if err != nil {
		return false, err
	}

	// 清理相关缓存
	em.invalidateUserRoleCache(ctx, user)
	em.invalidatePermissionCache(ctx, user)

	em.logger.Infof("Added named grouping policy: tenant=%d, domain=%s, ptype=%s, user=%s, role=%s, added=%t",
		tenantID, domain, ptype, user, role, added)
	return added, nil
}

// RemoveNamedGroupingPolicy 移除命名角色继承策略
// 🔥 使用 RBAC with Domains 模型，将租户ID作为domain
func (em *EnforcerManager) RemoveNamedGroupingPolicy(ctx context.Context, tenantID uint64, ptype, user, role string) (bool, error) {
	// 🔥 构建domain
	domain := strconv.FormatUint(tenantID, 10)

	enforcer, err := em.getEnforcerForTenant(ctx, tenantID)
	if err != nil {
		return false, err
	}

	// 🔥 传入domain
	removed, err := enforcer.RemoveNamedGroupingPolicy(ptype, user, role, domain)
	if err != nil {
		return false, err
	}

	// 清理相关缓存
	em.invalidateUserRoleCache(ctx, user)
	em.invalidatePermissionCache(ctx, user)

	em.logger.Infof("Removed named grouping policy: tenant=%d, domain=%s, ptype=%s, user=%s, role=%s, removed=%t",
		tenantID, domain, ptype, user, role, removed)
	return removed, nil
}

// getEnforcerForTenant 获取指定租户的执行器
func (em *EnforcerManager) getEnforcerForTenant(ctx context.Context, tenantID uint64) (*casbin.SyncedEnforcer, error) {
	// 执行器是按租户隔离的，所以需要在上下文中设置正确的租户ID
	newCtx := context.WithValue(ctx, "tenantId", tenantID)
	return em.GetEnforcer(newCtx)
}

// invalidatePermissionCache 清理权限缓存
func (em *EnforcerManager) invalidatePermissionCache(ctx context.Context, subject string) {
	// 这里应该清理所有与该主体相关的权限缓存
	// 简化实现，可以进一步优化
	if em.cacheManager != nil {
		em.logger.Infof("Invalidating permission cache for subject: %s", subject)
	}
}

// invalidateUserRoleCache 清理用户角色缓存
func (em *EnforcerManager) invalidateUserRoleCache(ctx context.Context, user string) {
	// 这里应该清理该用户的角色缓存
	if em.cacheManager != nil {
		em.logger.Infof("Invalidating user role cache for user: %s", user)
	}
}
