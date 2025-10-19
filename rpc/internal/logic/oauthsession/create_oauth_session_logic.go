package oauthsession

import (
	"context"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateOauthSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateOauthSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOauthSessionLogic {
	return &CreateOauthSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// OAuth Session management
func (l *CreateOauthSessionLogic) CreateOauthSession(in *core.CreateOauthSessionReq) (*core.BaseIDResp, error) {
	// todo: add your logic here and delete this line

	return &core.BaseIDResp{}, nil
}
