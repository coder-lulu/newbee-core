package plugins

import (
	"context"
	"fmt"
	"time"

	"github.com/coder-lulu/newbee-common/v2/orm/ent/hooks"
	"github.com/coder-lulu/newbee-common/v2/tenant"
	"github.com/coder-lulu/newbee-core/rpc/ent"
	tenant_ent "github.com/coder-lulu/newbee-core/rpc/ent/tenant"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/redisfunc"

	"github.com/zeromicro/go-zero/core/logx"
)

// CoreTenantPlugin 核心租户初始化插件
// 此插件封装了现有的租户初始化逻辑，确保100%兼容性
type CoreTenantPlugin struct {
	svcCtx *svc.ServiceContext
	logger logx.Logger
	config CorePluginConfig
}

// CorePluginConfig 核心插件配置
type CorePluginConfig struct {
	Timeout string `yaml:"timeout"`
}

// CorePluginFactory 核心插件工厂
type CorePluginFactory struct {
	svcCtx *svc.ServiceContext
	logger logx.Logger
}

// NewCorePluginFactory 创建核心插件工厂
func NewCorePluginFactory(svcCtx *svc.ServiceContext, logger logx.Logger) *CorePluginFactory {
	return &CorePluginFactory{
		svcCtx: svcCtx,
		logger: logger,
	}
}

// CreatePlugin 创建核心插件实例
func (f *CorePluginFactory) CreatePlugin(config map[string]any) (tenant.TenantInitPlugin, error) {
	var pluginConfig CorePluginConfig

	// 解析配置
	if timeoutVal, exists := config["timeout"]; exists {
		if timeoutStr, ok := timeoutVal.(string); ok {
			pluginConfig.Timeout = timeoutStr
		}
	}

	// 设置默认超时
	if pluginConfig.Timeout == "" {
		pluginConfig.Timeout = "120s"
	}

	return &CoreTenantPlugin{
		svcCtx: f.svcCtx,
		logger: f.logger,
		config: pluginConfig,
	}, nil
}

// GetPluginType 获取插件类型
func (f *CorePluginFactory) GetPluginType() tenant.PluginType {
	return tenant.PluginTypeCore
}

// ValidateConfig 验证插件配置
func (f *CorePluginFactory) ValidateConfig(config map[string]any) error {
	if timeoutVal, exists := config["timeout"]; exists {
		if timeoutStr, ok := timeoutVal.(string); !ok {
			return fmt.Errorf("timeout must be a string")
		} else {
			if _, err := time.ParseDuration(timeoutStr); err != nil {
				return fmt.Errorf("invalid timeout format: %w", err)
			}
		}
	}

	return nil
}

// GetMetadata 获取插件元数据
func (p *CoreTenantPlugin) GetMetadata() tenant.PluginMetadata {
	timeout, _ := time.ParseDuration(p.config.Timeout)

	return tenant.PluginMetadata{
		Name:              "core",
		Version:           "1.0.0",
		Dependencies:      []string{}, // 核心插件无依赖
		Priority:          1,          // 最高优先级
		Description:       "Core tenant initialization plugin that wraps existing logic",
		Type:              tenant.PluginTypeCore,
		EstimatedDuration: timeout,
		SupportRollback:   true,
		Concurrent:        false, // 核心初始化不支持并发
	}
}

// Initialize 执行租户初始化
func (p *CoreTenantPlugin) Initialize(ctx context.Context, req *tenant.InitRequest) error {
	p.logger.Infow("Core plugin starting tenant initialization",
		logx.Field("tenant_id", req.TenantID),
		logx.Field("request_id", req.RequestID),
		logx.Field("mode", req.Mode))

	// 直接调用现有初始化逻辑的核心方法
	// 避免循环导入，我们直接在这里实现核心逻辑
	return p.executeCoreTenantInit(ctx, req)
}

