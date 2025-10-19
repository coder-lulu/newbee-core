package oauthaccount

import (
	"context"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/coreclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type BindOauthAccountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBindOauthAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindOauthAccountLogic {
	return &BindOauthAccountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BindOauthAccountLogic) BindOauthAccount(req *types.BindOauthAccountReq) (resp *types.BaseMsgResp, err error) {
	// Call RPC service to bind OAuth account
	result, err := l.svcCtx.CoreRpc.BindOauthAccount(l.ctx, &coreclient.BindOauthAccountReq{
		UserId:            req.UserId,
		ProviderType:      req.ProviderType,
		ProviderId:        req.ProviderId,
		AuthorizationCode: req.AuthorizationCode,
		State:             req.State,
	})

	if err != nil {
		return nil, err
	}

	return &types.BaseMsgResp{Msg: result.Msg}, nil
}
