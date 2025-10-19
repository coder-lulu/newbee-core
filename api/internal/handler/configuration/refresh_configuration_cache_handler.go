package configuration

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/configuration"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
)

// swagger:route delete /configuration/refreshCache configuration RefreshConfigurationCache
//
// Refresh configuration cache | 刷新参数配置缓存
//
// Refresh configuration cache | 刷新参数配置缓存
//
// Responses:
//  200: BaseMsgResp

func RefreshConfigurationCacheHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := configuration.NewRefreshConfigurationCacheLogic(r.Context(), svcCtx)
		resp, err := l.RefreshConfigurationCache()
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
