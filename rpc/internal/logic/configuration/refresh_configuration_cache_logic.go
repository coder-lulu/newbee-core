package configuration

import (
	"context"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/redisfunc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/coder-lulu/newbee-common/v2/config"
	"github.com/coder-lulu/newbee-common/v2/i18n"
	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshConfigurationCacheLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefreshConfigurationCacheLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshConfigurationCacheLogic {
	return &RefreshConfigurationCacheLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RefreshConfigurationCacheLogic) RefreshConfigurationCache(in *core.Empty) (*core.BaseResp, error) {
	// 清理所有配置缓存
	err := redisfunc.RemoveAllKeyByPrefix(l.ctx, config.RedisDynamicConfigurationPrefix, l.svcCtx.Redis)
	if err != nil {
		logx.Errorw("failed to refresh configuration cache", logx.Field("error", err))
		return nil, err
	}

	logx.Infow("configuration cache refreshed successfully")
	return &core.BaseResp{Msg: i18n.UpdateSuccess}, nil
}
