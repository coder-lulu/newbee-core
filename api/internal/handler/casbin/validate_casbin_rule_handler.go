package casbin

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/casbin"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route post /casbin/rules/validate casbin ValidateCasbinRule
//
// === 规则验证 === // Validate Casbin rule | 验证权限规则
//
// === 规则验证 === // Validate Casbin rule | 验证权限规则
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: ValidateCasbinRuleReq
//
// Responses:
//  200: ValidateCasbinRuleResp

func ValidateCasbinRuleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ValidateCasbinRuleReq
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := casbin.NewValidateCasbinRuleLogic(r.Context(), svcCtx)
		resp, err := l.ValidateCasbinRule(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
