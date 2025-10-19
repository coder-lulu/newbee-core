package auditlog

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/auditlog"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route post /audit-log/stats auditlog GetAuditLogStats
//
// Get audit log statistics | 获取审计日志统计
//
// Get audit log statistics | 获取审计日志统计
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: AuditLogStatsReq
//
// Responses:
//  200: AuditLogStatsResp

func GetAuditLogStatsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AuditLogStatsReq
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := auditlog.NewGetAuditLogStatsLogic(r.Context(), svcCtx)
		resp, err := l.GetAuditLogStats(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
