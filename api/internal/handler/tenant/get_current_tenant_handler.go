package tenant

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/tenant"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
)

// swagger:route get /tenant/current tenant GetCurrentTenant
//
// Get current active tenant | 获取当前激活租户
//
// Get current active tenant | 获取当前激活租户
//
// Responses:
//  200: CurrentTenantResp

func GetCurrentTenantHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := tenant.NewGetCurrentTenantLogic(r.Context(), svcCtx)
		resp, err := l.GetCurrentTenant()
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
