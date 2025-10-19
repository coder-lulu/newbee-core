package tenant

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/tenant"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route post /tenant/code tenant GetTenantByCode
//
// Get Tenant by Code | 通过租户码获取租户
//
// Get Tenant by Code | 通过租户码获取租户
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: TenantCodeReq
//
// Responses:
//  200: TenantInfoResp

func GetTenantByCodeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.TenantCodeReq
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := tenant.NewGetTenantByCodeLogic(r.Context(), svcCtx)
		resp, err := l.GetTenantByCode(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
