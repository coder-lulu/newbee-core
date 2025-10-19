package casbin

import (
	"context"
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/entctx/tenantctx"
	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/ent/casbinrule"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

// PolicyManager 策略管理器 - 负责Casbin策略的增删改查和同步
type PolicyManager struct {
	db              *ent.Client
	redis           redis.UniversalClient
	enforcerManager *EnforcerManager
	logger          logx.Logger
	cacheManager    *CacheManager
}

// NewPolicyManager 创建新的策略管理器
func NewPolicyManager(db *ent.Client, redisClient redis.UniversalClient, enforcerManager *EnforcerManager, logger logx.Logger) *PolicyManager {
	return &PolicyManager{
		db:              db,
		redis:           redisClient,
		enforcerManager: enforcerManager,
		logger:          logger,
		cacheManager:    NewCacheManager(redisClient, logger),
	}
}

// AddPolicy 添加权限策略
func (pm *PolicyManager) AddPolicy(ctx context.Context, tenantID uint64, subject, object, action string) (bool, error) {
	enforcer, err := pm.enforcerManager.GetEnforcer(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get enforcer: %w", err)
	}

	// 添加到Casbin引擎
	success, err := enforcer.AddPolicy(subject, object, action)
	if err != nil {
		return false, fmt.Errorf("failed to add policy to enforcer: %w", err)
	}

	if success {
		// 清除相关缓存
		pm.invalidatePolicyCache(ctx, tenantID, subject)
		
		pm.logger.Infof("Policy added successfully: subject=%s, object=%s, action=%s, tenant=%d", 
			subject, object, action, tenantID)
	}

	return success, nil
}

// RemovePolicy 移除权限策略
func (pm *PolicyManager) RemovePolicy(ctx context.Context, tenantID uint64, subject, object, action string) (bool, error) {
	enforcer, err := pm.enforcerManager.GetEnforcer(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get enforcer: %w", err)
	}

	// 从Casbin引擎移除
	success, err := enforcer.RemovePolicy(subject, object, action)
	if err != nil {
		return false, fmt.Errorf("failed to remove policy from enforcer: %w", err)
	}

	if success {
		// 清除相关缓存
		pm.invalidatePolicyCache(ctx, tenantID, subject)
		
		pm.logger.Infof("Policy removed successfully: subject=%s, object=%s, action=%s, tenant=%d", 
			subject, object, action, tenantID)
	}

	return success, nil
}

// AddGroupingPolicy 添加角色继承策略
func (pm *PolicyManager) AddGroupingPolicy(ctx context.Context, tenantID uint64, user, role string) (bool, error) {
	enforcer, err := pm.enforcerManager.GetEnforcer(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get enforcer: %w", err)
	}

	// 添加到Casbin引擎
	success, err := enforcer.AddGroupingPolicy(user, role)
	if err != nil {
		return false, fmt.Errorf("failed to add grouping policy to enforcer: %w", err)
	}

	if success {
		// 清除相关缓存
		pm.invalidateRoleCache(ctx, tenantID, user)
		
		pm.logger.Infof("Grouping policy added successfully: user=%s, role=%s, tenant=%d", 
			user, role, tenantID)
	}

	return success, nil
}

// RemoveGroupingPolicy 移除角色继承策略
func (pm *PolicyManager) RemoveGroupingPolicy(ctx context.Context, tenantID uint64, user, role string) (bool, error) {
	enforcer, err := pm.enforcerManager.GetEnforcer(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get enforcer: %w", err)
	}

	// 从Casbin引擎移除
	success, err := enforcer.RemoveGroupingPolicy(user, role)
	if err != nil {
		return false, fmt.Errorf("failed to remove grouping policy from enforcer: %w", err)
	}

	if success {
		// 清除相关缓存
		pm.invalidateRoleCache(ctx, tenantID, user)
		
		pm.logger.Infof("Grouping policy removed successfully: user=%s, role=%s, tenant=%d", 
			user, role, tenantID)
	}

	return success, nil
}

// GetPoliciesForSubject 获取主体的所有权限策略
func (pm *PolicyManager) GetPoliciesForSubject(ctx context.Context, subject string) ([][]string, error) {
	enforcer, err := pm.enforcerManager.GetEnforcer(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get enforcer: %w", err)
	}

	// 直接从Casbin引擎获取（暂时不使用缓存，因为CacheManager方法不匹配）
	policies, err := enforcer.GetPermissionsForUser(subject)
	if err != nil {
		return nil, err
	}
	
	return policies, nil
}

// GetRolesForUser 获取用户的所有角色
func (pm *PolicyManager) GetRolesForUser(ctx context.Context, user string) ([]string, error) {
	enforcer, err := pm.enforcerManager.GetEnforcer(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get enforcer: %w", err)
	}

	// 先尝试从缓存获取
	cachedRoles, err := pm.cacheManager.GetUserRolesFromCache(ctx, user)
	if err == nil && len(cachedRoles) > 0 {
		return cachedRoles, nil
	}

	// 从Casbin引擎获取
	roles, err := enforcer.GetRolesForUser(user)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles for user: %w", err)
	}
	
	// 缓存结果
	if len(roles) > 0 {
		err = pm.cacheManager.SetUserRolesToCache(ctx, user, roles)
		if err != nil {
			pm.logger.Errorf("Failed to cache user roles: %v", err)
		}
	}
	
	return roles, nil
}

