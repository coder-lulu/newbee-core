package oauthaccount

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/oauthaccount"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route post /oauth/accounts oauthaccount GetUserOauthAccounts
//
// Get user OAuth accounts | 获取用户OAuth账户
//
// Get user OAuth accounts | 获取用户OAuth账户
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: GetUserOauthAccountsReq
//
// Responses:
//  200: GetUserOauthAccountsResp

func GetUserOauthAccountsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetUserOauthAccountsReq
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := oauthaccount.NewGetUserOauthAccountsLogic(r.Context(), svcCtx)
		resp, err := l.GetUserOauthAccounts(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
