package casbin

import (
	"context"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateCasbinRuleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateCasbinRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCasbinRuleLogic {
	return &CreateCasbinRuleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateCasbinRuleLogic) CreateCasbinRule(req *types.CasbinRuleInfo) (resp *types.BaseMsgResp, err error) {
	// 将API类型转换为RPC类型
	rpcReq := &core.CasbinRuleInfo{
		Id:        req.Id,
		CreatedAt: req.CreatedAt,
		UpdatedAt: req.UpdatedAt,
		TenantId:  req.TenantId,
		
		// Casbin标准字段
		Ptype: req.Ptype,
		V0:    req.V0,
		V1:    req.V1,
		V2:    req.V2,
		V3:    req.V3,
		V4:    req.V4,
		V5:    req.V5,
		
		// 业务扩展字段
		ServiceName:   req.ServiceName,
		RuleName:      req.RuleName,
		Description:   req.Description,
		Category:      req.Category,
		Version:       req.Version,
		
		// 审批流程字段
		RequireApproval: req.RequireApproval,
		ApprovalStatus:  req.ApprovalStatus,
		ApprovedBy:      req.ApprovedBy,
		ApprovedAt:      req.ApprovedAt,
		
		// 时间控制字段
		EffectiveFrom: req.EffectiveFrom,
		EffectiveTo:   req.EffectiveTo,
		IsTemporary:   req.IsTemporary,
		
		// 管理字段
		Status:     req.Status,
		Metadata:   req.Metadata,
		Tags:       req.Tags,
		UsageCount: req.UsageCount,
		LastUsedAt: req.LastUsedAt,
	}

	// 调用RPC服务
	_, err = l.svcCtx.CoreRpc.CreateCasbinRule(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("Failed to create casbin rule via RPC: %v", err)
		return &types.BaseMsgResp{
			Code: 1,
			Msg:  "创建权限规则失败: " + err.Error(),
		}, nil
	}

	// 构建成功响应
	return &types.BaseMsgResp{
		Code: 0,
		Msg:  "权限规则创建成功",
	}, nil
}
