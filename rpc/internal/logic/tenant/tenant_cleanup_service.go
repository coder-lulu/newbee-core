package tenant

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/ent/configuration"
	"github.com/coder-lulu/newbee-core/rpc/ent/department"
	"github.com/coder-lulu/newbee-core/rpc/ent/dictionary"
	"github.com/coder-lulu/newbee-core/rpc/ent/dictionarydetail"
	"github.com/coder-lulu/newbee-core/rpc/ent/menu"
	"github.com/coder-lulu/newbee-core/rpc/ent/position"
	"github.com/coder-lulu/newbee-core/rpc/ent/role"
	"github.com/coder-lulu/newbee-core/rpc/ent/user"
	"github.com/coder-lulu/newbee-common/orm/ent/hooks"
)

// TenantCleanupService 租户清理服务
type TenantCleanupService struct {
	db     *ent.Client
	logger interface{ Infow(msg string, keysAndValues ...interface{}) }
}

// NewTenantCleanupService 创建租户清理服务
func NewTenantCleanupService(db *ent.Client, logger interface{ Infow(msg string, keysAndValues ...interface{}) }) *TenantCleanupService {
	return &TenantCleanupService{
		db:     db,
		logger: logger,
	}
}

// RollbackTenantInitialization 回滚租户初始化
func (s *TenantCleanupService) RollbackTenantInitialization(ctx context.Context, tenantID uint64) error {
	s.logger.Infow("开始回滚租户初始化", "tenant_id", tenantID)

	tx, err := s.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("开启事务失败: %w", err)
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	// 按依赖顺序删除数据
	if err := s.cleanupTenantData(ctx, tx, tenantID); err != nil {
		tx.Rollback()
		return fmt.Errorf("清理租户数据失败: %w", err)
	}

	// 重置租户配置
	if err := s.resetTenantConfig(ctx, tx, tenantID); err != nil {
		tx.Rollback()
		return fmt.Errorf("重置租户配置失败: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交回滚事务失败: %w", err)
	}

	s.logger.Infow("租户初始化回滚成功", "tenant_id", tenantID)
	return nil
}

// cleanupTenantData 清理租户数据
func (s *TenantCleanupService) cleanupTenantData(ctx context.Context, tx *ent.Tx, tenantID uint64) error {
	// 1. 删除用户数据（最高级别，有外键依赖）
	if err := s.cleanupUsers(ctx, tx, tenantID); err != nil {
		return err
	}

	// 2. 删除角色数据
	if err := s.cleanupRoles(ctx, tx, tenantID); err != nil {
		return err
	}

	// 3. 删除职位数据
	if err := s.cleanupPositions(ctx, tx, tenantID); err != nil {
		return err
	}

	// 4. 删除部门数据
	if err := s.cleanupDepartments(ctx, tx, tenantID); err != nil {
		return err
	}

	// 5. 删除字典详情数据
	if err := s.cleanupDictionaryDetails(ctx, tx, tenantID); err != nil {
		return err
	}

	// 6. 删除字典数据
	if err := s.cleanupDictionaries(ctx, tx, tenantID); err != nil {
		return err
	}

	// 7. 删除配置数据
	if err := s.cleanupConfigurations(ctx, tx, tenantID); err != nil {
		return err
	}

	// 8. 删除租户菜单数据
	if err := s.cleanupMenus(ctx, tx, tenantID); err != nil {
		return err
	}

	// 注意：API权限是系统级的，不需要为每个租户单独清理

	return nil
}

// cleanupUsers 清理用户数据
func (s *TenantCleanupService) cleanupUsers(ctx context.Context, tx *ent.Tx, tenantID uint64) error {
	deletedCount, err := tx.User.Delete().
		Where(user.TenantIDEQ(tenantID)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除用户数据失败: %w", err)
	}
	s.logger.Infow("清理用户数据完成", "tenant_id", tenantID, "deleted_count", deletedCount)
	return nil
}

// cleanupRoles 清理角色数据
func (s *TenantCleanupService) cleanupRoles(ctx context.Context, tx *ent.Tx, tenantID uint64) error {
	deletedCount, err := tx.Role.Delete().
		Where(role.TenantIDEQ(tenantID)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除角色数据失败: %w", err)
	}
	s.logger.Infow("清理角色数据完成", "tenant_id", tenantID, "deleted_count", deletedCount)
	return nil
}

// cleanupPositions 清理职位数据
func (s *TenantCleanupService) cleanupPositions(ctx context.Context, tx *ent.Tx, tenantID uint64) error {
	deletedCount, err := tx.Position.Delete().
		Where(position.TenantIDEQ(tenantID)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除职位数据失败: %w", err)
	}
	s.logger.Infow("清理职位数据完成", "tenant_id", tenantID, "deleted_count", deletedCount)
	return nil
}

// cleanupDepartments 清理部门数据
func (s *TenantCleanupService) cleanupDepartments(ctx context.Context, tx *ent.Tx, tenantID uint64) error {
	deletedCount, err := tx.Department.Delete().
		Where(department.TenantIDEQ(tenantID)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除部门数据失败: %w", err)
	}
	s.logger.Infow("清理部门数据完成", "tenant_id", tenantID, "deleted_count", deletedCount)
	return nil
}

