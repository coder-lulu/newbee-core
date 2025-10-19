package oauthprovider

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/oauthprovider"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route post /oauth/providers oauthprovider GetUserOauthProviders
//
// Get available OAuth providers for users | 获取用户可用的OAuth提供商
//
// Get available OAuth providers for users | 获取用户可用的OAuth提供商
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: UserOauthProviderListReq
//
// Responses:
//  200: UserOauthProviderListResp

func GetUserOauthProvidersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserOauthProviderListReq
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := oauthprovider.NewGetUserOauthProvidersLogic(r.Context(), svcCtx)
		resp, err := l.GetUserOauthProviders(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
