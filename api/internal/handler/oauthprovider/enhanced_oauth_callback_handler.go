package oauthprovider

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/oauthprovider"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route post /oauth/callback oauthprovider EnhancedOauthCallback
//
// Enhanced OAuth callback with parameters | 增强的OAuth回调处理
//
// Enhanced OAuth callback with parameters | 增强的OAuth回调处理
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: OauthCallbackReq
//
// Responses:
//  200: CallbackResp

func EnhancedOauthCallbackHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OauthCallbackReq
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := oauthprovider.NewEnhancedOauthCallbackLogic(r.Context(), svcCtx)
		resp, err := l.EnhancedOauthCallback(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