// OnInitializationComplete 在所有插件成功执行后触发，用于广播Casbin策略刷新
func (p *CoreTenantPlugin) OnInitializationComplete(ctx context.Context, req *tenant.InitRequest, result *tenant.ExecutionResult) error {
	if result == nil || result.Status != tenant.StatusSuccess {
		return nil
	}

	if p.svcCtx.Redis == nil {
		return fmt.Errorf("redis client is nil")
	}

	if err := redisfunc.PublishCasbinReload(ctx, p.svcCtx.Redis, p.svcCtx.Config.RedisConf.Db, req.TenantID, "core_plugin"); err != nil {
		return err
	}

	p.logger.Infow("✅ Broadcasted Casbin reload notification after tenant initialization",
		logx.Field("tenant_id", req.TenantID),
		logx.Field("request_id", result.RequestID))

	return nil
}

// executeCoreTenantInit 执行核心租户初始化逻辑
// 这里直接复制现有逻辑，避免循环导入
func (p *CoreTenantPlugin) executeCoreTenantInit(ctx context.Context, req *tenant.InitRequest) error {
	// 验证租户是否存在
	tenantInfo, err := p.svcCtx.DB.Tenant.Query().
		Where(tenant_ent.IDEQ(req.TenantID)).
		Only(hooks.NewSystemContext(ctx))
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("tenant %d not found", req.TenantID)
		}
		return fmt.Errorf("failed to query tenant: %w", err)
	}

	// 检查是否已经初始化过（除非是修复模式）
	if tenantInfo.Config != nil && req.Mode != tenant.InitModeRepair {
		if status, exists := tenantInfo.Config["status"]; exists {
			logx.Infof("当前租户状态: %s", status)
			if status == "completed" {
				p.logger.Infow("Tenant already initialized", logx.Field("tenant_id", req.TenantID))
				return nil // 不报错，视为成功
			}
		}
	}

	// 创建租户专属的上下文，避免SystemContext覆盖租户ID
	tenantCtx := hooks.SetTenantIDToContext(context.Background(), req.TenantID)

	// 在事务中执行初始化
	tx, err := p.svcCtx.DB.Tx(tenantCtx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	// 记录初始化开始状态
	initConfig := map[string]interface{}{
		"initialized_at": time.Now().Format(time.RFC3339),
		"version":        "2.0.0", // 标记为新版本
		"components":     []string{},
		"status":         "initializing",
	}

	_, err = tx.Tenant.UpdateOneID(req.TenantID).
		SetConfig(initConfig).
		Save(tenantCtx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update tenant status: %w", err)
	}

	// 执行初始化步骤（复制现有逻辑）

	// 第一阶段：初始化字典数据
	if err = p.initDictionaries(tenantCtx, tx, req.TenantID); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to init dictionaries: %w", err)
	}
	initConfig["components"] = append(initConfig["components"].([]string), "dictionaries")

	// 第二阶段：初始化系统配置
	if err = p.initConfigurations(tenantCtx, tx, req.TenantID); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to init configurations: %w", err)
	}
	initConfig["components"] = append(initConfig["components"].([]string), "configurations")

	// 第三阶段：初始化租户菜单
	if err = p.initTenantMenus(tenantCtx, tx, req.TenantID); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to init menus: %w", err)
	}
	initConfig["components"] = append(initConfig["components"].([]string), "tenant_menus")

	// 第四阶段：创建默认部门和职位
	dept, err := p.initDepartmentAndPositions(tenantCtx, tx, req.TenantID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to init departments and positions: %w", err)
	}
	initConfig["components"] = append(initConfig["components"].([]string), "departments", "positions")

	// 第五阶段：创建管理员角色和用户
	adminRole, adminUser, err := p.initAdminRoleAndUser(tenantCtx, tx, req, dept)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to init admin role and user: %w", err)
	}
	initConfig["components"] = append(initConfig["components"].([]string), "admin_role", "admin_user")

	// 🔥 Phase 3: 第六阶段：为管理员角色初始化数据权限规则
	if err = p.initAdminDataPermissions(tenantCtx, tx, adminRole, req.TenantID); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to init admin data permissions: %w", err)
	}
	initConfig["components"] = append(initConfig["components"].([]string), "data_permissions")

	// 第七阶段：更新初始化状态为部分完成
	// 注意：API权限初始化将在事务提交后进行
	initConfig["status"] = "partially_completed"
	initConfig["completed_at"] = time.Now().Format(time.RFC3339)
	initConfig["pending"] = []string{"api_permissions"}
	_, err = tx.Tenant.UpdateOneID(req.TenantID).
		SetConfig(initConfig).
		Save(tenantCtx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update tenant status: %w", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	p.logger.Infow("Transaction committed successfully, initializing API permissions",
		logx.Field("tenant_id", req.TenantID),
		logx.Field("admin_role_id", adminRole.ID))

	// ✅ 第八阶段：事务提交后初始化API权限
	// 原因：
	// 1. Casbin enforcer使用独立的数据库连接，不在事务tx中
	// 2. 必须等事务提交后，角色数据才对其他连接可见
	// 3. 如果API权限初始化失败，不影响租户创建（可后续补充）
	if err = p.initAdminAPIPermissions(tenantCtx, adminRole, req.TenantID); err != nil {
		p.logger.Errorw("Failed to init admin API permissions (tenant already created)",
			logx.Field("tenant_id", req.TenantID),
			logx.Field("admin_role_id", adminRole.ID),
			logx.Field("error", err.Error()))
		// 不返回错误，只记录日志
		// 租户创建成功，API权限可以后续手动添加
	} else {
		initConfig["components"] = append(initConfig["components"].([]string), "api_permissions")
		initConfig["status"] = "completed"
		delete(initConfig, "pending")

		// 更新租户状态为完全完成
		_, err = p.svcCtx.DB.Tenant.UpdateOneID(req.TenantID).
			SetConfig(initConfig).
			Save(tenantCtx)
		if err != nil {
			p.logger.Errorw("Failed to update final tenant status",
				logx.Field("tenant_id", req.TenantID),
				logx.Field("error", err.Error()))
			// 不返回错误，因为API权限已经初始化成功
		}
	}

	p.logger.Infow("Core tenant initialization completed successfully",
		logx.Field("tenant_id", req.TenantID),
		logx.Field("tenant_code", tenantInfo.Code),
		logx.Field("components", initConfig["components"]),
		logx.Field("admin_role_id", adminRole.ID),
		logx.Field("admin_user_id", adminUser.ID),
		logx.Field("department_id", dept.ID))

	return nil
}

