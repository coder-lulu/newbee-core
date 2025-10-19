package oauthprovider

import (
	"context"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOauthAccountByIdLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOauthAccountByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOauthAccountByIdLogic {
	return &GetOauthAccountByIdLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOauthAccountByIdLogic) GetOauthAccountById(req *types.IDReq) (resp *types.OauthAccountInfoResp, err error) {
	// todo: add your logic here and delete this line

	return
}
