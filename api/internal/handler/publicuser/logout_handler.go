package publicuser

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/coder-lulu/newbee-core/api/internal/logic/publicuser"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
)

// swagger:route get /user/logout publicuser Logout
//
// Log out | 退出登陆 (无需认证)
//
// Log out | 退出登陆 (无需认证)
//
// Responses:
//  200: BaseMsgResp

func LogoutHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := publicuser.NewLogoutLogic(r.Context(), svcCtx)
		resp, err := l.Logout()
		if err != nil {
			err = svcCtx.Trans.TransError(r.Context(), err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
