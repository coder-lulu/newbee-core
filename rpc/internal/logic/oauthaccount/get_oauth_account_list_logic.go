package oauthaccount

import (
	"context"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOauthAccountListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOauthAccountListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOauthAccountListLogic {
	return &GetOauthAccountListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetOauthAccountListLogic) GetOauthAccountList(in *core.OauthAccountListReq) (*core.OauthAccountListResp, error) {
	// todo: add your logic here and delete this line

	return &core.OauthAccountListResp{}, nil
}
