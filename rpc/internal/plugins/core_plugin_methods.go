package plugins

import (
	"context"
	"fmt"

	"github.com/coder-lulu/newbee-common/v2/orm/ent/entenum"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/hooks"
	"github.com/coder-lulu/newbee-common/v2/tenant"
	"github.com/coder-lulu/newbee-common/v2/utils/encrypt"
	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/ent/casbinrule"
	"github.com/coder-lulu/newbee-core/rpc/ent/configuration"
	"github.com/coder-lulu/newbee-core/rpc/ent/dictionary"
	"github.com/coder-lulu/newbee-core/rpc/ent/dictionarydetail"
	"github.com/coder-lulu/newbee-core/rpc/ent/menu"
	"github.com/coder-lulu/newbee-core/rpc/ent/position"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"

	"github.com/gofrs/uuid/v5"
	"github.com/zeromicro/go-zero/core/logx"
)

// initDictionaries 初始化字典数据
func (p *CoreTenantPlugin) initDictionaries(ctx context.Context, tx *ent.Tx, tenantID uint64) error {
	// 收集当前租户已有的字典，避免重复插入
	ctxWithTenant := hooks.SetTenantIDToContext(context.Background(), tenantID)
	existingDicts, err := tx.Dictionary.Query().
		Where(dictionary.TenantIDEQ(tenantID)).
		All(ctxWithTenant)
	if err != nil {
		return dberrorhandler.DefaultEntError(p.logger, err, nil)
	}

	existingNames := make(map[string]struct{}, len(existingDicts))
	for _, dict := range existingDicts {
		existingNames[dict.Name] = struct{}{}
	}

	// 使用系统上下文从超级租户读取模板字典
	systemCtx := hooks.NewSystemContext(ctx)
	superTenantID := entenum.GetTenantDefaultId(systemCtx)
	baseDicts, err := tx.Dictionary.Query().
		Where(dictionary.TenantIDEQ(superTenantID)).
		WithDictionaryDetails(func(q *ent.DictionaryDetailQuery) {
			q.Order(dictionarydetail.BySort(), dictionarydetail.ByID())
		}).
		All(systemCtx)
	if err != nil {
		return dberrorhandler.DefaultEntError(p.logger, err, nil)
	}

	if len(baseDicts) == 0 {
		logx.Infow("no template dictionaries found for super tenant",
			logx.Field("super_tenant_id", superTenantID))
		return nil
	}

	created := 0
	for _, base := range baseDicts {
		if base.Name == "" {
			continue
		}
		if _, exists := existingNames[base.Name]; exists {
			continue
		}

		newDict, err := tx.Dictionary.Create().
			SetTitle(base.Title).
			SetName(base.Name).
			SetDesc(base.Desc).
			SetStatus(base.Status).
			SetTenantID(tenantID).
			Save(ctxWithTenant)
		if err != nil {
			return dberrorhandler.DefaultEntError(p.logger, err, base)
		}

		existingNames[base.Name] = struct{}{}
		created++

		if len(base.Edges.DictionaryDetails) == 0 {
			continue
		}

		detailCreates := make([]*ent.DictionaryDetailCreate, 0, len(base.Edges.DictionaryDetails))
		for _, detail := range base.Edges.DictionaryDetails {
			detailCreates = append(detailCreates, tx.DictionaryDetail.Create().
				SetTitle(detail.Title).
				SetValue(detail.Value).
				SetListClass(detail.ListClass).
				SetCSSClass(detail.CSSClass).
				SetIsDefault(detail.IsDefault).
				SetSort(detail.Sort).
				SetStatus(detail.Status).
				SetTenantID(tenantID).
				SetDictionaryID(newDict.ID),
			)
		}

		if len(detailCreates) > 0 {
			if err := tx.DictionaryDetail.CreateBulk(detailCreates...).Exec(ctxWithTenant); err != nil {
				return dberrorhandler.DefaultEntError(p.logger, err, nil)
			}
		}
	}

	logx.Infow("tenant dictionaries synchronized",
		logx.Field("tenant_id", tenantID),
		logx.Field("created", created),
		logx.Field("template_count", len(baseDicts)))

	return nil
}

