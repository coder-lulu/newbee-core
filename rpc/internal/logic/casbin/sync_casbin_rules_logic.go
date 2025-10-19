package casbin

import (
	"context"
	"fmt"
	"time"

	"github.com/coder-lulu/newbee-common/orm/ent/entctx/tenantctx"
	"github.com/coder-lulu/newbee-core/rpc/ent/casbinrule"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncCasbinRulesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSyncCasbinRulesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncCasbinRulesLogic {
	return &SyncCasbinRulesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 系统管理
func (l *SyncCasbinRulesLogic) SyncCasbinRules(in *core.SyncCasbinRulesReq) (*core.SyncCasbinRulesResp, error) {
	startTime := time.Now()

	// 🔥 获取租户ID - 确保多租户隔离安全
	tenantID := tenantctx.GetTenantIDFromCtx(l.ctx)

	// 构建查询 - 必须包含租户ID过滤
	query := l.svcCtx.DB.CasbinRule.Query().Where(
		casbinrule.TenantIDEQ(tenantID), // 🔥 租户隔离
		casbinrule.StatusEQ(1),
	)

	// 按服务过滤
	var syncedServices []string
	if in.ServiceName != nil && *in.ServiceName != "" {
		query = query.Where(casbinrule.ServiceNameEQ(*in.ServiceName))
		syncedServices = []string{*in.ServiceName}
	} else {
		// 获取所有服务列表 - 必须包含租户ID过滤
		serviceResults, err := l.svcCtx.DB.CasbinRule.Query().
			Where(
				casbinrule.TenantIDEQ(tenantID), // 🔥 租户隔离
				casbinrule.StatusEQ(1),
			).
			GroupBy(casbinrule.FieldServiceName).
			Strings(l.ctx)
		if err != nil {
			l.Logger.Errorf("Get services failed: %v", err)
			return nil, err
		}
		syncedServices = serviceResults
	}

	// 统计需要同步的规则数量
	syncedCount, err := query.Count(l.ctx)
	if err != nil {
		l.Logger.Errorf("Count rules failed: %v", err)
		return nil, err
	}

	// 集成 Casbin 引擎进行规则同步
	if in.ForceReload != nil && *in.ForceReload {
		// 强制重新加载模式 - 清除所有租户缓存
		l.Logger.Infof("Force reload enabled, clearing all enforcer caches and reloading rules")
		l.svcCtx.EnforcerManager.ClearAllCache()
	} else {
		// 增量同步模式 - 只重新加载当前租户
		l.Logger.Infof("Incremental sync mode, reloading current tenant rules")
		err = l.svcCtx.EnforcerManager.ReloadPolicy(l.ctx)
		if err != nil {
			l.Logger.Errorf("Reload policy failed: %v", err)
			return nil, fmt.Errorf("reload policy failed: %v", err)
		}
	}

	duration := time.Since(startTime).Milliseconds()

	l.Logger.Infof("Casbin rules sync completed: services=%v, rules=%d, duration=%dms",
		syncedServices, syncedCount, duration)

	return &core.SyncCasbinRulesResp{
		SyncedCount:     int32(syncedCount),
		SyncedServices:  syncedServices,
		SyncDurationMs: duration,
	}, nil
}
