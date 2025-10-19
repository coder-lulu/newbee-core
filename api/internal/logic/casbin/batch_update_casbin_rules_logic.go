package casbin

import (
	"context"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchUpdateCasbinRulesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBatchUpdateCasbinRulesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchUpdateCasbinRulesLogic {
	return &BatchUpdateCasbinRulesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BatchUpdateCasbinRulesLogic) BatchUpdateCasbinRules(req *types.BatchUpdateCasbinRulesReq) (resp *types.BaseMsgResp, err error) {
	// 转换API类型为RPC类型
	rpcRules := make([]*core.CasbinRuleInfo, len(req.Rules))
	for i, apiRule := range req.Rules {
		rpcRules[i] = &core.CasbinRuleInfo{
			Id:        apiRule.Id,
			CreatedAt: apiRule.CreatedAt,
			UpdatedAt: apiRule.UpdatedAt,
			TenantId:  apiRule.TenantId,
			
			// Casbin标准字段
			Ptype: apiRule.Ptype,
			V0:    apiRule.V0,
			V1:    apiRule.V1,
			V2:    apiRule.V2,
			V3:    apiRule.V3,
			V4:    apiRule.V4,
			V5:    apiRule.V5,
			
			// 业务扩展字段
			ServiceName:   apiRule.ServiceName,
			RuleName:      apiRule.RuleName,
			Description:   apiRule.Description,
			Category:      apiRule.Category,
			Version:       apiRule.Version,
			
			// 审批流程字段
			RequireApproval: apiRule.RequireApproval,
			ApprovalStatus:  apiRule.ApprovalStatus,
			ApprovedBy:      apiRule.ApprovedBy,
			ApprovedAt:      apiRule.ApprovedAt,
			
			// 时间控制字段
			EffectiveFrom: apiRule.EffectiveFrom,
			EffectiveTo:   apiRule.EffectiveTo,
			IsTemporary:   apiRule.IsTemporary,
			
			// 管理字段
			Status:     apiRule.Status,
			Metadata:   apiRule.Metadata,
			Tags:       apiRule.Tags,
			UsageCount: apiRule.UsageCount,
			LastUsedAt: apiRule.LastUsedAt,
		}
	}

	rpcReq := &core.BatchUpdateCasbinRulesReq{
		Rules: rpcRules,
	}

	// 调用RPC服务
	_, err = l.svcCtx.CoreRpc.BatchUpdateCasbinRules(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("Failed to batch update casbin rules via RPC: %v", err)
		return &types.BaseMsgResp{
			Code: 1,
			Msg:  "批量更新权限规则失败: " + err.Error(),
		}, nil
	}

	// 构建成功响应
	return &types.BaseMsgResp{
		Code: 0,
		Msg:  "批量更新权限规则成功",
	}, nil
}
