package casbin

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/casbin"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route post /casbin/permission/check casbin CheckPermission
//
// === 权限验证 === // Check permission | 权限检查
//
// === 权限验证 === // Check permission | 权限检查
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: PermissionCheckReq
//
// Responses:
//  200: PermissionCheckResp

func CheckPermissionHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PermissionCheckReq
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := casbin.NewCheckPermissionLogic(r.Context(), svcCtx)
		resp, err := l.CheckPermission(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
