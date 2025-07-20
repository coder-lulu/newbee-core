package role

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/role"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route post /role/changeRoleStatus role ChangeRoleStatus
//
// Change role Status | 更新角色状态
//
// Change role Status | 更新角色状态
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: RoleChangeStatusReq
//
// Responses:
//  200: BaseMsgResp

func ChangeRoleStatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RoleChangeStatusReq
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := role.NewChangeRoleStatusLogic(r.Context(), svcCtx)
		resp, err := l.ChangeRoleStatus(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
