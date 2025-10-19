package oauthprovider

import (
	"context"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateOauthAccountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateOauthAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateOauthAccountLogic {
	return &UpdateOauthAccountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateOauthAccountLogic) UpdateOauthAccount(req *types.OauthAccountInfo) (resp *types.BaseMsgResp, err error) {
	// todo: add your logic here and delete this line

	return
}
