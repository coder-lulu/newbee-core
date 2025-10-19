package casbin

import (
	"context"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type ValidateCasbinRuleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewValidateCasbinRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateCasbinRuleLogic {
	return &ValidateCasbinRuleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ValidateCasbinRuleLogic) ValidateCasbinRule(req *types.ValidateCasbinRuleReq) (resp *types.ValidateCasbinRuleResp, err error) {
	// 将API类型转换为RPC类型
	rpcRule := &core.CasbinRuleInfo{
		Id:        req.Rule.Id,
		CreatedAt: req.Rule.CreatedAt,
		UpdatedAt: req.Rule.UpdatedAt,
		TenantId:  req.Rule.TenantId,
		
		// Casbin标准字段
		Ptype: req.Rule.Ptype,
		V0:    req.Rule.V0,
		V1:    req.Rule.V1,
		V2:    req.Rule.V2,
		V3:    req.Rule.V3,
		V4:    req.Rule.V4,
		V5:    req.Rule.V5,
		
		// 业务扩展字段
		ServiceName:   req.Rule.ServiceName,
		RuleName:      req.Rule.RuleName,
		Description:   req.Rule.Description,
		Category:      req.Rule.Category,
		Version:       req.Rule.Version,
		
		// 审批流程字段
		RequireApproval: req.Rule.RequireApproval,
		ApprovalStatus:  req.Rule.ApprovalStatus,
		ApprovedBy:      req.Rule.ApprovedBy,
		ApprovedAt:      req.Rule.ApprovedAt,
		
		// 时间控制字段
		EffectiveFrom: req.Rule.EffectiveFrom,
		EffectiveTo:   req.Rule.EffectiveTo,
		IsTemporary:   req.Rule.IsTemporary,
		
		// 管理字段
		Status:     req.Rule.Status,
		Metadata:   req.Rule.Metadata,
		Tags:       req.Rule.Tags,
		UsageCount: req.Rule.UsageCount,
		LastUsedAt: req.Rule.LastUsedAt,
	}

	rpcReq := &core.ValidateCasbinRuleReq{
		Rule:           rpcRule,
		CheckConflicts: req.CheckConflicts,
	}

	// 调用RPC服务
	rpcResp, err := l.svcCtx.CoreRpc.ValidateCasbinRule(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("Failed to validate casbin rule via RPC: %v", err)
		return &types.ValidateCasbinRuleResp{
			BaseDataInfo: types.BaseDataInfo{
				Code: 1,
				Msg:  "验证权限规则失败: " + err.Error(),
			},
		}, nil
	}

	// 将RPC类型转换为API类型
	apiData := types.ValidationResult{
		Valid:     rpcResp.Valid,
		Errors:    rpcResp.Errors,
		Warnings:  rpcResp.Warnings,
		Conflicts: rpcResp.Conflicts,
	}

	// 构建成功响应
	return &types.ValidateCasbinRuleResp{
		BaseDataInfo: types.BaseDataInfo{
			Code: 0,
			Msg:  "验证权限规则完成",
		},
		Data: apiData,
	}, nil
}
