package oauthprovider

import (
	"context"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type EnhancedOauthCallbackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewEnhancedOauthCallbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EnhancedOauthCallbackLogic {
	return &EnhancedOauthCallbackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *EnhancedOauthCallbackLogic) EnhancedOauthCallback(req *types.OauthCallbackReq) (resp *types.CallbackResp, err error) {
	// todo: add your logic here and delete this line

	return
}
