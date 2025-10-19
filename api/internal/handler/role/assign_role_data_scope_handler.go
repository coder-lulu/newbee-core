package role

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/role"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route post /role/dataScope role AssignRoleDataScope
//
// Assign Role DataScope | 授权数据权限
//
// Assign Role DataScope | 授权数据权限
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: RoleDataScopeReq
//
// Responses:
//  200: BaseMsgResp

func AssignRoleDataScopeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RoleDataScopeReq
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := role.NewAssignRoleDataScopeLogic(r.Context(), svcCtx)
		resp, err := l.AssignRoleDataScope(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
