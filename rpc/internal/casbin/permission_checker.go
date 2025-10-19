package casbin

import (
	"context"
	"fmt"
	"time"

	"github.com/coder-lulu/newbee-common/orm/ent/entctx/tenantctx"
	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

// PermissionChecker 权限检查器 - 负责高性能的权限验证和决策
type PermissionChecker struct {
	db              *ent.Client
	redis           redis.UniversalClient
	enforcerManager *EnforcerManager
	policyManager   *PolicyManager
	logger          logx.Logger
	cacheManager    *CacheManager
}

// PermissionCheckResult 权限检查结果
type PermissionCheckResult struct {
	Allowed         bool              `json:"allowed"`
	Reason          string            `json:"reason"`
	AppliedRules    []string          `json:"applied_rules"`
	DataFilters     map[string]string `json:"data_filters"`
	FieldMasks      []string          `json:"field_masks"`
	CheckDurationMs int64             `json:"check_duration_ms"`
	FromCache       bool              `json:"from_cache"`
}

// PermissionCheckContext 权限检查上下文
type PermissionCheckContext struct {
	ServiceName string            `json:"service_name"`
	Subject     string            `json:"subject"`
	Object      string            `json:"object"`
	Action      string            `json:"action"`
	Context     map[string]string `json:"context"`
	EnableCache bool              `json:"enable_cache"`
	AuditLog    bool              `json:"audit_log"`
}

// NewPermissionChecker 创建新的权限检查器
func NewPermissionChecker(db *ent.Client, redisClient redis.UniversalClient, enforcerManager *EnforcerManager, policyManager *PolicyManager, logger logx.Logger) *PermissionChecker {
	return &PermissionChecker{
		db:              db,
		redis:           redisClient,
		enforcerManager: enforcerManager,
		policyManager:   policyManager,
		logger:          logger,
		cacheManager:    NewCacheManager(redisClient, logger),
	}
}

// CheckPermission 检查权限 - 带缓存的高性能权限验证
func (pc *PermissionChecker) CheckPermission(ctx context.Context, checkCtx *PermissionCheckContext) (*PermissionCheckResult, error) {
	startTime := time.Now()
	tenantID := tenantctx.GetTenantIDFromCtx(ctx)
	
	result := &PermissionCheckResult{
		Allowed:      false,
		AppliedRules: []string{},
		DataFilters:  make(map[string]string),
		FieldMasks:   []string{},
		FromCache:    false,
	}

	// 1. 检查缓存（如果启用）
	if checkCtx.EnableCache {
		if cachedResult := pc.getFromPermissionCache(tenantID, checkCtx); cachedResult != nil {
			cachedResult.FromCache = true
			cachedResult.CheckDurationMs = time.Since(startTime).Milliseconds()
			return cachedResult, nil
		}
	}

	// 2. 使用Casbin引擎进行权限检查
	enforcer, err := pc.enforcerManager.GetEnforcer(ctx)
	if err != nil {
		result.Reason = fmt.Sprintf("Failed to get enforcer: %v", err)
		result.CheckDurationMs = time.Since(startTime).Milliseconds()
		return result, err
	}

	// 3. 执行基础权限检查
	allowed, err := enforcer.Enforce(checkCtx.Subject, checkCtx.Object, checkCtx.Action)
	if err != nil {
		result.Reason = fmt.Sprintf("Casbin enforce failed: %v", err)
		result.CheckDurationMs = time.Since(startTime).Milliseconds()
		return result, err
	}

	result.Allowed = allowed

	// 4. 如果权限允许，获取详细的权限信息
	if allowed {
		// 获取应用的规则
		appliedRules, err := pc.getAppliedRules(ctx, checkCtx)
		if err != nil {
			pc.logger.Errorf("Failed to get applied rules: %v", err)
		} else {
			result.AppliedRules = appliedRules
		}

		// 获取数据过滤条件
		dataFilters, err := pc.getDataFilters(ctx, checkCtx)
		if err != nil {
			pc.logger.Errorf("Failed to get data filters: %v", err)
		} else {
			result.DataFilters = dataFilters
		}

		// 获取字段掩码
		fieldMasks, err := pc.getFieldMasks(ctx, checkCtx)
		if err != nil {
			pc.logger.Errorf("Failed to get field masks: %v", err)
		} else {
			result.FieldMasks = fieldMasks
		}

		result.Reason = "Permission granted"
	} else {
		result.Reason = "Permission denied by policy"
	}

	// 5. 缓存结果（如果启用）
	if checkCtx.EnableCache {
		pc.setPermissionCache(tenantID, checkCtx, result)
	}

	// 6. 记录审计日志（如果启用）
	if checkCtx.AuditLog {
		pc.logPermissionCheck(ctx, checkCtx, result)
	}

	result.CheckDurationMs = time.Since(startTime).Milliseconds()
	return result, nil
}

// CheckPermissionWithRoles 带角色的权限检查 - 支持角色展开
func (pc *PermissionChecker) CheckPermissionWithRoles(ctx context.Context, subject, object, action, serviceName string) (*PermissionCheckResult, error) {
	// 1. 直接检查用户权限
	checkCtx := &PermissionCheckContext{
		ServiceName: serviceName,
		Subject:     subject,
		Object:      object,
		Action:      action,
		EnableCache: true,
		AuditLog:    false,
	}

	result, err := pc.CheckPermission(ctx, checkCtx)
	if err != nil {
		return nil, err
	}

	// 2. 如果直接权限不通过，检查角色权限
	if !result.Allowed {
		roles, err := pc.policyManager.GetRolesForUser(ctx, subject)
		if err != nil {
			pc.logger.Errorf("Failed to get roles for user %s: %v", subject, err)
		} else {
			// 检查每个角色的权限
			for _, role := range roles {
				roleCheckCtx := &PermissionCheckContext{
					ServiceName: serviceName,
					Subject:     role,
					Object:      object,
					Action:      action,
					EnableCache: true,
					AuditLog:    false,
				}
				
				roleResult, err := pc.CheckPermission(ctx, roleCheckCtx)
				if err != nil {
					pc.logger.Errorf("Failed to check permission for role %s: %v", role, err)
					continue
				}

				if roleResult.Allowed {
					// 用角色权限结果更新原结果
					result.Allowed = true
					result.Reason = fmt.Sprintf("Permission granted through role: %s", role)
					result.AppliedRules = append(result.AppliedRules, roleResult.AppliedRules...)
					
					// 合并数据过滤条件
					for k, v := range roleResult.DataFilters {
						result.DataFilters[k] = v
					}
					
					// 合并字段掩码
					result.FieldMasks = append(result.FieldMasks, roleResult.FieldMasks...)
					break
				}
			}
		}
	}

	return result, nil
}

// BatchCheckPermission 批量权限检查 - 高效的批量验证
func (pc *PermissionChecker) BatchCheckPermission(ctx context.Context, checks []*PermissionCheckContext) ([]*PermissionCheckResult, error) {
	results := make([]*PermissionCheckResult, len(checks))
	
	// 并行处理批量检查
	type checkResult struct {
		index  int
		result *PermissionCheckResult
		err    error
	}
	
	resultChan := make(chan checkResult, len(checks))
	
	// 启动goroutines进行并行检查
	for i, checkCtx := range checks {
		go func(index int, ctx context.Context, checkCtx *PermissionCheckContext) {
			result, err := pc.CheckPermission(ctx, checkCtx)
			resultChan <- checkResult{
				index:  index,
				result: result,
				err:    err,
			}
		}(i, ctx, checkCtx)
	}
	
	// 收集结果
	for i := 0; i < len(checks); i++ {
		checkRes := <-resultChan
		if checkRes.err != nil {
			pc.logger.Errorf("Batch check failed for index %d: %v", checkRes.index, checkRes.err)
			// 创建失败结果
			results[checkRes.index] = &PermissionCheckResult{
				Allowed: false,
				Reason:  fmt.Sprintf("Check failed: %v", checkRes.err),
			}
		} else {
			results[checkRes.index] = checkRes.result
		}
	}
	
	return results, nil
}

// getAppliedRules 获取应用的规则
func (pc *PermissionChecker) getAppliedRules(ctx context.Context, checkCtx *PermissionCheckContext) ([]string, error) {
	// 获取用户的所有权限策略
	policies, err := pc.policyManager.GetPoliciesForSubject(ctx, checkCtx.Subject)
	if err != nil {
		return nil, err
	}

	appliedRules := []string{}
	for _, policy := range policies {
		if len(policy) >= 3 && policy[1] == checkCtx.Object && policy[2] == checkCtx.Action {
			ruleDesc := fmt.Sprintf("Policy: %s -> %s:%s", policy[0], policy[1], policy[2])
			appliedRules = append(appliedRules, ruleDesc)
		}
	}

	return appliedRules, nil
}

// getDataFilters 获取数据过滤条件
func (pc *PermissionChecker) getDataFilters(ctx context.Context, checkCtx *PermissionCheckContext) (map[string]string, error) {
	filters := make(map[string]string)
	
	// 基于服务名称和动作类型生成数据过滤条件
	switch checkCtx.ServiceName {
	case "cmdb":
		// CMDB服务的数据权限过滤
		if checkCtx.Action == "read" || checkCtx.Action == "list" {
			filters["data_scope"] = "own_dept" // 默认部门权限
		}
	case "workflow":
		// 工作流服务的数据权限过滤  
		if checkCtx.Action == "approve" {
			filters["approval_level"] = "department"
		}
	}

	// 从上下文中提取额外的过滤条件
	if checkCtx.Context != nil {
		if tenantID, ok := checkCtx.Context["tenant_id"]; ok {
			filters["tenant_id"] = tenantID
		}
		if departmentID, ok := checkCtx.Context["department_id"]; ok {
			filters["department_id"] = departmentID
		}
	}

	return filters, nil
}

// getFieldMasks 获取字段掩码
func (pc *PermissionChecker) getFieldMasks(ctx context.Context, checkCtx *PermissionCheckContext) ([]string, error) {
	masks := []string{}
	
	// 基于权限敏感度确定字段掩码
	if checkCtx.Action == "read" {
		// 读取权限可能需要隐藏敏感字段
		masks = append(masks, "password", "private_key", "secret")
	}

	return masks, nil
}

// getFromPermissionCache 从缓存获取权限结果
func (pc *PermissionChecker) getFromPermissionCache(tenantID uint64, checkCtx *PermissionCheckContext) *PermissionCheckResult {
	// 这里可以使用现有的缓存方法获取结果，但由于结构不匹配，暂时返回nil
	// 未来可以扩展CacheManager来支持完整的PermissionCheckResult结构
	return nil
}

// setPermissionCache 设置权限结果缓存
func (pc *PermissionChecker) setPermissionCache(tenantID uint64, checkCtx *PermissionCheckContext, result *PermissionCheckResult) {
	// 构建缓存结果结构（匹配现有PermissionResult结构）
	permResult := &PermissionResult{
		Allowed:      result.Allowed,
		Reason:       result.Reason,
		AppliedRules: result.AppliedRules,
		// DataFilters和FieldMasks暂不包含在缓存结构中
	}
	
	// 使用现有的缓存方法，这里传入context.Background()作为占位符
	ctx := context.Background()
	err := pc.cacheManager.SetPermissionToCache(ctx, checkCtx.Subject, checkCtx.Object, checkCtx.Action, checkCtx.ServiceName, permResult)
	if err != nil {
		pc.logger.Errorf("Failed to cache permission result: %v", err)
	} else {
		pc.logger.Infof("Permission result cached for %s:%s:%s", checkCtx.Subject, checkCtx.Object, checkCtx.Action)
	}
}

// logPermissionCheck 记录权限检查审计日志
func (pc *PermissionChecker) logPermissionCheck(ctx context.Context, checkCtx *PermissionCheckContext, result *PermissionCheckResult) {
	tenantID := tenantctx.GetTenantIDFromCtx(ctx)
	
	pc.logger.Infof("Permission check audit: tenant=%d, service=%s, subject=%s, object=%s, action=%s, allowed=%t, duration=%dms",
		tenantID, checkCtx.ServiceName, checkCtx.Subject, checkCtx.Object, checkCtx.Action, 
		result.Allowed, result.CheckDurationMs)
}