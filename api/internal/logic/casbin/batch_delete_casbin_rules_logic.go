package casbin

import (
	"context"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchDeleteCasbinRulesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBatchDeleteCasbinRulesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchDeleteCasbinRulesLogic {
	return &BatchDeleteCasbinRulesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BatchDeleteCasbinRulesLogic) BatchDeleteCasbinRules(req *types.IDsReq) (resp *types.BaseMsgResp, err error) {
	// 将API类型转换为RPC类型
	rpcReq := &core.IDsReq{
		Ids: req.Ids,
	}

	// 调用RPC服务
	_, err = l.svcCtx.CoreRpc.BatchDeleteCasbinRules(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("Failed to batch delete casbin rules via RPC: %v", err)
		return &types.BaseMsgResp{
			Code: 1,
			Msg:  "批量删除权限规则失败: " + err.Error(),
		}, nil
	}

	// 构建成功响应
	return &types.BaseMsgResp{
		Code: 0,
		Msg:  "批量删除权限规则成功",
	}, nil
}
