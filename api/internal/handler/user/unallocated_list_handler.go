package user

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/user"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route post /user/unallocatedList user UnallocatedList
//
// UnallocatedUserList | 获取未授权给当前角色的用户列表
//
// UnallocatedUserList | 获取未授权给当前角色的用户列表
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: RoleUnallocatedUserListReq
//
// Responses:
//  200: UserListResp

func UnallocatedListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RoleUnallocatedUserListReq
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := user.NewUnallocatedListLogic(r.Context(), svcCtx)
		resp, err := l.UnallocatedList(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
