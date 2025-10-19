package casbin

import (
	"context"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCasbinRuleByIdLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCasbinRuleByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCasbinRuleByIdLogic {
	return &GetCasbinRuleByIdLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCasbinRuleByIdLogic) GetCasbinRuleById(req *types.IDReq) (resp *types.CasbinRuleInfoResp, err error) {
	// 将API类型转换为RPC类型
	rpcReq := &core.IDReq{
		Id: req.Id,
	}

	// 调用RPC服务
	rpcResp, err := l.svcCtx.CoreRpc.GetCasbinRuleById(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("Failed to get casbin rule by id via RPC: %v", err)
		return &types.CasbinRuleInfoResp{
			BaseDataInfo: types.BaseDataInfo{
				Code: 1,
				Msg:  "获取权限规则失败: " + err.Error(),
			},
		}, nil
	}

	// 将RPC类型转换为API类型
	apiData := types.CasbinRuleInfo{
		BaseIDInfo: types.BaseIDInfo{
			Id:        rpcResp.Id,
			CreatedAt: rpcResp.CreatedAt,
			UpdatedAt: rpcResp.UpdatedAt,
		},
		TenantId: rpcResp.TenantId,
		
		// Casbin标准字段
		Ptype: rpcResp.Ptype,
		V0:    rpcResp.V0,
		V1:    rpcResp.V1,
		V2:    rpcResp.V2,
		V3:    rpcResp.V3,
		V4:    rpcResp.V4,
		V5:    rpcResp.V5,
		
		// 业务扩展字段
		ServiceName:   rpcResp.ServiceName,
		RuleName:      rpcResp.RuleName,
		Description:   rpcResp.Description,
		Category:      rpcResp.Category,
		Version:       rpcResp.Version,
		
		// 审批流程字段
		RequireApproval: rpcResp.RequireApproval,
		ApprovalStatus:  rpcResp.ApprovalStatus,
		ApprovedBy:      rpcResp.ApprovedBy,
		ApprovedAt:      rpcResp.ApprovedAt,
		
		// 时间控制字段
		EffectiveFrom: rpcResp.EffectiveFrom,
		EffectiveTo:   rpcResp.EffectiveTo,
		IsTemporary:   rpcResp.IsTemporary,
		
		// 管理字段
		Status:     rpcResp.Status,
		Metadata:   rpcResp.Metadata,
		Tags:       rpcResp.Tags,
		UsageCount: rpcResp.UsageCount,
		LastUsedAt: rpcResp.LastUsedAt,
	}

	// 构建成功响应
	return &types.CasbinRuleInfoResp{
		BaseDataInfo: types.BaseDataInfo{
			Code: 0,
			Msg:  "获取权限规则成功",
		},
		Data: apiData,
	}, nil
}
