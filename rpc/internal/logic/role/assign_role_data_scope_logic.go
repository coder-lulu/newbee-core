package role

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/coder-lulu/newbee-common/v2/i18n"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/hooks"
	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/ent/casbinrule"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/entx"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
)

type AssignRoleDataScopeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAssignRoleDataScopeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignRoleDataScopeLogic {
	return &AssignRoleDataScopeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AssignRoleDataScopeLogic) AssignRoleDataScope(in *core.RoleDataScopeReq) (*core.BaseResp, error) {
	// 🔥 Phase 2: 参数验证
	if err := l.validateDataScopeRequest(in); err != nil {
		return nil, err
	}

	// 🔥 Phase 2: 使用事务同时更新sys_roles和sys_casbin_rules
	err := entx.WithTx(l.ctx, l.svcCtx.DB, func(tx *ent.Tx) error {
		// 1. 获取角色信息
		role, err := tx.Role.Get(l.ctx, in.Id)
		if err != nil {
			return fmt.Errorf("角色不存在: %w", err)
		}

		// 2. 🔥 Phase 3: 只更新custom_dept_ids (data_scope字段已移除)
		err = tx.Role.UpdateOneID(in.Id).
			SetNotNilCustomDeptIds(in.CustomDeptIds).
			Exec(l.ctx)
		if err != nil {
			return fmt.Errorf("更新角色custom_dept_ids失败: %w", err)
		}

		// 3. 🔥 Phase 2: 同步更新sys_casbin_rules表中的数据权限规则
		err = l.updateCasbinDataPermRules(tx, role, in)
		if err != nil {
			return fmt.Errorf("更新Casbin数据权限规则失败: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 4. 触发Casbin策略重新加载（通过Redis Watcher自动触发）
	logx.Infow("✅ 更新角色数据权限成功",
		logx.Field("role_id", in.Id),
		logx.Field("data_scope", in.DataScope),
		logx.Field("custom_dept_ids", in.CustomDeptIds))

	return &core.BaseResp{Msg: i18n.UpdateSuccess}, nil
}

// 🔥 Phase 2: 验证数据权限范围请求参数
func (l *AssignRoleDataScopeLogic) validateDataScopeRequest(req *core.RoleDataScopeReq) error {
	// 验证数据权限范围是否合法
	validScopes := map[uint32]bool{
		1: true, // all
		2: true, // custom_dept
		3: true, // own_dept_and_sub
		4: true, // own_dept
		5: true, // own
	}

	if !validScopes[req.DataScope] {
		return errorx.NewInvalidArgumentError(fmt.Sprintf("无效的数据权限范围: %d", req.DataScope))
	}

	// custom_dept时必须提供部门ID列表
	if req.DataScope == 2 && len(req.CustomDeptIds) == 0 {
		return errorx.NewInvalidArgumentError("自定义部门权限时必须提供部门ID列表")
	}

	return nil
}

// 🔥 Phase 2: 更新Casbin数据权限规则
func (l *AssignRoleDataScopeLogic) updateCasbinDataPermRules(
	tx *ent.Tx,
	role *ent.Role,
	req *core.RoleDataScopeReq,
) error {
	// 使用SystemContext绕过租户Hook
	systemCtx := hooks.NewSystemContext(l.ctx)

	// 1. 删除现有的数据权限规则（ptype=d）
	_, err := tx.CasbinRule.Delete().
		Where(
			casbinrule.PtypeEQ("d"),              // 数据权限规则
			casbinrule.V0EQ(role.Code),           // 角色代码
			casbinrule.TenantIDEQ(role.TenantID), // 租户ID
		).
		Exec(systemCtx)
	if err != nil {
		return fmt.Errorf("删除旧数据权限规则失败: %w", err)
	}

	// 2. 将dataScope枚举值转换为字符串
	dataScopeStr := l.dataScopeEnumToString(req.DataScope)

	// 3. 构造v4字段（自定义部门列表JSON）
	v4 := ""
	if req.DataScope == 2 && len(req.CustomDeptIds) > 0 { // custom_dept
		// 将部门ID列表转换为JSON数组
		deptIDStrs := make([]string, len(req.CustomDeptIds))
		for i, id := range req.CustomDeptIds {
			deptIDStrs[i] = fmt.Sprintf("%d", id)
		}
		v4JSON, _ := json.Marshal(deptIDStrs)
		v4 = string(v4JSON)
	}

	// 4. 创建新的数据权限规则
	_, err = tx.CasbinRule.Create().
		SetPtype("d").                              // 数据权限规则
		SetV0(role.Code).                           // subject: 角色代码
		SetV1(fmt.Sprintf("%d", role.TenantID)).    // domain: 租户ID
		SetV2("*").                                 // object: 资源类型（* 表示所有）
		SetV3(dataScopeStr).                        // action: 数据权限范围
		SetV4(v4).                                  // effect: 自定义部门列表
		SetServiceName("core").                     // 服务名称
		SetRuleName(fmt.Sprintf("%s数据权限", role.Name)).
		SetDescription(fmt.Sprintf("角色%s的数据权限规则，数据范围：%s", role.Name, dataScopeStr)).
		SetCategory("data_permission").
		SetVersion("1.0.0").
		SetRequireApproval(false).
		SetApprovalStatus("approved").
		SetStatus(1).
		SetTenantID(role.TenantID).
		Save(systemCtx)

	if err != nil {
		return fmt.Errorf("创建新数据权限规则失败: %w", err)
	}

	logx.Infow("✅ Casbin数据权限规则已更新",
		logx.Field("role_code", role.Code),
		logx.Field("role_name", role.Name),
		logx.Field("tenant_id", role.TenantID),
		logx.Field("data_scope", dataScopeStr),
		logx.Field("custom_dept_ids", req.CustomDeptIds))

	// 5. 触发Redis Watcher通知
	updateMsg := fmt.Sprintf("UpdatePolicy:tenant_%d:data_perm", role.TenantID)
	err = l.svcCtx.Redis.Publish(l.ctx, "casbin_watcher", updateMsg).Err()
	if err != nil {
		logx.Errorw("failed to publish data perm policy update notification", logx.Field("error", err.Error()))
		// 不返回错误，因为策略已经写入数据库，只是通知失败
	} else {
		logx.Info("✅ Published data permission policy update notification to Redis")
	}

	return nil
}

// 🔥 Phase 2: 将dataScope枚举值转换为字符串
func (l *AssignRoleDataScopeLogic) dataScopeEnumToString(dataScope uint32) string {
	switch dataScope {
	case 1:
		return "all"
	case 2:
		return "custom_dept"
	case 3:
		return "own_dept_and_sub"
	case 4:
		return "own_dept"
	case 5:
		return "own"
	default:
		return "own" // 默认值
	}
}
