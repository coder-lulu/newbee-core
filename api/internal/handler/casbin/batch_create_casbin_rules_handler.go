package casbin

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/casbin"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route post /casbin/rules/batch/create casbin BatchCreateCasbinRules
//
// === 批量操作 === // Batch create Casbin rules | 批量创建权限规则
//
// === 批量操作 === // Batch create Casbin rules | 批量创建权限规则
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: BatchCreateCasbinRulesReq
//
// Responses:
//  200: BaseMsgResp

func BatchCreateCasbinRulesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.BatchCreateCasbinRulesReq
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := casbin.NewBatchCreateCasbinRulesLogic(r.Context(), svcCtx)
		resp, err := l.BatchCreateCasbinRules(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
