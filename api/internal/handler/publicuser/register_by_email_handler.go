package publicuser

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/publicuser"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
)

// swagger:route post /user/register_by_email publicuser RegisterByEmail
//
// Register by Email | 邮箱注册
//
// Register by Email | 邮箱注册
//
// Parameters:
//  + name: body
//    require: true
//    in: body
//    type: RegisterByEmailReq
//
// Responses:
//  200: BaseMsgResp

func RegisterByEmailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RegisterByEmailReq
		if err := httpx.Parse(r, &req, true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := publicuser.NewRegisterByEmailLogic(r.Context(), svcCtx)
		resp, err := l.RegisterByEmail(&req)
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
