package casbin

import (
	"context"
	"fmt"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchCreateCasbinRulesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchCreateCasbinRulesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchCreateCasbinRulesLogic {
	return &BatchCreateCasbinRulesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 批量操作
func (l *BatchCreateCasbinRulesLogic) BatchCreateCasbinRules(in *core.BatchCreateCasbinRulesReq) (*core.BaseResp, error) {
	// 验证请求
	if len(in.Rules) == 0 {
		return nil, fmt.Errorf("rules cannot be empty")
	}

	// 使用事务批量创建
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

	var createdCount int
	createLogic := NewCreateCasbinRuleLogic(l.ctx, l.svcCtx)

	for i, rule := range in.Rules {
		// 验证每个规则的必需字段
		if rule.Ptype == "" {
			err = fmt.Errorf("rule %d: ptype is required", i)
			return nil, err
		}
		if rule.ServiceName == "" {
			err = fmt.Errorf("rule %d: service_name is required", i)
			return nil, err
		}

		// 重用单个创建逻辑（包含 Casbin 同步）
		_, createErr := createLogic.CreateCasbinRule(rule)
		if createErr != nil {
			err = fmt.Errorf("create rule %d failed: %v", i, createErr)
			return nil, err
		}
		createdCount++
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		l.Logger.Errorf("Commit transaction failed: %v", err)
		return nil, fmt.Errorf("commit transaction failed: %v", err)
	}

	l.Logger.Infof("Batch created %d casbin rules successfully", createdCount)

	return &core.BaseResp{
		Msg: fmt.Sprintf("成功批量创建 %d 条权限规则", createdCount),
	}, nil
}
