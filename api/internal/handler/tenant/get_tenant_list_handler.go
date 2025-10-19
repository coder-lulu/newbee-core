package tenant

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/tenant"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route post /tenant/list tenant GetTenantList
//
// Get tenant list | 获取租户列表
//
// Get tenant list | 获取租户列表
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: TenantListReq
//
// Responses:
//  200: TenantListResp

func GetTenantListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.TenantListReq
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := tenant.NewGetTenantListLogic(r.Context(), svcCtx)
		resp, err := l.GetTenantList(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