// initConfigurations 初始化系统配置
func (p *CoreTenantPlugin) initConfigurations(ctx context.Context, tx *ent.Tx, tenantID uint64) error {
	// 检查是否已存在配置数据
	ctxWithTenant := hooks.SetTenantIDToContext(context.Background(), tenantID)
	existingCount, err := tx.Configuration.Query().
		Where(configuration.TenantIDEQ(tenantID)).
		Count(ctxWithTenant)
	if err != nil {
		return dberrorhandler.DefaultEntError(p.logger, err, nil)
	}
	if existingCount > 0 {
		logx.Infow("Configurations already exist, skipping", logx.Field("tenant_id", tenantID))
		return nil
	}

	// 创建基础配置
	configs := []*ent.ConfigurationCreate{
		tx.Configuration.Create().
			SetName("系统名称").
			SetKey("system.title").
			SetValue("NewBee管理系统").
			SetCategory("system").
			SetRemark("系统页面标题").
			SetSort(1).
			SetState(true).
			SetTenantID(tenantID),
		tx.Configuration.Create().
			SetName("用户初始密码").
			SetKey("system.user.init_password").
			SetValue("123456").
			SetCategory("system").
			SetRemark("新用户的初始密码").
			SetSort(2).
			SetState(true).
			SetTenantID(tenantID),
		tx.Configuration.Create().
			SetName("会话超时时间").
			SetKey("system.session.timeout").
			SetValue("30").
			SetCategory("system").
			SetRemark("用户会话超时时间(分钟)").
			SetSort(3).
			SetState(true).
			SetTenantID(tenantID),
		tx.Configuration.Create().
			SetName("密码复杂度检查").
			SetKey("system.password.complexity_check").
			SetValue("true").
			SetCategory("security").
			SetRemark("是否启用密码复杂度检查").
			SetSort(4).
			SetState(true).
			SetTenantID(tenantID),
		tx.Configuration.Create().
			SetName("登录失败锁定").
			SetKey("system.login.lock_enabled").
			SetValue("true").
			SetCategory("security").
			SetRemark("是否启用登录失败锁定").
			SetSort(5).
			SetState(true).
			SetTenantID(tenantID),
		tx.Configuration.Create().
			SetName("最大登录失败次数").
			SetKey("system.login.max_retry_count").
			SetValue("5").
			SetCategory("security").
			SetRemark("最大登录失败次数").
			SetSort(6).
			SetState(true).
			SetTenantID(tenantID),
	}

	if err := tx.Configuration.CreateBulk(configs...).Exec(ctxWithTenant); err != nil {
		return dberrorhandler.DefaultEntError(p.logger, err, nil)
	}

	return nil
}

// initDepartmentAndPositions 初始化部门和职位
func (p *CoreTenantPlugin) initDepartmentAndPositions(ctx context.Context, tx *ent.Tx, tenantID uint64) (*ent.Department, error) {
	// 创建默认部门
	ctxWithTenant := hooks.SetTenantIDToContext(context.Background(), tenantID)
	dept, err := tx.Department.Create().
		SetName("总公司").
		SetAncestors("").
		SetLeader("").
		SetPhone("").
		SetEmail("").
		SetStatus(1).
		SetSort(1).
		SetParentID(0).
		SetTenantID(tenantID).
		Save(ctxWithTenant)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(p.logger, err, nil)
	}

	// 检查是否已存在职位数据
	existingPosCount, err := tx.Position.Query().
		Where(position.TenantIDEQ(tenantID)).
		Count(ctxWithTenant)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(p.logger, err, nil)
	}
	if existingPosCount == 0 {
		// 创建默认职位
		positions := []*ent.PositionCreate{
			tx.Position.Create().
				SetName("总经理").
				SetCode("general_manager").
				SetRemark("企业最高管理者").
				SetDeptID(dept.ID).
				SetStatus(1).
				SetSort(1).
				SetTenantID(tenantID),
			tx.Position.Create().
				SetName("部门经理").
				SetCode("department_manager").
				SetRemark("部门负责人").
				SetDeptID(dept.ID).
				SetStatus(1).
				SetSort(2).
				SetTenantID(tenantID),
			tx.Position.Create().
				SetName("普通员工").
				SetCode("employee").
				SetRemark("普通工作人员").
				SetDeptID(dept.ID).
				SetStatus(1).
				SetSort(3).
				SetTenantID(tenantID),
		}

		if err := tx.Position.CreateBulk(positions...).Exec(ctxWithTenant); err != nil {
			return nil, dberrorhandler.DefaultEntError(p.logger, err, nil)
		}
	}

	return dept, nil
}