// IsInitialized 检查租户是否已初始化
func (p *CoreTenantPlugin) IsInitialized(ctx context.Context, tenantID uint64) (bool, error) {
	// 检查租户配置中的初始化状态
	tenantInfo, err := p.svcCtx.DB.Tenant.Query().
		Where(tenant_ent.IDEQ(tenantID)).
		Only(hooks.NewSystemContext(ctx))
	if err != nil {
		if ent.IsNotFound(err) {
			return false, fmt.Errorf("tenant %d not found", tenantID)
		}
		return false, fmt.Errorf("failed to query tenant %d: %w", tenantID, err)
	}

	// 检查初始化状态
	if tenantInfo.Config != nil {
		if status, exists := tenantInfo.Config["status"]; exists {
			if status == "completed" {
				return true, nil
			}
		}
	}

	return false, nil
}

// HealthCheck 健康检查
func (p *CoreTenantPlugin) HealthCheck(ctx context.Context) error {
	// 检查数据库连接
	if p.svcCtx.DB == nil {
		return fmt.Errorf("database connection is nil")
	}

	// 执行简单的数据库查询
	_, err := p.svcCtx.DB.Tenant.Query().Count(hooks.NewSystemContext(ctx))
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// 以下是从现有逻辑复制的初始化方法，避免循环导入
// 这些方法与原始实现保持一致，确保100%兼容性
