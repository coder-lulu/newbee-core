package oauthaccount

import (
	"context"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/coreclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnbindOauthAccountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUnbindOauthAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnbindOauthAccountLogic {
	return &UnbindOauthAccountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UnbindOauthAccountLogic) UnbindOauthAccount(req *types.UnbindOauthAccountReq) (resp *types.BaseMsgResp, err error) {
	// Call RPC service to unbind OAuth account
	result, err := l.svcCtx.CoreRpc.UnbindOauthAccount(l.ctx, &coreclient.UnbindOauthAccountReq{
		UserId:     req.UserId,
		ProviderId: req.ProviderId,
	})

	if err != nil {
		return nil, err
	}

	return &types.BaseMsgResp{Msg: result.Msg}, nil
}
