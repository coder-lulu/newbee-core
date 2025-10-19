package tenant

import (
	"context"
	"fmt"
	"time"

	"github.com/coder-lulu/newbee-common/v2/tenant"
	"github.com/coder-lulu/newbee-core/rpc/internal/plugins"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/gofrs/uuid/v5"
	"github.com/zeromicro/go-zero/core/logx"
)

// InitTenantLogicV2 使用插件框架的新版本初始化逻辑
// 此版本与现有API完全兼容，但内部使用插件架构
type InitTenantLogicV2 struct {
	ctx           context.Context
	svcCtx        *svc.ServiceContext
	pluginManager *tenant.PluginManager
	stateManager  *tenant.StateManager
	logger        logx.Logger
	enabled       bool // 通过配置控制是否启用新框架
}

// NewInitTenantLogicV2 创建新版本的初始化逻辑
func NewInitTenantLogicV2(ctx context.Context, svcCtx *svc.ServiceContext) *InitTenantLogicV2 {
	logger := logx.WithContext(ctx)
	// 加载插件配置
	configLoader := tenant.NewConfigLoader("etc/tenant_plugins.yaml")
	config, err := configLoader.LoadConfig()
	if err != nil {
		logger.Errorw("Failed to load plugin config, falling back to legacy logic",
			logx.Field("error", err.Error()))
		return &InitTenantLogicV2{
			ctx:     ctx,
			svcCtx:  svcCtx,
			logger:  logger,
			enabled: false,
		}
	}

	// 创建插件管理器
	pluginManager := tenant.NewPluginManager(config, logger)
	stateManager := tenant.NewStateManager(svcCtx.DB, logger)

	// 注册核心插件工厂
	coreFactory := plugins.NewCorePluginFactory(svcCtx, logger)
	if err := pluginManager.RegisterFactory("core", coreFactory); err != nil {
		logger.Errorw("Failed to register core plugin factory",
			logx.Field("error", err.Error()))
		return &InitTenantLogicV2{
			ctx:     ctx,
			svcCtx:  svcCtx,
			logger:  logger,
			enabled: false,
		}
	}

	// 加载插件
	if err := pluginManager.LoadPlugins(); err != nil {
		logger.Errorw("Failed to load plugins, falling back to legacy logic",
			logx.Field("error", err.Error()))
		return &InitTenantLogicV2{
			ctx:     ctx,
			svcCtx:  svcCtx,
			logger:  logger,
			enabled: false,
		}
	}

	// 验证插件依赖
	if err := pluginManager.ValidatePlugins(); err != nil {
		logger.Errorw("Plugin validation failed, falling back to legacy logic",
			logx.Field("error", err.Error()))
		return &InitTenantLogicV2{
			ctx:     ctx,
			svcCtx:  svcCtx,
			logger:  logger,
			enabled: false,
		}
	}

	return &InitTenantLogicV2{
		ctx:           ctx,
		svcCtx:        svcCtx,
		pluginManager: pluginManager,
		stateManager:  stateManager,
		logger:        logger,
		enabled:       true,
	}
}

// InitTenant 执行租户初始化（保持API兼容）
func (l *InitTenantLogicV2) InitTenant(in *core.TenantInitReq) (*core.BaseResp, error) {
	// 如果新框架未启用，回退到旧逻辑
	if !l.enabled {
		legacyLogic := NewInitTenantLogicLegacy(l.ctx, l.svcCtx)
		return legacyLogic.InitTenant(in)
	}

	// 使用新框架执行初始化
	return l.initTenantWithPlugins(in)
}

// initTenantWithPlugins 使用插件框架执行初始化
func (l *InitTenantLogicV2) initTenantWithPlugins(in *core.TenantInitReq) (*core.BaseResp, error) {
	l.logger.Infow("Starting tenant initialization with plugin framework",
		logx.Field("tenant_id", in.TenantId))

	// 生成请求ID
	requestID := uuid.Must(uuid.NewV4()).String()

	// 转换请求格式
	initReq := &tenant.InitRequest{
		TenantID:      in.TenantId,
		AdminUsername: in.AdminUsername,
		AdminPassword: in.AdminPassword,
		AdminEmail:    in.AdminEmail,
		RequestID:     requestID,
		Mode:          tenant.InitModeFull,
		DryRun:        false,
		Timeout:       300 * time.Second,
	}

	// 获取启用的插件
	plugins := l.pluginManager.GetEnabledPlugins()
	if len(plugins) == 0 {
		return nil, fmt.Errorf("no enabled plugins found")
	}

	// 创建初始化状态
	if err := l.stateManager.CreateInitState(l.ctx, initReq, len(plugins)); err != nil {
		l.logger.Errorw("Failed to create initialization state",
			logx.Field("tenant_id", in.TenantId),
			logx.Field("error", err.Error()))
		return nil, fmt.Errorf("failed to create initialization state: %w", err)
	}

	// 更新状态为运行中
	if err := l.stateManager.UpdateInitStatus(l.ctx, in.TenantId, tenant.StatusRunning, ""); err != nil {
		l.logger.Errorw("Failed to update initialization status",
			logx.Field("tenant_id", in.TenantId),
			logx.Field("error", err.Error()))
	}

	// 执行插件初始化
	result, err := l.pluginManager.ExecutePlugins(l.ctx, initReq)
	if err != nil {
		l.logger.Errorw("Plugin execution failed",
			logx.Field("tenant_id", in.TenantId),
			logx.Field("error", err.Error()))

		// 更新失败状态
		l.stateManager.UpdateInitStatus(l.ctx, in.TenantId, tenant.StatusFailed, err.Error())

		// 如果配置了失败回滚，则执行回滚
		if l.shouldRollbackOnFailure() {
			if rollbackErr := l.rollbackTenant(l.ctx, in.TenantId, result.FailedPlugin); rollbackErr != nil {
				l.logger.Errorw("Rollback failed",
					logx.Field("tenant_id", in.TenantId),
					logx.Field("rollback_error", rollbackErr.Error()))
			}
		}

		return nil, fmt.Errorf("tenant initialization failed: %w", err)
	}

	// 更新成功状态
	if err := l.stateManager.UpdateInitStatus(l.ctx, in.TenantId, tenant.StatusSuccess, ""); err != nil {
		l.logger.Errorw("Failed to update success status",
			logx.Field("tenant_id", in.TenantId),
			logx.Field("error", err.Error()))
	}

	l.logger.Infow("Tenant initialization completed successfully",
		logx.Field("tenant_id", in.TenantId),
		logx.Field("request_id", requestID),
		logx.Field("duration", result.Duration),
		logx.Field("plugin_count", len(result.PluginResults)))

	return &core.BaseResp{
		Msg: fmt.Sprintf("租户初始化成功，耗时 %v", result.Duration),
	}, nil
}

