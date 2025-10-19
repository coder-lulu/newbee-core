package oauthaccount

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/oauthaccount"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route post /oauth/unbind oauthaccount UnbindOauthAccount
//
// Unbind OAuth account | 解绑OAuth账户
//
// Unbind OAuth account | 解绑OAuth账户
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: UnbindOauthAccountReq
//
// Responses:
//  200: BaseMsgResp

func UnbindOauthAccountHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UnbindOauthAccountReq
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := oauthaccount.NewUnbindOauthAccountLogic(r.Context(), svcCtx)
		resp, err := l.UnbindOauthAccount(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
