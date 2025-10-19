package casbin

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/casbin"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route post /casbin/permission/batch/check casbin BatchCheckPermission
//
// Batch check permission | 批量权限检查
//
// Batch check permission | 批量权限检查
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: BatchPermissionCheckReq
//
// Responses:
//  200: BatchPermissionCheckResp

func BatchCheckPermissionHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.BatchPermissionCheckReq
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := casbin.NewBatchCheckPermissionLogic(r.Context(), svcCtx)
		resp, err := l.BatchCheckPermission(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
