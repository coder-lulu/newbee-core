package configuration

import (
	"context"

	"fmt"
	"github.com/coder-lulu/newbee-common/config"
	"github.com/coder-lulu/newbee-common/i18n"
	"github.com/coder-lulu/newbee-core/rpc/ent/configuration"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteConfigurationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteConfigurationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteConfigurationLogic {
	return &DeleteConfigurationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteConfigurationLogic) DeleteConfiguration(in *core.IDsReq) (*core.BaseResp, error) {
	// delete redis cache
	check, err := l.svcCtx.DB.Configuration.Query().Where(configuration.IDIn(in.Ids...)).All(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	for _, item := range check {
		err := l.svcCtx.Redis.Del(l.ctx, fmt.Sprintf("%s%s:%s", config.RedisDynamicConfigurationPrefix, item.Category, item.Key)).Err()
		if err != nil {
			logx.Errorw("failed to delete dynamic configuration", logx.Field("category", item.Category), logx.Field("key", item.Key))
			return nil, err
		}
	}

	_, err = l.svcCtx.DB.Configuration.Delete().Where(configuration.IDIn(in.Ids...)).Exec(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &core.BaseResp{Msg: i18n.DeleteSuccess}, nil
}