// initAdminRoleAndUser 初始化管理员角色和用户
func (p *CoreTenantPlugin) initAdminRoleAndUser(ctx context.Context, tx *ent.Tx, req *tenant.InitRequest, dept *ent.Department) (*ent.Role, *ent.User, error) {
	// 创建管理员角色
	ctxWithTenant := hooks.SetTenantIDToContext(context.Background(), req.TenantID)
	adminRole, err := tx.Role.Create().
		SetName("超级管理员").
		SetCode("admin").
		SetDefaultRouter("/workspace").
		SetRemark("租户超级管理员角色").
		SetStatus(1).
		SetSort(1).
		// 🔥 Phase 3: data_scope field removed - now managed via sys_casbin_rules
		SetTenantID(req.TenantID).
		Save(ctxWithTenant)
	if err != nil {
		return nil, nil, dberrorhandler.DefaultEntError(p.logger, err, req)
	}

	// ✅ 查询租户菜单（使用tx确保能查到事务中创建的菜单）
	// 🔒 安全修复：添加租户ID过滤，防止并发初始化时查到其他租户的菜单
	menus, err := tx.Menu.Query().
		Where(
			menu.DisabledEQ(false),
			menu.TenantIDEQ(req.TenantID), // 🔒 关键修复：只查询当前租户的菜单
		).
		All(ctxWithTenant)
	if err != nil {
		logx.Errorw("Failed to query tenant menus",
			logx.Field("tenant_id", req.TenantID),
			logx.Field("error", err.Error()))
		return nil, nil, fmt.Errorf("failed to query tenant menus: %w", err)
	}

	// ✅ 为管理员角色分配所有菜单权限
	if len(menus) > 0 {
		menuIDs := make([]uint64, len(menus))
		for i, m := range menus {
			menuIDs[i] = m.ID
		}

		_, err = tx.Role.UpdateOneID(adminRole.ID).
			AddMenuIDs(menuIDs...).
			Save(ctxWithTenant)
		if err != nil {
			return nil, nil, dberrorhandler.DefaultEntError(p.logger, err, req)
		}

		logx.Infow("✅ Assigned all menus to admin role",
			logx.Field("tenant_id", req.TenantID),
			logx.Field("role_id", adminRole.ID),
			logx.Field("role_code", adminRole.Code),
			logx.Field("menu_count", len(menuIDs)))
	} else {
		logx.Infow("⚠️ No menus available for admin role assignment",
			logx.Field("tenant_id", req.TenantID),
			logx.Field("role_id", adminRole.ID))
	}

	// 创建管理员用户
	username := "admin"
	if req.AdminUsername != nil && *req.AdminUsername != "" {
		username = *req.AdminUsername
	}

	password := "123456"
	if req.AdminPassword != nil && *req.AdminPassword != "" {
		password = *req.AdminPassword
	}

	// 获取租户信息以构建默认邮箱
	tenantInfo, err := p.svcCtx.DB.Tenant.Get(hooks.NewSystemContext(ctx), req.TenantID)
	if err != nil {
		return nil, nil, dberrorhandler.DefaultEntError(p.logger, err, req)
	}

	email := "admin@" + tenantInfo.Code + ".com"
	if req.AdminEmail != nil && *req.AdminEmail != "" {
		email = *req.AdminEmail
	}

	encryptedPassword := encrypt.BcryptEncrypt(password)
	userUUID := uuid.Must(uuid.NewV4())

	adminUser, err := tx.User.Create().
		SetID(userUUID).
		SetUsername(username).
		SetPassword(encryptedPassword).
		SetNickname("超级管理员").
		SetDescription("租户超级管理员").
		SetHomePath("/workspace").
		SetEmail(email).
		SetStatus(1).
		SetDepartmentID(dept.ID).
		SetTenantID(req.TenantID).
		Save(ctxWithTenant)
	if err != nil {
		return nil, nil, dberrorhandler.DefaultEntError(p.logger, err, req)
	}

	// 为用户分配管理员角色
	_, err = tx.User.UpdateOneID(adminUser.ID).
		AddRoleIDs(adminRole.ID).
		Save(ctxWithTenant)
	if err != nil {
		return nil, nil, dberrorhandler.DefaultEntError(p.logger, err, req)
	}

	return adminRole, adminUser, nil
}

