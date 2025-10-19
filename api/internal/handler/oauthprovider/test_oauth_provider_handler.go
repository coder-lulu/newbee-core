package oauthprovider

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/oauthprovider"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route post /oauth_provider/test oauthprovider TestOauthProvider
//
// Test oauth provider connection | 测试第三方提供商连接
//
// Test oauth provider connection | 测试第三方提供商连接
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: OauthProviderTestReq
//
// Responses:
//  200: OauthProviderTestResp

func TestOauthProviderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OauthProviderTestReq
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := oauthprovider.NewTestOauthProviderLogic(r.Context(), svcCtx)
		resp, err := l.TestOauthProvider(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
