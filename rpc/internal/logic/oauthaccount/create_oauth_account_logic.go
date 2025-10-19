package oauthaccount

import (
	"context"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateOauthAccountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateOauthAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOauthAccountLogic {
	return &CreateOauthAccountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// OAuth Account Binding management
func (l *CreateOauthAccountLogic) CreateOauthAccount(in *core.OauthAccountInfo) (*core.BaseIDResp, error) {
	// todo: add your logic here and delete this line

	return &core.BaseIDResp{}, nil
}
