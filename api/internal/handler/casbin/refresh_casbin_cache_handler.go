package casbin

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/casbin"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route post /casbin/system/cache/refresh casbin RefreshCasbinCache
//
// Refresh Casbin cache | 刷新权限缓存
//
// Refresh Casbin cache | 刷新权限缓存
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: RefreshCasbinCacheReq
//
// Responses:
//  200: RefreshCasbinCacheResp

func RefreshCasbinCacheHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RefreshCasbinCacheReq
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := casbin.NewRefreshCasbinCacheLogic(r.Context(), svcCtx)
		resp, err := l.RefreshCasbinCache(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
