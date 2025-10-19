package plugins

import (
	"context"
	"fmt"
	"strconv"

	"github.com/coder-lulu/newbee-common/middleware/keys"
	"github.com/coder-lulu/newbee-common/orm/ent/hooks"
	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/ent/configuration"
	"github.com/coder-lulu/newbee-core/rpc/ent/department"
	"github.com/coder-lulu/newbee-core/rpc/ent/dictionary"
	"github.com/coder-lulu/newbee-core/rpc/ent/dictionarydetail"
	"github.com/coder-lulu/newbee-core/rpc/ent/menu"
	"github.com/coder-lulu/newbee-core/rpc/ent/position"
	"github.com/coder-lulu/newbee-core/rpc/ent/role"
	"github.com/coder-lulu/newbee-core/rpc/ent/user"

	"github.com/zeromicro/go-zero/core/logx"
)

// Rollback 回滚租户初始化
func (p *CoreTenantPlugin) Rollback(ctx context.Context, tenantID uint64) error {
	p.logger.Infow("Core plugin starting rollback",
		logx.Field("tenant_id", tenantID))

	// 实现核心数据的回滚逻辑
	// 注意：这需要仔细实现，避免数据丢失

	// 获取租户上下文
	// 使用统一的ContextManager设置租户ID
	cm := keys.NewContextManager()
	tenantIDStr := strconv.FormatUint(tenantID, 10)
	tenantCtx := cm.SetTenantID(ctx, tenantIDStr)

	// 启动事务
	tx, err := p.svcCtx.DB.Tx(tenantCtx)
	if err != nil {
		return fmt.Errorf("failed to start rollback transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 删除租户相关的初始化数据（按创建的相反顺序）

	// 1. 删除用户角色关联
	if err = p.rollbackUserRoles(tenantCtx, tx, tenantID); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to rollback user roles: %w", err)
	}

	// 2. 删除管理员用户
	if err = p.rollbackUsers(tenantCtx, tx, tenantID); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to rollback users: %w", err)
	}

	// 3. 删除角色菜单关联
	if err = p.rollbackRoleMenus(tenantCtx, tx, tenantID); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to rollback role menus: %w", err)
	}

	// 4. 删除角色
	if err = p.rollbackRoles(tenantCtx, tx, tenantID); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to rollback roles: %w", err)
	}

	// 5. 删除职位
	if err = p.rollbackPositions(tenantCtx, tx, tenantID); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to rollback positions: %w", err)
	}

	// 6. 删除部门
	if err = p.rollbackDepartments(tenantCtx, tx, tenantID); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to rollback departments: %w", err)
	}

	// 7. 删除菜单副本
	if err = p.rollbackMenus(tenantCtx, tx, tenantID); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to rollback menus: %w", err)
	}

	// 8. 删除配置
	if err = p.rollbackConfigurations(tenantCtx, tx, tenantID); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to rollback configurations: %w", err)
	}

	// 9. 删除字典详情
	if err = p.rollbackDictionaryDetails(tenantCtx, tx, tenantID); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to rollback dictionary details: %w", err)
	}

	// 10. 删除字典
	if err = p.rollbackDictionaries(tenantCtx, tx, tenantID); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to rollback dictionaries: %w", err)
	}

	// 11. 重置租户初始化状态
	_, err = tx.Tenant.UpdateOneID(tenantID).
		ClearConfig().
		Save(hooks.NewSystemContext(tenantCtx))
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to reset tenant status: %w", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback transaction: %w", err)
	}

	p.logger.Infow("Core plugin rollback completed successfully",
		logx.Field("tenant_id", tenantID))

	return nil
}

// 以下是回滚辅助方法
// 注意：这些方法需要谨慎实现，确保不会意外删除重要数据

func (p *CoreTenantPlugin) rollbackUserRoles(ctx context.Context, tx *ent.Tx, tenantID uint64) error {
	// 删除用户角色关联
	_, err := tx.User.Update().
		Where(user.TenantIDEQ(tenantID)).
		ClearRoles().
		Save(ctx)
	return err
}

func (p *CoreTenantPlugin) rollbackUsers(ctx context.Context, tx *ent.Tx, tenantID uint64) error {
	// 删除租户用户
	_, err := tx.User.Delete().
		Where(user.TenantIDEQ(tenantID)).
		Exec(ctx)
	return err
}

func (p *CoreTenantPlugin) rollbackRoleMenus(ctx context.Context, tx *ent.Tx, tenantID uint64) error {
	// 删除角色菜单关联
	_, err := tx.Role.Update().
		Where(role.TenantIDEQ(tenantID)).
		ClearMenus().
		Save(ctx)
	return err
}

func (p *CoreTenantPlugin) rollbackRoles(ctx context.Context, tx *ent.Tx, tenantID uint64) error {
	// 删除租户角色
	_, err := tx.Role.Delete().
		Where(role.TenantIDEQ(tenantID)).
		Exec(ctx)
	return err
}

func (p *CoreTenantPlugin) rollbackPositions(ctx context.Context, tx *ent.Tx, tenantID uint64) error {
	// 删除租户职位
	_, err := tx.Position.Delete().
		Where(position.TenantIDEQ(tenantID)).
		Exec(ctx)
	return err
}

func (p *CoreTenantPlugin) rollbackDepartments(ctx context.Context, tx *ent.Tx, tenantID uint64) error {
	// 删除租户部门
	_, err := tx.Department.Delete().
		Where(department.TenantIDEQ(tenantID)).
		Exec(ctx)
	return err
}

func (p *CoreTenantPlugin) rollbackMenus(ctx context.Context, tx *ent.Tx, tenantID uint64) error {
	// 删除租户菜单副本
	_, err := tx.Menu.Delete().
		Where(menu.TenantIDEQ(tenantID)).
		Exec(ctx)
	return err
}

func (p *CoreTenantPlugin) rollbackConfigurations(ctx context.Context, tx *ent.Tx, tenantID uint64) error {
	// 删除租户配置
	_, err := tx.Configuration.Delete().
		Where(configuration.TenantIDEQ(tenantID)).
		Exec(ctx)
	return err
}

func (p *CoreTenantPlugin) rollbackDictionaryDetails(ctx context.Context, tx *ent.Tx, tenantID uint64) error {
	// 删除租户字典详情
	_, err := tx.DictionaryDetail.Delete().
		Where(dictionarydetail.TenantIDEQ(tenantID)).
		Exec(ctx)
	return err
}

func (p *CoreTenantPlugin) rollbackDictionaries(ctx context.Context, tx *ent.Tx, tenantID uint64) error {
	// 删除租户字典
	_, err := tx.Dictionary.Delete().
		Where(dictionary.TenantIDEQ(tenantID)).
		Exec(ctx)
	return err
}
