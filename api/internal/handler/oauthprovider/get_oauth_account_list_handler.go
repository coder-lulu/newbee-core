package oauthprovider

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/oauthprovider"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route post /oauth_account/list oauthprovider GetOauthAccountList
//
// Get oauth account list | 获取OAuth账户列表
//
// Get oauth account list | 获取OAuth账户列表
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: OauthAccountListReq
//
// Responses:
//  200: OauthAccountListResp

func GetOauthAccountListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OauthAccountListReq
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := oauthprovider.NewGetOauthAccountListLogic(r.Context(), svcCtx)
		resp, err := l.GetOauthAccountList(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
