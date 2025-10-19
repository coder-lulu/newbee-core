package role

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/role"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route post /role/multiAuthUser role MultiAuthUser
//
// Auth User Role | 用户角色授权
//
// Auth User Role | 用户角色授权
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: RoleAuthReq
//
// Responses:
//  200: BaseMsgResp

func MultiAuthUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RoleAuthReq
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := role.NewMultiAuthUserLogic(r.Context(), svcCtx)
		resp, err := l.MultiAuthUser(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
