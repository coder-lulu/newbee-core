package oauthprovider

import (
	"context"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type TestOauthProviderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTestOauthProviderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TestOauthProviderLogic {
	return &TestOauthProviderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TestOauthProviderLogic) TestOauthProvider(req *types.OauthProviderTestReq) (resp *types.OauthProviderTestResp, err error) {
	// todo: add your logic here and delete this line

	return
}
