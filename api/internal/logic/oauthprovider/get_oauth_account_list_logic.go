package oauthprovider

import (
	"context"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOauthAccountListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOauthAccountListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOauthAccountListLogic {
	return &GetOauthAccountListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOauthAccountListLogic) GetOauthAccountList(req *types.OauthAccountListReq) (resp *types.OauthAccountListResp, err error) {
	// todo: add your logic here and delete this line

	return
}
