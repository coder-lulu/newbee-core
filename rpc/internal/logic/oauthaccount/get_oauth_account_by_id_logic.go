package oauthaccount

import (
	"context"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOauthAccountByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOauthAccountByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOauthAccountByIdLogic {
	return &GetOauthAccountByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetOauthAccountByIdLogic) GetOauthAccountById(in *core.IDReq) (*core.OauthAccountInfo, error) {
	// todo: add your logic here and delete this line

	return &core.OauthAccountInfo{}, nil
}