// shouldRollbackOnFailure 检查是否应该在失败时回滚
func (l *InitTenantLogicV2) shouldRollbackOnFailure() bool {
	// 从配置中读取回滚策略
	// 这里简化实现，实际可以从配置文件读取
	return true
}

// rollbackTenant 回滚租户初始化
func (l *InitTenantLogicV2) rollbackTenant(ctx context.Context, tenantID uint64, failedPlugin string) error {
	l.logger.Infow("Starting tenant rollback",
		logx.Field("tenant_id", tenantID),
		logx.Field("failed_plugin", failedPlugin))

	// 获取当前状态，确定需要回滚的插件
	state, err := l.stateManager.GetState(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get rollback state: %w", err)
	}

	// 获取执行顺序（用于反向回滚）
	plugins, err := l.pluginManager.GetExecutionOrder()
	if err != nil {
		return fmt.Errorf("failed to get execution order for rollback: %w", err)
	}

	// 反向执行回滚
	rollbackCount := 0
	for i := len(plugins) - 1; i >= 0; i-- {
		plugin := plugins[i]
		pluginName := plugin.GetMetadata().Name

		// 跳过未执行或失败的插件
		pluginState, exists := state.Plugins[pluginName]
		if !exists || pluginState.Status != tenant.StatusSuccess {
			continue
		}

		l.logger.Infow("Rolling back plugin",
			logx.Field("plugin", pluginName),
			logx.Field("tenant_id", tenantID))

		// 执行插件回滚
		if err := plugin.Rollback(ctx, tenantID); err != nil {
			l.logger.Errorw("Plugin rollback failed",
				logx.Field("plugin", pluginName),
				logx.Field("tenant_id", tenantID),
				logx.Field("error", err.Error()))
			// 继续回滚其他插件，但记录错误
		} else {
			rollbackCount++
			l.logger.Infow("Plugin rollback completed",
				logx.Field("plugin", pluginName),
				logx.Field("tenant_id", tenantID))
		}
	}

	// 更新状态为已回滚
	if err := l.stateManager.UpdateInitStatus(ctx, tenantID, tenant.StatusRolledback,
		fmt.Sprintf("Rolled back %d plugins", rollbackCount)); err != nil {
		l.logger.Errorw("Failed to update rollback status",
			logx.Field("tenant_id", tenantID),
			logx.Field("error", err.Error()))
	}

	l.logger.Infow("Tenant rollback completed",
		logx.Field("tenant_id", tenantID),
		logx.Field("rollback_count", rollbackCount))

	return nil
}

// GetInitializationStatus 获取初始化状态（新增方法）
func (l *InitTenantLogicV2) GetInitializationStatus(tenantID uint64) (*tenant.InitializationState, error) {
	if !l.enabled {
		return nil, fmt.Errorf("plugin framework not enabled")
	}

	return l.stateManager.GetState(l.ctx, tenantID)
}

// RetryInitialization 重试失败的初始化（新增方法）
func (l *InitTenantLogicV2) RetryInitialization(tenantID uint64) error {
	if !l.enabled {
		return fmt.Errorf("plugin framework not enabled")
	}

	state, err := l.stateManager.GetState(l.ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get current state: %w", err)
	}

	if state.Status != tenant.StatusFailed {
		return fmt.Errorf("tenant %d is not in failed state, current status: %s", tenantID, state.Status)
	}

	// 重构请求并重试
	initReq := &tenant.InitRequest{
		TenantID:  tenantID,
		RequestID: uuid.Must(uuid.NewV4()).String(),
		Mode:      tenant.InitModeRepair, // 使用修复模式
		DryRun:    false,
		Timeout:   300 * time.Second,
	}

	// 重置状态
	plugins := l.pluginManager.GetEnabledPlugins()
	if err := l.stateManager.CreateInitState(l.ctx, initReq, len(plugins)); err != nil {
		return fmt.Errorf("failed to reset state: %w", err)
	}

	// 执行重试
	_, err = l.pluginManager.ExecutePlugins(l.ctx, initReq)
	if err != nil {
		l.stateManager.UpdateInitStatus(l.ctx, tenantID, tenant.StatusFailed, err.Error())
		return fmt.Errorf("retry failed: %w", err)
	}

	l.stateManager.UpdateInitStatus(l.ctx, tenantID, tenant.StatusSuccess, "Retry completed successfully")
	return nil
}

// ListPlugins 列出可用插件（新增方法）
func (l *InitTenantLogicV2) ListPlugins() map[string]tenant.PluginMetadata {
	if !l.enabled {
		return nil
	}

	return l.pluginManager.ListPlugins()
}
