package configuration

import (
	"context"

	"github.com/coder-lulu/newbee-common/i18n"
	"github.com/coder-lulu/newbee-common/utils/dynamicconf"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateConfigurationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateConfigurationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateConfigurationLogic {
	return &CreateConfigurationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateConfigurationLogic) CreateConfiguration(in *core.ConfigurationInfo) (*core.BaseIDResp, error) {
	result, err := l.svcCtx.DB.Configuration.Create().
		SetNotNilSort(in.Sort).
		SetNotNilState(in.State).
		SetNotNilName(in.Name).
		SetNotNilKey(in.Key).
		SetNotNilValue(in.Value).
		SetNotNilCategory(in.Category).
		SetNotNilRemark(in.Remark).
		Save(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	if in.Category != nil && in.Key != nil && in.Value != nil {
		err := dynamicconf.SetDynamicConfigurationToRedis(l.svcCtx.Redis, *in.Category, *in.Key, *in.Value)
		if err != nil {
			return nil, err
		}
	}

	return &core.BaseIDResp{Id: result.ID, Msg: i18n.CreateSuccess}, nil
}
