package casbin

import (
	"context"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchDeleteCasbinRulesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchDeleteCasbinRulesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchDeleteCasbinRulesLogic {
	return &BatchDeleteCasbinRulesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BatchDeleteCasbinRulesLogic) BatchDeleteCasbinRules(in *core.IDsReq) (*core.BaseResp, error) {
	// 直接重用单个删除逻辑，它已经支持批量删除
	deleteLogic := NewDeleteCasbinRuleLogic(l.ctx, l.svcCtx)
	return deleteLogic.DeleteCasbinRule(in)
}
