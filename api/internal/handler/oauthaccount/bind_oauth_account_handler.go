package oauthaccount

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/oauthaccount"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route post /oauth/bind oauthaccount BindOauthAccount
//
// User OAuth Account Management APIs | 用户OAuth账户管理API // Bind OAuth account | 绑定OAuth账户
//
// User OAuth Account Management APIs | 用户OAuth账户管理API // Bind OAuth account | 绑定OAuth账户
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: BindOauthAccountReq
//
// Responses:
//  200: BaseMsgResp

func BindOauthAccountHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.BindOauthAccountReq
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := oauthaccount.NewBindOauthAccountLogic(r.Context(), svcCtx)
		resp, err := l.BindOauthAccount(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
