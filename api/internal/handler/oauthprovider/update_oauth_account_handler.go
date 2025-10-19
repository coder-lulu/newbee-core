package oauthprovider

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/oauthprovider"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route post /oauth_account/update oauthprovider UpdateOauthAccount
//
// Update oauth account | 更新OAuth账户
//
// Update oauth account | 更新OAuth账户
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: OauthAccountInfo
//
// Responses:
//  200: BaseMsgResp

func UpdateOauthAccountHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OauthAccountInfo
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := oauthprovider.NewUpdateOauthAccountLogic(r.Context(), svcCtx)
		resp, err := l.UpdateOauthAccount(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
