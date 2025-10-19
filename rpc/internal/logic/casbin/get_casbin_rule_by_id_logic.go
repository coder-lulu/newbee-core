package casbin

import (
	"context"
	"fmt"

	"github.com/coder-lulu/newbee-common/v2/orm/ent/entctx/tenantctx"
	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/ent/casbinrule"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCasbinRuleByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCasbinRuleByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCasbinRuleByIdLogic {
	return &GetCasbinRuleByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCasbinRuleByIdLogic) GetCasbinRuleById(in *core.IDReq) (*core.CasbinRuleInfo, error) {
	// 验证ID
	if in.Id == 0 {
		return nil, fmt.Errorf("id is required")
	}

	// 🔥 获取租户ID - 确保多租户隔离安全
	tenantID := tenantctx.GetTenantIDFromCtx(l.ctx)

	// 查询规则 - 必须包含租户ID过滤
	rule, err := l.svcCtx.DB.CasbinRule.Query().
		Where(
			casbinrule.TenantIDEQ(tenantID), // 🔥 租户隔离
			casbinrule.IDEQ(in.Id),
		).
		Only(l.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("casbin rule not found")
		}
		l.Logger.Errorf("Get casbin rule failed: %v", err)
		return nil, fmt.Errorf("get casbin rule failed: %v", err)
	}

	// 转换为protobuf模型，复用列表查询的转换函数
	listLogic := NewGetCasbinRuleListLogic(l.ctx, l.svcCtx)
	ruleInfo := listLogic.convertToRuleInfo(rule)

	l.Logger.Infof("Get casbin rule successfully, ID: %d, ptype: %s, service: %s",
		rule.ID, rule.Ptype, rule.ServiceName)

	return ruleInfo, nil
}
