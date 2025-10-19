package casbin

import (
	"context"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshCasbinCacheLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRefreshCasbinCacheLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshCasbinCacheLogic {
	return &RefreshCasbinCacheLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RefreshCasbinCacheLogic) RefreshCasbinCache(req *types.RefreshCasbinCacheReq) (resp *types.RefreshCasbinCacheResp, err error) {
	// 将API类型转换为RPC类型
	rpcReq := &core.RefreshCasbinCacheReq{
		ServiceName: req.ServiceName,
		CacheType:   req.CacheType,
	}

	// 调用RPC服务
	rpcResp, err := l.svcCtx.CoreRpc.RefreshCasbinCache(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("Failed to refresh casbin cache via RPC: %v", err)
		return &types.RefreshCasbinCacheResp{
			BaseDataInfo: types.BaseDataInfo{
				Code: 1,
				Msg:  "刷新权限缓存失败: " + err.Error(),
			},
			Data: types.RefreshCacheResult{
				Success:       false,
				Message:       "刷新失败",
				ClearedEntries: 0,
			},
		}, nil
	}

	// 转换RPC响应为API类型
	return &types.RefreshCasbinCacheResp{
		BaseDataInfo: types.BaseDataInfo{
			Code: 0,
			Msg:  "权限缓存刷新完成",
		},
		Data: types.RefreshCacheResult{
			Success:        rpcResp.Success,
			Message:        rpcResp.Message,
			ClearedEntries: rpcResp.ClearedEntries,
		},
	}, nil
}
