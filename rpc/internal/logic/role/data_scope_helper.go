package role

import (
	"context"
	"fmt"

	"github.com/coder-lulu/newbee-common/v2/orm/ent/hooks"
	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/ent/casbinrule"
	"github.com/zeromicro/go-zero/core/logx"
)

// 🔥 Phase 3: 数据权限范围查询辅助函数

// getDataScopeFromCasbin 从sys_casbin_rules查询角色的数据权限范围
func getDataScopeFromCasbin(ctx context.Context, db *ent.Client, roleCode string, tenantID uint64) (uint32, error) {
	// 使用SystemContext绕过租户隔离
	systemCtx := hooks.NewSystemContext(ctx)

	// 查询数据权限规则（ptype='d'）
	rule, err := db.CasbinRule.Query().
		Where(
			casbinrule.PtypeEQ("d"),                      // 数据权限规则
			casbinrule.V0EQ(roleCode),                    // 角色代码
			casbinrule.V1EQ(fmt.Sprintf("%d", tenantID)), // 租户ID
		).
		First(systemCtx)

	if err != nil {
		// 如果没有找到规则，返回默认值
		if ent.IsNotFound(err) {
			logx.Infow("No data permission rule found, using default",
				logx.Field("role_code", roleCode),
				logx.Field("tenant_id", tenantID))
			return 5, nil // 默认为 own (最严格的权限)
		}
		return 0, err
	}

	// 将v3字段（数据权限范围字符串）转换为枚举值
	return dataScopeStringToEnum(rule.V3), nil
}

// dataScopeStringToEnum 将数据权限范围字符串转换为枚举值
func dataScopeStringToEnum(dataScope string) uint32 {
	switch dataScope {
	case "all":
		return 1
	case "custom_dept":
		return 2
	case "own_dept_and_sub":
		return 3
	case "own_dept":
		return 4
	case "own":
		return 5
	default:
		return 5 // 默认为 own (最严格的权限)
	}
}

// dataScopeEnumToString 将数据权限范围枚举值转换为字符串（用于更新操作）
func dataScopeEnumToString(dataScope uint32) string {
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
