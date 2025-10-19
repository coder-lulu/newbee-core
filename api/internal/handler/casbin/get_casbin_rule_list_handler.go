package casbin

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/casbin"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route post /casbin/rules/list casbin GetCasbinRuleList
//
// Get Casbin rule list | 获取权限规则列表
//
// Get Casbin rule list | 获取权限规则列表
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: CasbinRuleListReq
//
// Responses:
//  200: CasbinRuleListResp

func GetCasbinRuleListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CasbinRuleListReq
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := casbin.NewGetCasbinRuleListLogic(r.Context(), svcCtx)
		resp, err := l.GetCasbinRuleList(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