// cleanupDictionaryDetails 清理字典详情数据
func (s *TenantCleanupService) cleanupDictionaryDetails(ctx context.Context, tx *ent.Tx, tenantID uint64) error {
	deletedCount, err := tx.DictionaryDetail.Delete().
		Where(dictionarydetail.TenantIDEQ(tenantID)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除字典详情数据失败: %w", err)
	}
	s.logger.Infow("清理字典详情数据完成", "tenant_id", tenantID, "deleted_count", deletedCount)
	return nil
}

// cleanupDictionaries 清理字典数据
func (s *TenantCleanupService) cleanupDictionaries(ctx context.Context, tx *ent.Tx, tenantID uint64) error {
	deletedCount, err := tx.Dictionary.Delete().
		Where(dictionary.TenantIDEQ(tenantID)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除字典数据失败: %w", err)
	}
	s.logger.Infow("清理字典数据完成", "tenant_id", tenantID, "deleted_count", deletedCount)
	return nil
}

// cleanupConfigurations 清理配置数据
func (s *TenantCleanupService) cleanupConfigurations(ctx context.Context, tx *ent.Tx, tenantID uint64) error {
	deletedCount, err := tx.Configuration.Delete().
		Where(configuration.TenantIDEQ(tenantID)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除配置数据失败: %w", err)
	}
	s.logger.Infow("清理配置数据完成", "tenant_id", tenantID, "deleted_count", deletedCount)
	return nil
}

// cleanupMenus 清理菜单数据
func (s *TenantCleanupService) cleanupMenus(ctx context.Context, tx *ent.Tx, tenantID uint64) error {
	deletedCount, err := tx.Menu.Delete().
		Where(menu.TenantIDEQ(tenantID)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除菜单数据失败: %w", err)
	}
	s.logger.Infow("清理菜单数据完成", "tenant_id", tenantID, "deleted_count", deletedCount)
	return nil
}

// resetTenantConfig 重置租户配置
func (s *TenantCleanupService) resetTenantConfig(ctx context.Context, tx *ent.Tx, tenantID uint64) error {
	_, err := tx.Tenant.UpdateOneID(tenantID).
		ClearConfig().
		Save(hooks.NewSystemContext(ctx))
	if err != nil {
		return fmt.Errorf("重置租户配置失败: %w", err)
	}
	return nil
}

// GetTenantInitializationStatus 获取租户初始化状态
func (s *TenantCleanupService) GetTenantInitializationStatus(ctx context.Context, tenantID uint64) (*TenantInitConfig, error) {
	tenantInfo, err := s.db.Tenant.Get(hooks.NewSystemContext(ctx), tenantID)
	if err != nil {
		return nil, fmt.Errorf("获取租户信息失败: %w", err)
	}

	if tenantInfo.Config == nil || len(tenantInfo.Config) == 0 {
		return &TenantInitConfig{
			Status: "not_initialized",
		}, nil
	}

	var initConfig TenantInitConfig
	configBytes, err := json.Marshal(tenantInfo.Config)
	if err != nil {
		return nil, fmt.Errorf("序列化配置失败: %w", err)
	}
	
	if err := json.Unmarshal(configBytes, &initConfig); err != nil {
		return nil, fmt.Errorf("解析初始化配置失败: %w", err)
	}

	return &initConfig, nil
}

// ValidateTenantInitialization 验证租户初始化完整性
func (s *TenantCleanupService) ValidateTenantInitialization(ctx context.Context, tenantID uint64) ([]ValidationError, error) {
	var errors []ValidationError

	// 检查字典数据
	dictCount, err := s.db.Dictionary.Query().
		Where(dictionary.TenantIDEQ(tenantID)).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("检查字典数据失败: %w", err)
	}
	if dictCount == 0 {
		errors = append(errors, ValidationError{
			Component: "dictionaries",
			Message:   "未找到字典数据",
			Severity:  "error",
		})
	}

	// 检查配置数据
	configCount, err := s.db.Configuration.Query().
		Where(configuration.TenantIDEQ(tenantID)).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("检查配置数据失败: %w", err)
	}
	if configCount == 0 {
		errors = append(errors, ValidationError{
			Component: "configurations",
			Message:   "未找到配置数据",
			Severity:  "error",
		})
	}

	// 检查部门数据
	deptCount, err := s.db.Department.Query().
		Where(department.TenantIDEQ(tenantID)).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("检查部门数据失败: %w", err)
	}
	if deptCount == 0 {
		errors = append(errors, ValidationError{
			Component: "departments",
			Message:   "未找到部门数据",
			Severity:  "error",
		})
	}

	// 检查用户数据
	userCount, err := s.db.User.Query().
		Where(user.TenantIDEQ(tenantID)).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("检查用户数据失败: %w", err)
	}
	if userCount == 0 {
		errors = append(errors, ValidationError{
			Component: "users",
			Message:   "未找到用户数据",
			Severity:  "error",
		})
	}

	// 检查角色数据
	roleCount, err := s.db.Role.Query().
		Where(role.TenantIDEQ(tenantID)).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("检查角色数据失败: %w", err)
	}
	if roleCount == 0 {
		errors = append(errors, ValidationError{
			Component: "roles",
			Message:   "未找到角色数据",
			Severity:  "error",
		})
	}

	return errors, nil
}

// ValidationError 验证错误
type ValidationError struct {
	Component string `json:"component"`
	Message   string `json:"message"`
	Severity  string `json:"severity"` // error, warning, info
}