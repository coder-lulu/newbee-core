package authority

import (
	"context"

	"github.com/zeromicro/go-zero/core/errorx"

	"github.com/coder-lulu/newbee-common/v2/i18n"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetApiAuthorityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetApiAuthorityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetApiAuthorityLogic {
	return &GetApiAuthorityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetApiAuthorityLogic) GetApiAuthority(req *types.IDReq) (resp *types.ApiAuthorityListResp, err error) {
	roleData, err := l.svcCtx.CoreRpc.GetRoleById(l.ctx, &core.IDReq{Id: req.Id})
	if err != nil {
		return nil, err
	}

	data, err := l.svcCtx.Casbin.GetFilteredPolicy(0, *roleData.Code)
	if err != nil {
		logx.Error("failed to get old Casbin policy", logx.Field("detail", err))
		return nil, errorx.NewInternalError(err.Error())
	}

	resp = &types.ApiAuthorityListResp{}
	resp.Msg = l.svcCtx.Trans.Trans(l.ctx, i18n.Success)
	resp.Data.Total = uint64(len(data))
	for _, v := range data {
		// Casbin策略格式: p = sub, dom, obj, act, eft
		// v[0]=角色代码, v[1]=租户ID, v[2]=API路径, v[3]=HTTP方法, v[4]=效果
		resp.Data.Data = append(resp.Data.Data, types.ApiAuthorityInfo{
			Path:   v[2], // API路径 (obj)
			Method: v[3], // HTTP方法 (act)
		})
	}
	return resp, nil
}