// initTenantMenus 为租户初始化菜单副本
func (p *CoreTenantPlugin) initTenantMenus(ctx context.Context, tx *ent.Tx, tenantID uint64) error {
	// 检查是否已存在租户菜单数据
	existingCount, err := tx.Menu.Query().
		Where(menu.TenantIDEQ(tenantID)).
		Count(ctx)
	if err != nil {
		return dberrorhandler.DefaultEntError(p.logger, err, nil)
	}
	if existingCount > 0 {
		logx.Infow("Tenant menus already exist, skipping", logx.Field("tenant_id", tenantID))
		return nil
	}

	// 🔒 安全修复：明确从租户ID=0（系统级模板）复制菜单
	// 使用系统上下文获取全局菜单模板
	baseMenus, err := p.svcCtx.DB.Menu.Query().
		Where(
			menu.DisabledEQ(false),
			menu.TenantIDEQ(0), // 🔒 关键修复：明确从系统级模板租户(ID=0)复制
		).
		Order(ent.Asc(menu.FieldParentID), ent.Asc(menu.FieldSort)).
		All(hooks.NewSystemContext(ctx))
	if err != nil || len(baseMenus) == 0 {
		logx.Infow("No base menus found for tenant initialization, using createCompleteMenus",
			logx.Field("tenant_id", tenantID))
		// 如果没有基础菜单，创建完整的菜单结构（包含所有系统菜单和按钮权限）
		return p.createCompleteMenus(ctx, tx, tenantID)
	}

	// 为租户创建菜单副本
	var menuCreates []*ent.MenuCreate
	oldToNewMenuID := make(map[uint64]uint64) // 用于映射旧菜单ID到新菜单ID

	// 第一轮：创建所有菜单，但暂不设置parent_id
	for _, baseMenu := range baseMenus {
		menuCreate := tx.Menu.Create().
			SetMenuLevel(baseMenu.MenuLevel).
			SetMenuType(baseMenu.MenuType).
			SetPath(baseMenu.Path).
			SetName(baseMenu.Name).
			SetRedirect(baseMenu.Redirect).
			SetComponent(baseMenu.Component).
			SetTitle(baseMenu.Title).
			SetIcon(baseMenu.Icon).
			SetPermission(baseMenu.Permission).
			SetSort(baseMenu.Sort).
			SetHideMenu(baseMenu.HideMenu).
			SetIgnoreKeepAlive(baseMenu.IgnoreKeepAlive).
			SetAffix(baseMenu.Affix).
			SetHideTab(baseMenu.HideTab).
			SetHideChildrenInMenu(baseMenu.HideChildrenInMenu).
			SetCarryParam(baseMenu.CarryParam).
			SetHideBreadcrumb(baseMenu.HideBreadcrumb).
			SetFrameSrc(baseMenu.FrameSrc).
			SetRealPath(baseMenu.RealPath).
			SetDisabled(baseMenu.Disabled).
			SetTenantID(tenantID)

		// 暂时设置parent_id为0，稍后更新
		menuCreates = append(menuCreates, menuCreate.SetParentID(0))
	}

	// 批量创建菜单
	createdMenus, err := tx.Menu.CreateBulk(menuCreates...).Save(ctx)
	if err != nil {
		return dberrorhandler.DefaultEntError(p.logger, err, nil)
	}

	// 建立旧ID到新ID的映射
	for i, createdMenu := range createdMenus {
		oldToNewMenuID[baseMenus[i].ID] = createdMenu.ID
	}

	// 第二轮：更新parent_id关系
	for i, baseMenu := range baseMenus {
		if baseMenu.ParentID != 0 {
			if newParentID, exists := oldToNewMenuID[baseMenu.ParentID]; exists {
				_, err := tx.Menu.UpdateOneID(createdMenus[i].ID).
					SetParentID(newParentID).
					Save(ctx)
				if err != nil {
					return dberrorhandler.DefaultEntError(p.logger, err, nil)
				}
			}
		}
	}

	logx.Infow("Created tenant menu copies",
		logx.Field("tenant_id", tenantID),
		logx.Field("menu_count", len(createdMenus)))

	return nil
}

