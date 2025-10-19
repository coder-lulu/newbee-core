package casbin

import (
	"context"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncCasbinRulesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSyncCasbinRulesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncCasbinRulesLogic {
	return &SyncCasbinRulesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SyncCasbinRulesLogic) SyncCasbinRules(req *types.SyncCasbinRulesReq) (resp *types.SyncCasbinRulesResp, err error) {
	// 将API类型转换为RPC类型
	rpcReq := &core.SyncCasbinRulesReq{
		ServiceName: req.ServiceName,
		ForceReload: req.ForceReload,
	}

	// 调用RPC服务
	rpcResp, err := l.svcCtx.CoreRpc.SyncCasbinRules(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("Failed to sync casbin rules via RPC: %v", err)
		return &types.SyncCasbinRulesResp{
			BaseDataInfo: types.BaseDataInfo{
				Code: 1,
				Msg:  "同步权限规则失败: " + err.Error(),
			},
		}, nil
	}

	// 将RPC类型转换为API类型
	apiData := types.SyncResult{
		SyncedCount:    rpcResp.SyncedCount,
		SyncedServices: rpcResp.SyncedServices,
		SyncDurationMs: rpcResp.SyncDurationMs,
	}

	// 构建成功响应
	return &types.SyncCasbinRulesResp{
		BaseDataInfo: types.BaseDataInfo{
			Code: 0,
			Msg:  "同步权限规则成功",
		},
		Data: apiData,
	}, nil
}
