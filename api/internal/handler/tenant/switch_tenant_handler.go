package tenant

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/tenant"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route post /tenant/switch tenant SwitchTenant
//
// Switch tenant for super admin | 超级管理员切换租户
//
// Switch tenant for super admin | 超级管理员切换租户
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: TenantSwitchReq
//
// Responses:
//  200: BaseMsgResp

func SwitchTenantHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.TenantSwitchReq
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := tenant.NewSwitchTenantLogic(r.Context(), svcCtx)
		resp, err := l.SwitchTenant(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
