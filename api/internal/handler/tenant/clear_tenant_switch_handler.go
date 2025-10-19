package tenant

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/tenant"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
)

// swagger:route get /tenant/dynamic/clear tenant ClearTenantSwitch
//
// Clear tenant switch | 清除租户切换
//
// Clear tenant switch | 清除租户切换
//
// Responses:
//  200: BaseMsgResp

func ClearTenantSwitchHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := tenant.NewClearTenantSwitchLogic(r.Context(), svcCtx)
		resp, err := l.ClearTenantSwitch()
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
