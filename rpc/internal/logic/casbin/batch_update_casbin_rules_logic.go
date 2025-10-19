package casbin

import (
	"context"
	"fmt"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchUpdateCasbinRulesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchUpdateCasbinRulesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchUpdateCasbinRulesLogic {
	return &BatchUpdateCasbinRulesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BatchUpdateCasbinRulesLogic) BatchUpdateCasbinRules(in *core.BatchUpdateCasbinRulesReq) (*core.BaseResp, error) {
	// 验证请求
	if len(in.Rules) == 0 {
		return nil, fmt.Errorf("rules cannot be empty")
	}

	// 使用事务批量更新
	tx, err := l.svcCtx.DB.Tx(l.ctx)
	if err != nil {
		l.Logger.Errorf("Start transaction failed: %v", err)
		return nil, fmt.Errorf("start transaction failed: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var updatedCount int
	updateLogic := NewUpdateCasbinRuleLogic(l.ctx, l.svcCtx)

	for i, rule := range in.Rules {
		// 验证每个规则的ID
		if rule.Id == nil || *rule.Id == 0 {
			err = fmt.Errorf("rule %d: id is required", i)
			return nil, err
		}

		// 重用单个更新逻辑
		_, updateErr := updateLogic.UpdateCasbinRule(rule)
		if updateErr != nil {
			err = fmt.Errorf("update rule %d failed: %v", i, updateErr)
			return nil, err
		}
		updatedCount++
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		l.Logger.Errorf("Commit transaction failed: %v", err)
		return nil, fmt.Errorf("commit transaction failed: %v", err)
	}

	l.Logger.Infof("Batch updated %d casbin rules successfully", updatedCount)

	return &core.BaseResp{
		Msg: fmt.Sprintf("成功批量更新 %d 条权限规则", updatedCount),
	}, nil
}
