package casbin

import (
	"context"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCasbinRuleListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCasbinRuleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCasbinRuleListLogic {
	return &GetCasbinRuleListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCasbinRuleListLogic) GetCasbinRuleList(req *types.CasbinRuleListReq) (resp *types.CasbinRuleListResp, err error) {
	// 将API类型转换为RPC类型
	rpcReq := &core.CasbinRuleListReq{
		Page:                req.Page,
		PageSize:            req.PageSize,
		ServiceName:         req.ServiceName,
		Ptype:               req.Ptype,
		V0:                  req.V0,
		V1:                  req.V1,
		Status:              req.Status,
		ApprovalStatus:      req.ApprovalStatus,
		Category:            req.Category,
		IsTemporary:         req.IsTemporary,
		EffectiveFromStart:  req.EffectiveFromStart,
		EffectiveFromEnd:    req.EffectiveFromEnd,
		Keyword:             req.Keyword,
	}

	// 调用RPC服务
	rpcResp, err := l.svcCtx.CoreRpc.GetCasbinRuleList(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("Failed to get casbin rule list via RPC: %v", err)
		return &types.CasbinRuleListResp{
			BaseDataInfo: types.BaseDataInfo{
				Code: 1,
				Msg:  "获取权限规则列表失败: " + err.Error(),
			},
		}, nil
	}

	// 将RPC类型转换为API类型
	var apiList []types.CasbinRuleInfo
	for _, rpcItem := range rpcResp.Data {
		apiItem := types.CasbinRuleInfo{
			BaseIDInfo: types.BaseIDInfo{
				Id:        rpcItem.Id,
				CreatedAt: rpcItem.CreatedAt,
				UpdatedAt: rpcItem.UpdatedAt,
			},
			TenantId: rpcItem.TenantId,
			
			// Casbin标准字段
			Ptype: rpcItem.Ptype,
			V0:    rpcItem.V0,
			V1:    rpcItem.V1,
			V2:    rpcItem.V2,
			V3:    rpcItem.V3,
			V4:    rpcItem.V4,
			V5:    rpcItem.V5,
			
			// 业务扩展字段
			ServiceName:   rpcItem.ServiceName,
			RuleName:      rpcItem.RuleName,
			Description:   rpcItem.Description,
			Category:      rpcItem.Category,
			Version:       rpcItem.Version,
			
			// 审批流程字段
			RequireApproval: rpcItem.RequireApproval,
			ApprovalStatus:  rpcItem.ApprovalStatus,
			ApprovedBy:      rpcItem.ApprovedBy,
			ApprovedAt:      rpcItem.ApprovedAt,
			
			// 时间控制字段
			EffectiveFrom: rpcItem.EffectiveFrom,
			EffectiveTo:   rpcItem.EffectiveTo,
			IsTemporary:   rpcItem.IsTemporary,
			
			// 管理字段
			Status:     rpcItem.Status,
			Metadata:   rpcItem.Metadata,
			Tags:       rpcItem.Tags,
			UsageCount: rpcItem.UsageCount,
			LastUsedAt: rpcItem.LastUsedAt,
		}
		apiList = append(apiList, apiItem)
	}

	// 构建成功响应
	return &types.CasbinRuleListResp{
		BaseDataInfo: types.BaseDataInfo{
			Code: 0,
			Msg:  "获取权限规则列表成功",
		},
		Data: types.CasbinRuleListInfo{
			BaseListInfo: types.BaseListInfo{
				Total: rpcResp.Total,
			},
			Data: apiList,
		},
	}, nil
}
