package oauthprovider

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/oauthprovider"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route post /oauth/statistics oauthprovider GetOauthStatistics
//
// Get OAuth statistics | 获取OAuth统计数据
//
// Get OAuth statistics | 获取OAuth统计数据
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: OauthStatisticsReq
//
// Responses:
//  200: OauthStatisticsResp

func GetOauthStatisticsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OauthStatisticsReq
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := oauthprovider.NewGetOauthStatisticsLogic(r.Context(), svcCtx)
		resp, err := l.GetOauthStatistics(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
