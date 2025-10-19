package casbin

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/casbin"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route post /casbin/system/sync casbin SyncCasbinRules
//
// === 系统管理 === // Sync Casbin rules | 同步权限规则
//
// === 系统管理 === // Sync Casbin rules | 同步权限规则
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: SyncCasbinRulesReq
//
// Responses:
//  200: SyncCasbinRulesResp

func SyncCasbinRulesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SyncCasbinRulesReq
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := casbin.NewSyncCasbinRulesLogic(r.Context(), svcCtx)
		resp, err := l.SyncCasbinRules(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
