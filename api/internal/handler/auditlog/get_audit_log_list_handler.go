package auditlog

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/auditlog"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route post /audit-log/list auditlog GetAuditLogList
//
// Get audit log list | 获取审计日志列表
//
// Get audit log list | 获取审计日志列表
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: AuditLogListReq
//
// Responses:
//  200: AuditLogListResp

func GetAuditLogListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AuditLogListReq
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := auditlog.NewGetAuditLogListLogic(r.Context(), svcCtx)
		resp, err := l.GetAuditLogList(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
