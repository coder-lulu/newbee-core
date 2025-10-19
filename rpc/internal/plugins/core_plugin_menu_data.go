package plugins

import (
	"context"

	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/zeromicro/go-zero/core/logx"
)

// createCompleteMenus 创建完整的租户菜单结构
// 参考 init_database_menu_data.go，包含所有系统菜单、按钮权限等
// 采用分阶段创建,正确处理ParentID引用
func (p *CoreTenantPlugin) createCompleteMenus(ctx context.Context, tx *ent.Tx, tenantID uint64) error {
	// ========== 第一阶段: 创建顶级父菜单 ==========
	systemParent, err := tx.Menu.Create().
		SetMenuLevel(1).
		SetMenuType(0).
		SetParentID(0).
		SetPath("/system").
		SetName("System").
		SetComponent("Layout").
		SetSort(999).
		SetTitle("系统管理").
		SetIcon("eos-icons:system-group").
		SetHideMenu(false).
		SetServiceName("Core").
		SetTenantID(tenantID).
		Save(ctx)
	if err != nil {
		logx.Errorw("Failed to create system parent menu",
			logx.Field("tenant_id", tenantID),
			logx.Field("error", err.Error()))
		return dberrorhandler.DefaultEntError(p.logger, err, nil)
	}

	logx.Infow("✅ Created system parent menu",
		logx.Field("tenant_id", tenantID),
		logx.Field("menu_id", systemParent.ID))

	// ========== 第二阶段: 创建所有功能菜单 ==========

	// 用户管理
	userMenu, err := tx.Menu.Create().
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(systemParent.ID).
		SetPath("user").
		SetName("UserManagement").
		SetComponent("system/user/index").
		SetSort(1).
		SetTitle("用户管理").
		SetIcon("ph:user-duotone").
		SetHideMenu(false).
		SetPermission("system:user:list").
		SetServiceName("Core").
		SetTenantID(tenantID).
		Save(ctx)
	if err != nil {
		return dberrorhandler.DefaultEntError(p.logger, err, nil)
	}

	// 角色管理
	roleMenu, err := tx.Menu.Create().
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(systemParent.ID).
		SetPath("role").
		SetName("RoleManagement").
		SetComponent("system/role/index").
		SetSort(2).
		SetTitle("角色管理").
		SetIcon("eos-icons:role-binding-outlined").
		SetHideMenu(false).
		SetPermission("system:role:list").
		SetServiceName("Core").
		SetTenantID(tenantID).
		Save(ctx)
	if err != nil {
		return dberrorhandler.DefaultEntError(p.logger, err, nil)
	}

	// 菜单管理
	menuMenu, err := tx.Menu.Create().
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(systemParent.ID).
		SetPath("menu").
		SetName("MenuManagement").
		SetComponent("system/menu/index").
		SetSort(3).
		SetTitle("菜单管理").
		SetIcon("ic:sharp-menu").
		SetHideMenu(false).
		SetPermission("system:menu:list").
		SetServiceName("Core").
		SetTenantID(tenantID).
		Save(ctx)
	if err != nil {
		return dberrorhandler.DefaultEntError(p.logger, err, nil)
	}

	// 部门管理
	deptMenu, err := tx.Menu.Create().
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(systemParent.ID).
		SetPath("dept").
		SetName("DeptManagement").
		SetComponent("system/dept/index").
		SetSort(4).
		SetTitle("部门管理").
		SetIcon("ic:outline-people-alt").
		SetHideMenu(false).
		SetPermission("system:dept:list").
		SetServiceName("Core").
		SetTenantID(tenantID).
		Save(ctx)
	if err != nil {
		return dberrorhandler.DefaultEntError(p.logger, err, nil)
	}

	// 岗位管理
	postMenu, err := tx.Menu.Create().
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(systemParent.ID).
		SetPath("post").
		SetName("PostManagement").
		SetComponent("system/post/index").
		SetSort(5).
		SetTitle("岗位管理").
		SetIcon("icon-park-outline:appointment").
		SetHideMenu(false).
		SetPermission("system:post:list").
		SetServiceName("Core").
		SetTenantID(tenantID).
		Save(ctx)
	if err != nil {
		return dberrorhandler.DefaultEntError(p.logger, err, nil)
	}

	// 字典管理
	dictMenu, err := tx.Menu.Create().
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(systemParent.ID).
		SetPath("dict").
		SetName("DictManagement").
		SetComponent("system/dict/index").
		SetSort(6).
		SetHideChildrenInMenu(true).
		SetTitle("字典管理").
		SetIcon("fluent-mdl2:dictionary").
		SetHideMenu(false).
		SetPermission("system:dict:list").
		SetServiceName("Core").
		SetTenantID(tenantID).
		Save(ctx)
	if err != nil {
		return dberrorhandler.DefaultEntError(p.logger, err, nil)
	}

	// 参数设置
	configMenu, err := tx.Menu.Create().
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(systemParent.ID).
		SetPath("config").
		SetName("ConfigManagement").
		SetComponent("system/config/index").
		SetSort(7).
		SetTitle("参数设置").
		SetIcon("icon-park-twotone:setting-two").
		SetHideMenu(false).
		SetPermission("system:config:list").
		SetServiceName("Core").
		SetTenantID(tenantID).
		Save(ctx)
	if err != nil {
		return dberrorhandler.DefaultEntError(p.logger, err, nil)
	}

	// 接口管理
	apiMenu, err := tx.Menu.Create().
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(systemParent.ID).
		SetPath("api").
		SetName("ApiManagement").
		SetComponent("system/api-interface/index").
		SetSort(8).
		SetTitle("接口管理").
		SetIcon("tabler:api").
		SetHideMenu(false).
		SetPermission("system:api:list").
		SetServiceName("Core").
		SetTenantID(tenantID).
		Save(ctx)
	if err != nil {
		return dberrorhandler.DefaultEntError(p.logger, err, nil)
	}

	// 三方登录
	oauthMenu, err := tx.Menu.Create().
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(systemParent.ID).
		SetPath("oauth").
		SetName("OauthManagement").
		SetComponent("system/oauth/index").
		SetSort(9).
		SetTitle("三方登录").
		SetIcon("tabler:brand-oauth").
		SetHideMenu(false).
		SetPermission("system:oauth:list").
		SetServiceName("Core").
		SetTenantID(tenantID).
		Save(ctx)
	if err != nil {
		return dberrorhandler.DefaultEntError(p.logger, err, nil)
	}

	// 令牌管理
	tokenMenu, err := tx.Menu.Create().
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(systemParent.ID).
		SetPath("token").
		SetName("TokenManagement").
		SetComponent("system/token/index").
		SetSort(10).
		SetTitle("令牌管理").
		SetIcon("tabler:hexagon-letter-j").
		SetHideMenu(false).
		SetPermission("system:token:list").
		SetServiceName("Core").
		SetTenantID(tenantID).
		Save(ctx)
	if err != nil {
		return dberrorhandler.DefaultEntError(p.logger, err, nil)
	}

	// 租户管理
	tenantMenu, err := tx.Menu.Create().
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(systemParent.ID).
		SetPath("tenant").
		SetName("TenantManagement").
		SetComponent("system/tenant/index").
		SetSort(11).
		SetTitle("租户管理").
		SetIcon("ant-design:team-outlined").
		SetHideMenu(false).
		SetPermission("system:tenant:list").
		SetServiceName("Core").
		SetTenantID(tenantID).
		Save(ctx)
	if err != nil {
		return dberrorhandler.DefaultEntError(p.logger, err, nil)
	}

	// 审计管理父菜单（二级目录）
	auditParent, err := tx.Menu.Create().
		SetMenuLevel(2).
		SetMenuType(0).
		SetParentID(systemParent.ID).
		SetPath("audit").
		SetName("AuditManagement").
		SetComponent("ParentView").
		SetSort(100).
		SetTitle("审计管理").
		SetIcon("ant-design:audit-outlined").
		SetHideMenu(false).
		SetServiceName("Core").
		SetTenantID(tenantID).
		Save(ctx)
	if err != nil {
		return dberrorhandler.DefaultEntError(p.logger, err, nil)
	}

	// 审计日志
	auditLogMenu, err := tx.Menu.Create().
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(auditParent.ID).
		SetPath("audit-log").
		SetName("AuditLog").
		SetComponent("system/audit-log/index").
		SetSort(1).
		SetTitle("审计日志").
		SetIcon("ant-design:file-text-outlined").
		SetHideMenu(false).
		SetPermission("sys:audit-log:list").
		SetServiceName("Core").
		SetTenantID(tenantID).
		Save(ctx)
	if err != nil {
		return dberrorhandler.DefaultEntError(p.logger, err, nil)
	}

	// 审计统计（不需要按钮权限,所以忽略返回值）
	_, err = tx.Menu.Create().
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(auditParent.ID).
		SetPath("stats").
		SetName("AuditStats").
		SetComponent("system/audit-log/stats").
		SetSort(2).
		SetTitle("审计统计").
		SetIcon("ant-design:bar-chart-outlined").
		SetHideMenu(false).
		SetPermission("sys:audit-log:stats").
		SetServiceName("Core").
		SetTenantID(tenantID).
		Save(ctx)
	if err != nil {
		return dberrorhandler.DefaultEntError(p.logger, err, nil)
	}

	logx.Infow("✅ Created all feature menus",
		logx.Field("tenant_id", tenantID),
		logx.Field("feature_menu_count", 14))

	// ========== 第三阶段: 创建所有按钮权限 ==========
	var buttonPermissions []*ent.MenuCreate

	// 用户管理按钮权限
	userPermissions := []struct {
		path       string
		name       string
		title      string
		permission string
		sort       uint32
	}{
		{"/system/user/query", "QueryUser", "查询用户", "system:user:query", 1},
		{"/system/user/create", "CreateUser", "创建用户", "system:user:create", 2},
		{"/system/user/update", "UpdateUser", "修改用户", "system:user:update", 3},
		{"/system/user/delete", "DeleteUser", "删除用户", "system:user:delete", 4},
		{"/system/user/export", "ExportUser", "导出用户", "system:user:export", 5},
		{"/system/user/import", "ImportUser", "导入用户", "system:user:import", 6},
		{"/system/user/resetPwd", "ResetUserPassword", "重置用户密码", "system:user:resetPwd", 7},
	}

	for _, perm := range userPermissions {
		buttonPermissions = append(buttonPermissions, tx.Menu.Create().
			SetMenuLevel(3).
			SetMenuType(2).
			SetParentID(userMenu.ID).
			SetPath(perm.path).
			SetName(perm.name).
			SetSort(perm.sort).
			SetTitle(perm.title).
			SetIcon("#").
			SetHideMenu(true).
			SetPermission(perm.permission).
			SetServiceName("Core").
			SetTenantID(tenantID))
	}

	// 角色管理按钮权限
	rolePermissions := []struct {
		path       string
		name       string
		title      string
		permission string
		sort       uint32
	}{
		{"/system/role/query", "QueryRole", "查询角色", "system:role:query", 1},
		{"/system/role/create", "CreateRole", "创建角色", "system:role:create", 2},
		{"/system/role/update", "UpdateRole", "更新角色", "system:role:update", 3},
		{"/system/role/delete", "DeleteRole", "删除角色", "system:role:delete", 4},
		{"/system/role/export", "ExportRole", "导出角色", "system:role:export", 5},
		{"/system/role/changeDataScope", "ChangeRoleDataScope", "修改角色数据权限", "system:role:changeDataScope", 6},
		{"/system/role/changeApiAuth", "ChangeRoleApiAuth", "修改角色接口权限", "system:role:changeApiAuth", 7},
		{"/system/role/assignUser", "AssignRoleUser", "分配角色", "system:role:assignUser", 8},
	}

	for _, perm := range rolePermissions {
		buttonPermissions = append(buttonPermissions, tx.Menu.Create().
			SetMenuLevel(3).
			SetMenuType(2).
			SetParentID(roleMenu.ID).
			SetPath(perm.path).
			SetName(perm.name).
			SetSort(perm.sort).
			SetTitle(perm.title).
			SetIcon("#").
			SetHideMenu(true).
			SetPermission(perm.permission).
			SetServiceName("Core").
			SetTenantID(tenantID))
	}

	// 菜单管理按钮权限
	menuPermissions := []struct {
		path       string
		name       string
		title      string
		permission string
		sort       uint32
	}{
		{"/system/menu/query", "QueryMenu", "查询菜单", "system:menu:query", 1},
		{"/system/menu/create", "CreateMenu", "创建菜单", "system:menu:create", 2},
		{"/system/menu/update", "UpdateMenu", "更新菜单", "system:menu:update", 3},
		{"/system/menu/delete", "DeleteMenu", "删除菜单", "system:menu:delete", 4},
	}

	for _, perm := range menuPermissions {
		buttonPermissions = append(buttonPermissions, tx.Menu.Create().
			SetMenuLevel(3).
			SetMenuType(2).
			SetParentID(menuMenu.ID).
			SetPath(perm.path).
			SetName(perm.name).
			SetSort(perm.sort).
			SetTitle(perm.title).
			SetIcon("#").
			SetHideMenu(true).
			SetPermission(perm.permission).
			SetServiceName("Core").
			SetTenantID(tenantID))
	}

	// 部门管理按钮权限
	deptPermissions := []struct {
		path       string
		name       string
		title      string
		permission string
		sort       uint32
	}{
		{"/system/dept/query", "QueryDept", "查询部门", "system:dept:query", 1},
		{"/system/dept/create", "CreateDept", "创建部门", "system:dept:create", 2},
		{"/system/dept/update", "UpdateDept", "更新部门", "system:dept:update", 3},
		{"/system/dept/delete", "DeleteDept", "删除部门", "system:dept:delete", 4},
	}

	for _, perm := range deptPermissions {
		buttonPermissions = append(buttonPermissions, tx.Menu.Create().
			SetMenuLevel(3).
			SetMenuType(2).
			SetParentID(deptMenu.ID).
			SetPath(perm.path).
			SetName(perm.name).
			SetSort(perm.sort).
			SetTitle(perm.title).
			SetIcon("#").
			SetHideMenu(true).
			SetPermission(perm.permission).
			SetServiceName("Core").
			SetTenantID(tenantID))
	}

	// 岗位管理按钮权限
	postPermissions := []struct {
		path       string
		name       string
		title      string
		permission string
		sort       uint32
	}{
		{"/system/post/query", "QueryPost", "查询岗位", "system:post:query", 1},
		{"/system/post/create", "CreatePost", "创建岗位", "system:post:create", 2},
		{"/system/post/update", "UpdatePost", "更新岗位", "system:post:update", 3},
		{"/system/post/delete", "DeletePost", "删除岗位", "system:post:delete", 4},
		{"/system/post/export", "ExportPost", "导出岗位", "system:post:export", 5},
	}

	for _, perm := range postPermissions {
		buttonPermissions = append(buttonPermissions, tx.Menu.Create().
			SetMenuLevel(3).
			SetMenuType(2).
			SetParentID(postMenu.ID).
			SetPath(perm.path).
			SetName(perm.name).
			SetSort(perm.sort).
			SetTitle(perm.title).
			SetIcon("#").
			SetHideMenu(true).
			SetPermission(perm.permission).
			SetServiceName("Core").
			SetTenantID(tenantID))
	}

	// 字典管理按钮权限
	dictPermissions := []struct {
		path       string
		name       string
		title      string
		permission string
		sort       uint32
	}{
		{"/system/dict/query", "QueryDict", "查询字典", "system:dict:query", 1},
		{"/system/dict/create", "CreateDict", "创建字典", "system:dict:create", 2},
		{"/system/dict/update", "UpdateDict", "更新字典", "system:dict:update", 3},
		{"/system/dict/delete", "DeleteDict", "删除字典", "system:dict:delete", 4},
		{"/system/dict/export", "ExportDict", "导出字典", "system:dict:export", 5},
	}

	for _, perm := range dictPermissions {
		buttonPermissions = append(buttonPermissions, tx.Menu.Create().
			SetMenuLevel(3).
			SetMenuType(2).
			SetParentID(dictMenu.ID).
			SetPath(perm.path).
			SetName(perm.name).
			SetSort(perm.sort).
			SetTitle(perm.title).
			SetIcon("#").
			SetHideMenu(true).
			SetPermission(perm.permission).
			SetServiceName("Core").
			SetTenantID(tenantID))
	}

	// 参数设置按钮权限
	configPermissions := []struct {
		path       string
		name       string
		title      string
		permission string
		sort       uint32
	}{
		{"/system/config/query", "QueryConfig", "查询参数", "system:config:query", 1},
		{"/system/config/create", "CreateConfig", "创建参数", "system:config:create", 2},
		{"/system/config/update", "UpdateConfig", "更新参数", "system:config:update", 3},
		{"/system/config/delete", "DeleteConfig", "删除参数", "system:config:delete", 4},
		{"/system/config/export", "ExportConfig", "导出参数", "system:config:export", 5},
	}

	for _, perm := range configPermissions {
		buttonPermissions = append(buttonPermissions, tx.Menu.Create().
			SetMenuLevel(3).
			SetMenuType(2).
			SetParentID(configMenu.ID).
			SetPath(perm.path).
			SetName(perm.name).
			SetSort(perm.sort).
			SetTitle(perm.title).
			SetIcon("#").
			SetHideMenu(true).
			SetPermission(perm.permission).
			SetServiceName("Core").
			SetTenantID(tenantID))
	}

	// 接口管理按钮权限
	apiPermissions := []struct {
		path       string
		name       string
		title      string
		permission string
		sort       uint32
	}{
		{"/system/api/query", "QueryApi", "查询接口", "system:api:query", 1},
		{"/system/api/create", "CreateApi", "创建接口", "system:api:create", 2},
		{"/system/api/update", "UpdateApi", "更新接口", "system:api:update", 3},
		{"/system/api/delete", "DeleteApi", "删除接口", "system:api:delete", 4},
		{"/system/api/export", "ExportApi", "导出接口", "system:api:export", 5},
	}

	for _, perm := range apiPermissions {
		buttonPermissions = append(buttonPermissions, tx.Menu.Create().
			SetMenuLevel(3).
			SetMenuType(2).
			SetParentID(apiMenu.ID).
			SetPath(perm.path).
			SetName(perm.name).
			SetSort(perm.sort).
			SetTitle(perm.title).
			SetIcon("#").
			SetHideMenu(true).
			SetPermission(perm.permission).
			SetServiceName("Core").
			SetTenantID(tenantID))
	}

	// 三方登录按钮权限
	oauthPermissions := []struct {
		path       string
		name       string
		title      string
		permission string
		sort       uint32
	}{
		{"/system/oauth/query", "QueryOauth", "查询三方登录", "system:oauth:query", 1},
		{"/system/oauth/create", "CreateOauth", "创建三方登录", "system:oauth:create", 2},
		{"/system/oauth/update", "UpdateOauth", "更新三方登录", "system:oauth:update", 3},
		{"/system/oauth/delete", "DeleteOauth", "删除三方登录", "system:oauth:delete", 4},
	}

	for _, perm := range oauthPermissions {
		buttonPermissions = append(buttonPermissions, tx.Menu.Create().
			SetMenuLevel(3).
			SetMenuType(2).
			SetParentID(oauthMenu.ID).
			SetPath(perm.path).
			SetName(perm.name).
			SetSort(perm.sort).
			SetTitle(perm.title).
			SetIcon("#").
			SetHideMenu(true).
			SetPermission(perm.permission).
			SetServiceName("Core").
			SetTenantID(tenantID))
	}

	// 令牌管理按钮权限
	tokenPermissions := []struct {
		path       string
		name       string
		title      string
		permission string
		sort       uint32
	}{
		{"/system/token/query", "QueryToken", "查询令牌", "system:token:query", 1},
		{"/system/token/delete", "DeleteToken", "删除令牌", "system:token:delete", 2},
	}

	for _, perm := range tokenPermissions {
		buttonPermissions = append(buttonPermissions, tx.Menu.Create().
			SetMenuLevel(3).
			SetMenuType(2).
			SetParentID(tokenMenu.ID).
			SetPath(perm.path).
			SetName(perm.name).
			SetSort(perm.sort).
			SetTitle(perm.title).
			SetIcon("#").
			SetHideMenu(true).
			SetPermission(perm.permission).
			SetServiceName("Core").
			SetTenantID(tenantID))
	}

	// 租户管理按钮权限
	tenantPermissions := []struct {
		path       string
		name       string
		title      string
		permission string
		sort       uint32
	}{
		{"/system/tenant/query", "QueryTenant", "查询租户", "system:tenant:query", 1},
		{"/system/tenant/add", "AddTenant", "创建租户", "system:tenant:add", 2},
		{"/system/tenant/edit", "EditTenant", "编辑租户", "system:tenant:edit", 3},
		{"/system/tenant/remove", "RemoveTenant", "删除租户", "system:tenant:remove", 4},
		{"/system/tenant/export", "ExportTenant", "导出租户", "system:tenant:export", 5},
		{"/system/tenant/init", "InitTenant", "初始化租户", "system:tenant:init", 6},
	}

	for _, perm := range tenantPermissions {
		buttonPermissions = append(buttonPermissions, tx.Menu.Create().
			SetMenuLevel(3).
			SetMenuType(2).
			SetParentID(tenantMenu.ID).
			SetPath(perm.path).
			SetName(perm.name).
			SetSort(perm.sort).
			SetTitle(perm.title).
			SetIcon("#").
			SetHideMenu(true).
			SetPermission(perm.permission).
			SetServiceName("Core").
			SetTenantID(tenantID))
	}

	// 审计日志按钮权限
	auditPermissions := []struct {
		path       string
		name       string
		title      string
		permission string
		sort       uint32
	}{
		{"/system/audit-log/detail", "AuditLogDetail", "查看审计日志详情", "sys:audit-log:detail", 1},
		{"/system/audit-log/delete", "DeleteAuditLog", "删除审计日志", "sys:audit-log:delete", 2},
		{"/system/audit-log/export", "ExportAuditLog", "导出审计日志", "sys:audit-log:export", 3},
	}

	for _, perm := range auditPermissions {
		buttonPermissions = append(buttonPermissions, tx.Menu.Create().
			SetMenuLevel(3).
			SetMenuType(2).
			SetParentID(auditLogMenu.ID).
			SetPath(perm.path).
			SetName(perm.name).
			SetSort(perm.sort).
			SetTitle(perm.title).
			SetIcon("#").
			SetHideMenu(true).
			SetPermission(perm.permission).
			SetServiceName("Core").
			SetTenantID(tenantID))
	}

	// 批量创建所有按钮权限
	err = tx.Menu.CreateBulk(buttonPermissions...).Exec(ctx)
	if err != nil {
		logx.Errorw("Failed to create button permissions",
			logx.Field("tenant_id", tenantID),
			logx.Field("button_count", len(buttonPermissions)),
			logx.Field("error", err.Error()))
		return dberrorhandler.DefaultEntError(p.logger, err, nil)
	}

	logx.Infow("✅ Successfully created complete tenant menus",
		logx.Field("tenant_id", tenantID),
		logx.Field("feature_menus", 14),
		logx.Field("button_permissions", len(buttonPermissions)),
		logx.Field("total_menus", 14+len(buttonPermissions)))

	return nil
}
