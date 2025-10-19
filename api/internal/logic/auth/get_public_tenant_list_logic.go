package auth

import (
	"context"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPublicTenantListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPublicTenantListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPublicTenantListLogic {
	return &GetPublicTenantListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPublicTenantListLogic) GetPublicTenantList() (resp *types.PublicTenantListResp, err error) {
	// 调用 RPC 服务获取公开租户列表
	data, err := l.svcCtx.CoreRpc.GetPublicTenantList(l.ctx, &core.Empty{})
	if err != nil {
		return nil, err
	}

	// 转换数据格式为前端期望的格式
	resp = &types.PublicTenantListResp{}
	resp.Msg = l.svcCtx.Trans.Trans(l.ctx, "common.success")
	resp.TenantEnabled = data.TenantEnabled

	list := make([]types.PublicTenantInfo, 0)
	for _, v := range data.VoList {
		list = append(list, types.PublicTenantInfo{
			TenantId:    v.TenantId,
			CompanyName: v.CompanyName,
			Domain:      v.Domain,
		})
	}

	resp.VoList = list
	return resp, nil
}
