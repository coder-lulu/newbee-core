package casbin

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/casbin"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route post /casbin/permission/summary casbin GetUserPermissionSummary
//
// === 权限查询 === // Get user permission summary | 获取用户权限摘要
//
// === 权限查询 === // Get user permission summary | 获取用户权限摘要
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: UserPermissionSummaryReq
//
// Responses:
//  200: UserPermissionSummaryResp

func GetUserPermissionSummaryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserPermissionSummaryReq
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := casbin.NewGetUserPermissionSummaryLogic(r.Context(), svcCtx)
		resp, err := l.GetUserPermissionSummary(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
