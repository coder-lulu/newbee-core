package auth

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/auth"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
)

// swagger:route get /auth/tenant/list auth GetPublicTenantList
//
// Get public tenant list | 获取公开租户列表（无需认证）
//
// Get public tenant list | 获取公开租户列表（无需认证）
//
// Responses:
//  200: PublicTenantListResp

func GetPublicTenantListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := auth.NewGetPublicTenantListLogic(r.Context(), svcCtx)
		resp, err := l.GetPublicTenantList()
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
