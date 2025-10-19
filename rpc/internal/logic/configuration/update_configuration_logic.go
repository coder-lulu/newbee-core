package configuration

import (
	"context"

	"github.com/coder-lulu/newbee-common/v2/i18n"
	"github.com/coder-lulu/newbee-common/v2/utils/dynamicconf"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateConfigurationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateConfigurationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateConfigurationLogic {
	return &UpdateConfigurationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateConfigurationLogic) UpdateConfiguration(in *core.ConfigurationInfo) (*core.BaseResp, error) {
	err := l.svcCtx.DB.Configuration.UpdateOneID(*in.Id).
		SetNotNilSort(in.Sort).
		SetNotNilState(in.State).
		SetNotNilName(in.Name).
		SetNotNilKey(in.Key).
		SetNotNilValue(in.Value).
		SetNotNilCategory(in.Category).
		SetNotNilRemark(in.Remark).
		Exec(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	if in.Category != nil && in.Key != nil && in.Value != nil {
		err := dynamicconf.SetDynamicConfigurationToRedis(l.svcCtx.Redis, *in.Category, *in.Key, *in.Value)
		if err != nil {
			return nil, err
		}
	}

	return &core.BaseResp{Msg: i18n.UpdateSuccess}, nil
}
