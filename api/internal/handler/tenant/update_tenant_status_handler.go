package tenant

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/tenant"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route post /tenant/status tenant UpdateTenantStatus
//
// Update tenant status | 更新租户状态
//
// Update tenant status | 更新租户状态
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: TenantStatusReq
//
// Responses:
//  200: BaseMsgResp

func UpdateTenantStatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.TenantStatusReq
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := tenant.NewUpdateTenantStatusLogic(r.Context(), svcCtx)
		resp, err := l.UpdateTenantStatus(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