// createBasicMenus 创建基础菜单结构（当没有基础菜单模板时）
func (p *CoreTenantPlugin) createBasicMenus(ctx context.Context, tx *ent.Tx, tenantID uint64) error {
	basicMenus := []struct {
		MenuLevel uint32
		MenuType  uint32
		Path      string
		Name      string
		Component string
		Title     string
		Icon      string
		ParentID  uint64
		Sort      uint32
	}{
		// 父菜单使用绝对路径
		{1, 0, "/system", "System", "Layout", "系统管理", "lucide:computer", 0, 2},
		// 子菜单使用相对路径（不以/开头），避免路由拼接时出现双斜杠
		{2, 1, "user", "SystemUser", "system/user/index", "用户管理", "lucide:circle-user-round", 2, 1},
		{2, 1, "role", "SystemRole", "system/role/index", "角色管理", "lucide:circle-user", 2, 2},
		{2, 1, "dept", "SystemDepartment", "system/dept/index", "部门管理", "lucide:git-branch-plus", 2, 3},
		{2, 1, "menu", "SystemMenu", "system/menu/index", "菜单管理", "lucide:menu", 2, 4},
		{2, 1, "tenant", "SystemTenant", "system/tenant/index", "租户管理", "lucide:building", 2, 5},
	}

	var topMenus []*ent.Menu

	// 创建顶级菜单
	for _, menu := range basicMenus {
		if menu.ParentID == 0 {
			created, err := tx.Menu.Create().
				SetMenuLevel(menu.MenuLevel).
				SetMenuType(menu.MenuType).
				SetPath(menu.Path).
				SetName(menu.Name).
				SetComponent(menu.Component).
				SetTitle(menu.Title).
				SetIcon(menu.Icon).
				SetParentID(0).
				SetSort(menu.Sort).
				SetHideMenu(false).
				SetIgnoreKeepAlive(false).
				SetAffix(false).
				SetHideTab(false).
				SetHideChildrenInMenu(false).
				SetCarryParam(false).
				SetHideBreadcrumb(false).
				SetDisabled(false).
				SetTenantID(tenantID).
				Save(ctx)
			if err != nil {
				return dberrorhandler.DefaultEntError(p.logger, err, nil)
			}
			topMenus = append(topMenus, created)
		}
	}

	// 创建子菜单
	for _, menu := range basicMenus {
		if menu.ParentID != 0 {
			var parentID uint64
			for _, topMenu := range topMenus {
				if uint64(topMenu.Sort) == menu.ParentID {
					parentID = topMenu.ID
					break
				}
			}

			if parentID != 0 {
				_, err := tx.Menu.Create().
					SetMenuLevel(menu.MenuLevel).
					SetMenuType(menu.MenuType).
					SetPath(menu.Path).
					SetName(menu.Name).
					SetComponent(menu.Component).
					SetTitle(menu.Title).
					SetIcon(menu.Icon).
					SetParentID(parentID).
					SetSort(menu.Sort).
					SetHideMenu(false).
					SetIgnoreKeepAlive(false).
					SetAffix(false).
					SetHideTab(false).
					SetHideChildrenInMenu(false).
					SetCarryParam(false).
					SetHideBreadcrumb(false).
					SetDisabled(false).
					SetTenantID(tenantID).
					Save(ctx)
				if err != nil {
					return dberrorhandler.DefaultEntError(p.logger, err, nil)
				}
			}
		}
	}

	logx.Infow("Created basic tenant menus", logx.Field("tenant_id", tenantID))
	return nil
}

