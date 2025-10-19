package auditlog

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/auditlog"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route post /audit-log auditlog GetAuditLogById
//
// Get audit log by ID | 通过ID获取审计日志
//
// Get audit log by ID | 通过ID获取审计日志
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: AuditLogReq
//
// Responses:
//  200: AuditLogResp

func GetAuditLogByIdHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AuditLogReq
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := auditlog.NewGetAuditLogByIdLogic(r.Context(), svcCtx)
		resp, err := l.GetAuditLogById(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
