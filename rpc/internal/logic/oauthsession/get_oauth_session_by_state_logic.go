package oauthsession

import (
	"context"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOauthSessionByStateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOauthSessionByStateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOauthSessionByStateLogic {
	return &GetOauthSessionByStateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetOauthSessionByStateLogic) GetOauthSessionByState(in *core.GetOauthSessionByStateReq) (*core.OauthSessionInfo, error) {
	// todo: add your logic here and delete this line

	return &core.OauthSessionInfo{}, nil
}