// 🔥 Phase 3: 为管理员角色初始化数据权限规则
// initAdminDataPermissions 为管理员角色初始化数据权限规则到sys_casbin_rules
func (p *CoreTenantPlugin) initAdminDataPermissions(ctx context.Context, tx *ent.Tx, adminRole *ent.Role, tenantID uint64) error {
	// 使用SystemContext绕过租户隔离
	systemCtx := hooks.NewSystemContext(ctx)

	// 清理可能存在的旧数据权限规则（ptype='d'）
	_, err := tx.CasbinRule.Delete().
		Where(
			casbinrule.PtypeEQ("d"),
			casbinrule.V0EQ(adminRole.Code),
			casbinrule.TenantIDEQ(tenantID),
		).
		Exec(systemCtx)
	if err != nil {
		logx.Errorw("Failed to delete old data permission rules",
			logx.Field("role_code", adminRole.Code),
			logx.Field("tenant_id", tenantID),
			logx.Field("error", err.Error()))
		return fmt.Errorf("failed to delete old data permission rules: %w", err)
	}

	// 为租户管理员创建全部数据权限（dataScope="all"）
	// ptype=d: 数据权限规则
	// v0: 角色代码
	// v1: 租户ID（domain）
	// v2: 资源类型（* 表示所有资源）
	// v3: 数据权限范围（all, custom_dept, own_dept_and_sub, own_dept, own）
	// v4: 自定义部门ID列表（JSON数组，仅custom_dept时使用）
	_, err = tx.CasbinRule.Create().
		SetPtype("d").                                      // 数据权限规则类型
		SetV0(adminRole.Code).                              // subject: 角色代码
		SetV1(fmt.Sprintf("%d", tenantID)).                 // domain: 租户ID
		SetV2("*").                                         // object: 资源类型（* 表示所有）
		SetV3("*").                                         // action: 数据权限范围（全部数据）
		SetV4("").                                          // effect: 自定义部门列表（all权限不需要）
		SetServiceName("core").                             // 服务名称
		SetRuleName(fmt.Sprintf("%s数据权限", adminRole.Name)). // 规则名称
		SetDescription(fmt.Sprintf("角色%s的默认数据权限规则，数据范围：all（全部数据）", adminRole.Name)).
		SetCategory("data_permission"). // 规则分类
		SetVersion("1.0.0").
		SetRequireApproval(false).
		SetApprovalStatus("approved").
		SetStatus(1).
		SetTenantID(tenantID).
		Save(systemCtx)

	if err != nil {
		logx.Errorw("Failed to create data permission rule",
			logx.Field("role_code", adminRole.Code),
			logx.Field("role_name", adminRole.Name),
			logx.Field("tenant_id", tenantID),
			logx.Field("error", err.Error()))
		return fmt.Errorf("failed to create data permission rule: %w", err)
	}

	logx.Infow("✅ Successfully created data permission rule for admin role",
		logx.Field("tenant_id", tenantID),
		logx.Field("role_id", adminRole.ID),
		logx.Field("role_code", adminRole.Code),
		logx.Field("role_name", adminRole.Name),
		logx.Field("data_scope", "all"))

	// ⚠️ 不在租户初始化时发布Redis通知
	// 原因：
	// 1. 新租户刚创建，其他服务实例还没有加载该租户的Casbin enforcer
	// 2. 首次访问时会自动从数据库加载策略
	// 3. 避免EntAdapter BatchAdapter接口兼容性问题（API服务的DefaultUpdateCallback会调用SelfAddPolicies）
	//
	// 如果需要通知其他服务，应在租户初始化完成后、首次访问前手动触发策略重新加载

	return nil
}