// GetUsersForRole 获取角色的所有用户
func (pm *PolicyManager) GetUsersForRole(ctx context.Context, role string) ([]string, error) {
	enforcer, err := pm.enforcerManager.GetEnforcer(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get enforcer: %w", err)
	}

	// 直接从Casbin引擎获取（用户-角色反向查询较少使用，暂不缓存）
	users, err := enforcer.GetUsersForRole(role)
	if err != nil {
		return nil, fmt.Errorf("failed to get users for role: %w", err)
	}
	
	return users, nil
}

// SyncPoliciesFromDB 从数据库同步所有策略到Casbin引擎
func (pm *PolicyManager) SyncPoliciesFromDB(ctx context.Context, serviceName string) error {
	tenantID := tenantctx.GetTenantIDFromCtx(ctx)
	
	// 查询租户的所有权限规则
	rules, err := pm.db.CasbinRule.Query().
		Where(
			casbinrule.TenantIDEQ(tenantID),
			casbinrule.ServiceNameEQ(serviceName),
			casbinrule.StatusEQ(1), // 只同步启用的规则
		).
		All(ctx)
	if err != nil {
		return fmt.Errorf("failed to query casbin rules from database: %w", err)
	}

	enforcer, err := pm.enforcerManager.GetEnforcer(ctx)
	if err != nil {
		return fmt.Errorf("failed to get enforcer: %w", err)
	}

	// 清空当前策略
	enforcer.ClearPolicy()

	// 重新加载策略
	successCount := 0
	for _, rule := range rules {
		err := pm.addRuleToEnforcer(enforcer, rule)
		if err != nil {
			pm.logger.Errorf("Failed to add rule to enforcer: %v, rule ID: %d", err, rule.ID)
		} else {
			successCount++
		}
	}

	// 保存到Casbin持久化层
	err = enforcer.SavePolicy()
	if err != nil {
		return fmt.Errorf("failed to save policies: %w", err)
	}

	pm.logger.Infof("Synced %d/%d policies from database for service %s, tenant %d", 
		successCount, len(rules), serviceName, tenantID)

	return nil
}

// addRuleToEnforcer 将规则添加到Casbin执行器
func (pm *PolicyManager) addRuleToEnforcer(enforcer *casbin.SyncedEnforcer, rule *ent.CasbinRule) error {
	params := pm.buildCasbinParams(rule)
	
	switch rule.Ptype {
	case "p":
		if len(params) >= 3 {
			_, err := enforcer.AddPolicy(params[0], params[1], params[2])
			return err
		}
	case "g":
		if len(params) >= 2 {
			_, err := enforcer.AddGroupingPolicy(params[0], params[1])
			return err
		}
	case "g2", "g3", "g4", "g5":
		if len(params) >= 2 {
			_, err := enforcer.AddNamedGroupingPolicy(rule.Ptype, params[0], params[1])
			return err
		}
	}
	return nil
}

// buildCasbinParams 构建Casbin参数
func (pm *PolicyManager) buildCasbinParams(rule *ent.CasbinRule) []string {
	params := []string{}
	if rule.V0 != "" {
		params = append(params, rule.V0)
	}
	if rule.V1 != "" {
		params = append(params, rule.V1)
	}
	if rule.V2 != "" {
		params = append(params, rule.V2)
	}
	if rule.V3 != "" {
		params = append(params, rule.V3)
	}
	if rule.V4 != "" {
		params = append(params, rule.V4)
	}
	if rule.V5 != "" {
		params = append(params, rule.V5)
	}
	return params
}

// invalidatePolicyCache 使策略缓存失效
func (pm *PolicyManager) invalidatePolicyCache(ctx context.Context, tenantID uint64, subject string) {
	// 策略缓存失效 - 目前策略不使用专门缓存，此方法保留用于未来扩展
	pm.logger.Infof("Policy cache invalidated for subject: %s, tenant: %d", subject, tenantID)
}

// invalidateRoleCache 使角色缓存失效
func (pm *PolicyManager) invalidateRoleCache(ctx context.Context, tenantID uint64, user string) {
	// 使用现有的用户缓存失效方法
	err := pm.cacheManager.InvalidateUserCache(ctx, user)
	if err != nil {
		pm.logger.Errorf("Failed to invalidate user cache: %v", err)
	}
}