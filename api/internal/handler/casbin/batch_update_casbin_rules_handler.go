package casbin

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/casbin"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route put /casbin/rules/batch/update casbin BatchUpdateCasbinRules
//
// Batch update Casbin rules | 批量更新权限规则
//
// Batch update Casbin rules | 批量更新权限规则
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: BatchUpdateCasbinRulesReq
//
// Responses:
//  200: BaseMsgResp

func BatchUpdateCasbinRulesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.BatchUpdateCasbinRulesReq
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := casbin.NewBatchUpdateCasbinRulesLogic(r.Context(), svcCtx)
		resp, err := l.BatchUpdateCasbinRules(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