// 🔥 Phase 3: 为管理员角色初始化API权限规则
// initAdminAPIPermissions 直接使用ent创建API权限规则到sys_casbin_rules表
// ⚠️ 不再使用Casbin enforcer（它会连接到casbin_rules表），而是直接操作sys_casbin_rules
func (p *CoreTenantPlugin) initAdminAPIPermissions(ctx context.Context, adminRole *ent.Role, tenantID uint64) error {
	// 使用系统上下文查询所有API（因为API表是系统级数据，不隔离租户）
	systemCtx := hooks.NewSystemContext(ctx)

	apis, err := p.svcCtx.DB.API.Query().All(systemCtx)
	if err != nil {
		return dberrorhandler.DefaultEntError(p.logger, err, nil)
	}

	if len(apis) == 0 {
		logx.Infow("No APIs found in system, skipping API permission initialization",
			logx.Field("tenant_id", tenantID),
			logx.Field("role_id", adminRole.ID))
		return nil
	}

	logx.Infow("Initializing API permissions for admin role",
		logx.Field("tenant_id", tenantID),
		logx.Field("role_id", adminRole.ID),
		logx.Field("role_code", adminRole.Code),
		logx.Field("api_count", len(apis)))

	// 清理该角色在该租户下的旧API权限规则（ptype='p'）
	tenantIDStr := fmt.Sprintf("%d", tenantID)
	deletedCount, err := p.svcCtx.DB.CasbinRule.Delete().
		Where(
			casbinrule.PtypeEQ("p"),
			casbinrule.V0EQ(adminRole.Code),
			casbinrule.V1EQ(tenantIDStr),
			casbinrule.TenantIDEQ(tenantID),
		).
		Exec(systemCtx)
	if err != nil {
		logx.Errorw("Failed to delete old API permission rules",
			logx.Field("role_code", adminRole.Code),
			logx.Field("tenant_id", tenantID),
			logx.Field("error", err.Error()))
		return fmt.Errorf("failed to delete old API permission rules: %w", err)
	}
	if deletedCount > 0 {
		logx.Infow("Removed old API permission rules",
			logx.Field("role_code", adminRole.Code),
			logx.Field("tenant_id", tenantID),
			logx.Field("count", deletedCount))
	}

	// 批量创建API权限规则
	// RBAC with Domains 格式:
	// ptype=p: API权限规则
	// v0: 角色代码 (subject)
	// v1: 租户ID (domain)
	// v2: API路径 (object)
	// v3: HTTP方法 (action)
	// v4: 效果 (effect: allow/deny)
	var apiRuleCreates []*ent.CasbinRuleCreate
	for _, api := range apis {
		ruleCreate := p.svcCtx.DB.CasbinRule.Create().
			SetPtype("p").          // API权限规则类型
			SetV0(adminRole.Code).  // subject: 角色代码
			SetV1(tenantIDStr).     // domain: 租户ID
			SetV2(api.Path).        // object: API路径
			SetV3(api.Method).      // action: HTTP方法
			SetV4("allow").         // effect: 允许访问
			SetServiceName("core"). // 服务名称
			SetRuleName(fmt.Sprintf("%s-%s权限", adminRole.Name, api.Path)).
			SetDescription(fmt.Sprintf("角色%s访问%s %s的权限", adminRole.Name, api.Method, api.Path)).
			SetCategory("api_permission"). // 规则分类
			SetVersion("1.0.0").
			SetRequireApproval(false).
			SetApprovalStatus("approved").
			SetStatus(1).
			SetTenantID(tenantID)

		apiRuleCreates = append(apiRuleCreates, ruleCreate)
	}

	// 批量执行创建
	err = p.svcCtx.DB.CasbinRule.CreateBulk(apiRuleCreates...).
		Exec(systemCtx)
	if err != nil {
		logx.Errorw("Failed to create API permission rules",
			logx.Field("role_code", adminRole.Code),
			logx.Field("tenant_id", tenantID),
			logx.Field("policy_count", len(apiRuleCreates)),
			logx.Field("error", err.Error()))
		return fmt.Errorf("failed to create API permission rules: %w", err)
	}

	logx.Infow("✅ Successfully initialized API permissions for admin role",
		logx.Field("tenant_id", tenantID),
		logx.Field("role_id", adminRole.ID),
		logx.Field("role_code", adminRole.Code),
		logx.Field("policy_count", len(apiRuleCreates)))

	// 🔔 Redis 通知在初始化完成钩子中统一发布，避免事务阶段发送导致的竞态

	return nil
}
