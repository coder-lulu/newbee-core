package casbin

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/casbin"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route put /casbin/rules casbin UpdateCasbinRule
//
// Update Casbin rule | 更新权限规则
//
// Update Casbin rule | 更新权限规则
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: CasbinRuleInfo
//
// Responses:
//  200: BaseMsgResp

func UpdateCasbinRuleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CasbinRuleInfo
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := casbin.NewUpdateCasbinRuleLogic(r.Context(), svcCtx)
		resp, err := l.UpdateCasbinRule(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
