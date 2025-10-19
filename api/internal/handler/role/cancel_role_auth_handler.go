package role

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/role"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route post /role/cancelAuthUser role CancelRoleAuth
//
// Cancel User Role Auth | 取消用户角色授权
//
// Cancel User Role Auth | 取消用户角色授权
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: RoleAuthReq
//
// Responses:
//  200: BaseMsgResp

func CancelRoleAuthHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RoleAuthReq
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := role.NewCancelRoleAuthLogic(r.Context(), svcCtx)
		resp, err := l.CancelRoleAuth(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
