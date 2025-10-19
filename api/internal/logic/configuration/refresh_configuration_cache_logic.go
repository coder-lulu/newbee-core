package configuration

import (
	"context"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshConfigurationCacheLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRefreshConfigurationCacheLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshConfigurationCacheLogic {
	return &RefreshConfigurationCacheLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RefreshConfigurationCacheLogic) RefreshConfigurationCache() (resp *types.BaseMsgResp, err error) {
	result, err := l.svcCtx.CoreRpc.RefreshConfigurationCache(l.ctx, &core.Empty{})
	if err != nil {
		return nil, err
	}

	return &types.BaseMsgResp{Msg: result.Msg}, nil
}
