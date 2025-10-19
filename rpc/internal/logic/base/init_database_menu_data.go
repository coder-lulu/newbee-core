package base

import (
	"context"

	"github.com/coder-lulu/newbee-common/v2/enum/common"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/hooks"
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/coder-lulu/newbee-core/rpc/ent"
)

// insert initial menu data
func (l *InitDatabaseLogic) insertMenuData(ctx context.Context) error {
	ctxWithTenant := hooks.SetTenantIDToContext(context.Background(), 1)
	var menus []*ent.MenuCreate

	// 系统管理 ID 1
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(1).
		SetMenuLevel(1).
		SetMenuType(0).
		SetParentID(common.DefaultParentId).
		SetPath("/system").
		SetName("系统管理").
		SetComponent("Layout").
		SetSort(999).
		SetTitle("系统管理").
		SetIcon("eos-icons:system-group").
		SetHideMenu(false).
		SetParams("").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 用户管理 ID

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(2).
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(1).
		SetPath("user").
		SetName("用户管理").
		SetComponent("system/user/index").
		SetSort(1).
		SetTitle("用户管理").
		SetIcon("ph:user-duotone").
		SetHideMenu(false).
		SetParams("").
		SetPermission("system:user:list").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 查询用户 ID 3
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(3).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(2).
		SetPath("/system/user/query").
		SetName("查询用户").
		SetSort(1).
		SetTitle("查询用户").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:user:query").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 创建用户 ID 4
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(4).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(2).
		SetPath("/system/user/create").
		SetName("创建用户").
		SetSort(2).
		SetTitle("创建用户").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:user:create").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 修改用户 ID 5
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(5).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(2).
		SetPath("/system/user/update").
		SetName("修改用户").
		SetSort(3).
		SetTitle("修改用户").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:user:update").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 删除用户 ID 6
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(6).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(2).
		SetPath("/system/user/delete").
		SetName("删除用户").
		SetSort(4).
		SetTitle("删除用户").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:user:delete").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 导出用户 ID 7
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(7).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(2).
		SetPath("/system/user/export").
		SetName("导出用户").
		SetSort(5).
		SetTitle("导出用户").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:user:export").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 导入用户 ID 8
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(8).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(2).
		SetPath("/system/user/import").
		SetName("导入用户").
		SetSort(6).
		SetTitle("导入用户").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:user:import").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 重置用户密码 ID 9
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(9).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(2).
		SetPath("/system/user/resetPwd").
		SetName("重置用户密码").
		SetSort(7).
		SetTitle("重置用户密码").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:user:resetPwd").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 角色管理  ID 10
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(10).
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(1).
		SetPath("role").
		SetName("角色管理").
		SetComponent("system/role/index").
		SetSort(2).
		SetTitle("角色管理").
		SetIcon("eos-icons:role-binding-outlined").
		SetHideMenu(false).
		SetParams("").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 查询角色 ID 11
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(11).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(10).
		SetPath("/system/role/query").
		SetName("查询角色").
		SetSort(1).
		SetTitle("查询角色").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:role:query").
		SetServiceName("Core").
		SetTenantID(1),
	)
	// 创建角色 ID 12
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(12).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(10).
		SetPath("/system/role/create").
		SetName("创建角色").
		SetSort(2).
		SetTitle("创建角色").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:role:create").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 更新角色 ID 13
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(13).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(10).
		SetPath("/system/role/update").
		SetName("更新角色").
		SetSort(3).
		SetTitle("更新角色").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:role:update").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 删除角色 ID 14
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(14).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(10).
		SetPath("/system/role/delete").
		SetName("删除角色").
		SetSort(4).
		SetTitle("删除角色").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:role:delete").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 导出角色 ID 15
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(15).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(10).
		SetPath("/system/role/export").
		SetName("导出角色").
		SetSort(5).
		SetTitle("导出角色").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:role:export").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 修改角色数据权限 ID 16
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(16).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(10).
		SetPath("/system/role/changeDataScope").
		SetName("修改角色数据权限").
		SetSort(6).
		SetTitle("修改角色数据权限").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:role:changeDataScope").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 修改角色接口权限 ID 17
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(17).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(10).
		SetPath("/system/role/changeApiAuth").
		SetName("修改角色接口权限").
		SetSort(7).
		SetTitle("修改角色接口权限").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:role:changeApiAuth").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 分配角色 ID 18
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(18).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(10).
		SetPath("/system/role/assignUser").
		SetName("分配角色").
		SetSort(8).
		SetTitle("分配角色").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:role:assignUser").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 菜单管理 ID 19
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(19).
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(1).
		SetPath("menu").
		SetName("菜单管理").
		SetComponent("system/menu/index").
		SetSort(3).
		SetTitle("菜单管理").
		SetIcon("ic:sharp-menu").
		SetHideMenu(false).
		SetParams("").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 查询菜单 ID 20
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(20).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(19).
		SetPath("/system/menu/query").
		SetName("查询菜单").
		SetSort(1).
		SetTitle("查询菜单").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:menu:query").
		SetServiceName("Core").
		SetTenantID(1),
	)
	// 创建菜单 ID 21
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(21).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(19).
		SetPath("/system/menu/create").
		SetName("创建菜单").
		SetSort(2).
		SetTitle("创建菜单").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:menu:create").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 更新菜单 ID 22
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(22).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(19).
		SetPath("/system/menu/update").
		SetName("更新菜单").
		SetSort(3).
		SetTitle("更新菜单").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:menu:update").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 删除菜单 ID 23
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(23).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(19).
		SetPath("/system/menu/delete").
		SetName("删除菜单").
		SetSort(4).
		SetTitle("删除菜单").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:menu:delete").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 部门管理 ID 24

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(24).
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(1).
		SetPath("dept").
		SetName("部门管理").
		SetComponent("system/dept/index").
		SetSort(4).
		SetTitle("部门管理").
		SetIcon("ic:outline-people-alt").
		SetHideMenu(false).
		SetParams("").
		SetServiceName("Core").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(25).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(24).
		SetPath("/system/dept/query").
		SetName("查询部门").
		SetSort(1).
		SetTitle("查询部门").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:dept:query").
		SetServiceName("Core").
		SetTenantID(1),
	)
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(26).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(24).
		SetPath("/system/dept/create").
		SetName("创建部门").
		SetSort(2).
		SetTitle("创建部门").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:dept:create").
		SetServiceName("Core").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(27).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(24).
		SetPath("/system/dept/update").
		SetName("更新部门").
		SetSort(3).
		SetTitle("更新部门").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:dept:update").
		SetServiceName("Core").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(28).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(24).
		SetPath("/system/dept/delete").
		SetName("删除部门").
		SetSort(4).
		SetTitle("删除部门").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:dept:delete").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 岗位管理 ID 29
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(29).
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(1).
		SetPath("post").
		SetName("岗位管理").
		SetComponent("system/post/index").
		SetSort(5).
		SetTitle("岗位管理").
		SetIcon("icon-park-outline:appointment").
		SetHideMenu(false).
		SetParams("").
		SetServiceName("Core").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(30).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(29).
		SetPath("/system/post/query").
		SetName("查询岗位").
		SetSort(1).
		SetTitle("查询岗位").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:post:query").
		SetServiceName("Core").
		SetTenantID(1),
	)
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(31).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(29).
		SetPath("/system/post/create").
		SetName("创建岗位").
		SetSort(2).
		SetTitle("创建岗位").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:post:create").
		SetServiceName("Core").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(32).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(29).
		SetPath("/system/post/update").
		SetName("更新岗位").
		SetSort(3).
		SetTitle("更新岗位").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:post:update").
		SetServiceName("Core").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(33).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(29).
		SetPath("/system/post/delete").
		SetName("删除岗位").
		SetSort(4).
		SetTitle("删除岗位").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:post:delete").
		SetServiceName("Core").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(34).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(29).
		SetPath("/system/post/export").
		SetName("导出岗位").
		SetSort(5).
		SetTitle("导出岗位").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:post:export").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 字典管理 ID 35

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(35).
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(1).
		SetPath("dict").
		SetName("字典管理").
		SetComponent("system/dict/index").
		SetSort(6).
		SetHideChildrenInMenu(true).
		SetTitle("字典管理").
		SetIcon("fluent-mdl2:dictionary").
		SetHideMenu(false).
		SetParams("").
		SetServiceName("Core").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(36).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(35).
		SetPath("/system/dict/query").
		SetName("查询字典").
		SetSort(1).
		SetTitle("查询字典").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:dict:query").
		SetServiceName("Core").
		SetTenantID(1),
	)
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(37).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(35).
		SetPath("/system/dict/create").
		SetName("创建字典").
		SetSort(2).
		SetTitle("创建字典").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:dict:create").
		SetServiceName("Core").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(38).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(35).
		SetPath("/system/dict/update").
		SetName("更新字典").
		SetSort(3).
		SetTitle("更新字典").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:dict:update").
		SetServiceName("Core").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(39).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(35).
		SetPath("/system/dict/delete").
		SetName("删除字典").
		SetSort(4).
		SetTitle("删除字典").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:dict:delete").
		SetServiceName("Core").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(40).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(35).
		SetPath("/system/dict/export").
		SetName("导出字典").
		SetSort(5).
		SetTitle("导出字典").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:dict:export").
		SetServiceName("Core").
		SetTenantID(1),
	)

	//参数设置 ID 41

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(41).
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(1).
		SetPath("config").
		SetName("参数设置").
		SetComponent("system/config/index").
		SetSort(7).
		SetTitle("参数设置").
		SetIcon("icon-park-twotone:setting-two").
		SetHideMenu(false).
		SetParams("").
		SetServiceName("Core").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(42).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(41).
		SetPath("/system/config/query").
		SetName("查询参数").
		SetSort(1).
		SetTitle("查询参数").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:config:query").
		SetServiceName("Core").
		SetTenantID(1),
	)
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(43).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(41).
		SetPath("/system/config/create").
		SetName("创建参数").
		SetSort(2).
		SetTitle("创建参数").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:config:create").
		SetServiceName("Core").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(44).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(41).
		SetPath("/system/config/update").
		SetName("更新参数").
		SetSort(3).
		SetTitle("更新参数").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:config:update").
		SetServiceName("Core").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(45).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(41).
		SetPath("/system/config/delete").
		SetName("删除参数").
		SetSort(4).
		SetTitle("删除参数").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:config:delete").
		SetServiceName("Core").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(46).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(41).
		SetPath("/system/config/export").
		SetName("导出参数").
		SetSort(5).
		SetTitle("导出参数").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:config:export").
		SetServiceName("Core").
		SetTenantID(1),
	)

	//接口管理 ID 47

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(47).
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(1).
		SetPath("api").
		SetName("接口管理").
		SetComponent("system/api-interface/index").
		SetSort(8).
		SetTitle("接口管理").
		SetIcon("tabler:api").
		SetHideMenu(false).
		SetParams("").
		SetServiceName("Core").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(48).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(47).
		SetPath("/system/api/query").
		SetName("查询接口").
		SetSort(1).
		SetTitle("查询接口").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:api:query").
		SetServiceName("Core").
		SetTenantID(1),
	)
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(49).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(47).
		SetPath("/system/api/create").
		SetName("创建接口").
		SetSort(2).
		SetTitle("创建接口").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:api:create").
		SetServiceName("Core").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(50).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(47).
		SetPath("/system/api/update").
		SetName("更新接口").
		SetSort(3).
		SetTitle("更新接口").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:api:update").
		SetServiceName("Core").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(51).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(47).
		SetPath("/system/api/delete").
		SetName("删除接口").
		SetSort(4).
		SetTitle("删除接口").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:api:delete").
		SetServiceName("Core").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(52).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(47).
		SetPath("/system/api/export").
		SetName("导出接口").
		SetSort(5).
		SetTitle("导出接口").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:api:export").
		SetServiceName("Core").
		SetTenantID(1),
	)

	//三方登录  ID 53

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(53).
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(1).
		SetPath("oauth").
		SetName("三方登录").
		SetComponent("system/oauth/index").
		SetSort(9).
		SetTitle("三方登录").
		SetIcon("tabler:brand-oauth").
		SetHideMenu(false).
		SetParams("").
		SetServiceName("Core").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(54).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(53).
		SetPath("/system/oauth/query").
		SetName("查询三方登录").
		SetSort(1).
		SetTitle("查询三方登录").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:oauth:query").
		SetServiceName("Core").
		SetTenantID(1),
	)
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(55).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(53).
		SetPath("/system/oauth/create").
		SetName("创建三方登录").
		SetSort(2).
		SetTitle("创建三方登录").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:oauth:create").
		SetServiceName("Core").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(56).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(53).
		SetPath("/system/oauth/update").
		SetName("更新三方登录").
		SetSort(3).
		SetTitle("更新三方登录").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:oauth:update").
		SetServiceName("Core").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(57).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(53).
		SetPath("/system/oauth/delete").
		SetName("删除三方登录").
		SetSort(4).
		SetTitle("删除三方登录").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:oauth:delete").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// token管理 ID 58

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(58).
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(1).
		SetPath("token").
		SetName("令牌管理").
		SetComponent("system/token/index").
		SetSort(10).
		SetTitle("令牌管理").
		SetIcon("tabler:hexagon-letter-j").
		SetHideMenu(false).
		SetParams("").
		SetPermission("system:token:list").
		SetServiceName("Core").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(59).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(58).
		SetPath("/system/token/query").
		SetName("查询令牌").
		SetSort(1).
		SetTitle("查询令牌").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:token:query").
		SetServiceName("Core").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(60).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(58).
		SetPath("/system/token/delete").
		SetName("删除令牌").
		SetSort(2).
		SetTitle("删除令牌").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:token:delete").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 租户管理 ID 61
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(61).
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(1).
		SetPath("tenant").
		SetName("租户管理").
		SetComponent("system/tenant/index").
		SetSort(11).
		SetTitle("租户管理").
		SetIcon("ant-design:team-outlined").
		SetHideMenu(false).
		SetParams("").
		SetPermission("system:tenant:list").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 租户管理权限 ID 60-66
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(62).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(61).
		SetPath("/system/tenant/query").
		SetName("查询租户").
		SetSort(1).
		SetTitle("查询租户").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:tenant:query").
		SetServiceName("Core").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(63).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(61).
		SetPath("/system/tenant/add").
		SetName("创建租户").
		SetSort(2).
		SetTitle("创建租户").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:tenant:add").
		SetServiceName("Core").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(64).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(61).
		SetPath("/system/tenant/edit").
		SetName("编辑租户").
		SetSort(3).
		SetTitle("编辑租户").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:tenant:edit").
		SetServiceName("Core").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(65).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(61).
		SetPath("/system/tenant/remove").
		SetName("删除租户").
		SetSort(4).
		SetTitle("删除租户").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:tenant:remove").
		SetServiceName("Core").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(66).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(61).
		SetPath("/system/tenant/export").
		SetName("导出租户").
		SetSort(5).
		SetTitle("导出租户").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:tenant:export").
		SetServiceName("Core").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(67).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(61).
		SetPath("/system/tenant/init").
		SetName("初始化租户").
		SetSort(6).
		SetTitle("初始化租户").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:tenant:init").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 审计管理 ID 68
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(68).
		SetMenuLevel(2).
		SetMenuType(0).
		SetParentID(1).
		SetPath("audit").
		SetName("审计管理").
		SetComponent("ParentView").
		SetSort(100).
		SetTitle("审计管理").
		SetIcon("ant-design:audit-outlined").
		SetHideMenu(false).
		SetParams("").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 审计日志 ID 69
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(69).
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(68).
		SetPath("audit-log").
		SetName("审计日志").
		SetComponent("system/audit-log/index").
		SetSort(1).
		SetTitle("审计日志").
		SetIcon("ant-design:file-text-outlined").
		SetHideMenu(false).
		SetParams("").
		SetPermission("sys:audit-log:list").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 审计统计 ID 70
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(70).
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(68).
		SetPath("stats").
		SetName("审计统计").
		SetComponent("system/audit-log/stats").
		SetSort(2).
		SetTitle("审计统计").
		SetIcon("ant-design:bar-chart-outlined").
		SetHideMenu(false).
		SetParams("").
		SetPermission("sys:audit-log:stats").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 审计日志详情权限 ID 71
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(71).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(69).
		SetPath("/system/audit-log/detail").
		SetName("查看审计日志详情").
		SetSort(1).
		SetTitle("查看审计日志详情").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("sys:audit-log:detail").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 删除审计日志权限 ID 72
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(72).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(69).
		SetPath("/system/audit-log/delete").
		SetName("删除审计日志").
		SetSort(2).
		SetTitle("删除审计日志").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("sys:audit-log:delete").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 导出审计日志权限 ID 73
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(73).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(69).
		SetPath("/system/audit-log/export").
		SetName("导出审计日志").
		SetSort(3).
		SetTitle("导出审计日志").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("sys:audit-log:export").
		SetServiceName("Core").
		SetTenantID(1),
	)

	// 配置管理 ID 100
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(100).
		SetMenuLevel(1).
		SetMenuType(0).
		SetParentID(common.DefaultParentId).
		SetPath("/cmdb").
		SetName("配置管理").
		SetComponent("Layout").
		SetSort(800).
		SetTitle("配置管理").
		SetIcon("mdi:server-network").
		SetHideMenu(false).
		SetParams("").
		SetServiceName("Cmdb").
		SetTenantID(1),
	)

	// CI实例管理 ID 101
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(101).
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(100).
		SetPath("cis").
		SetName("CI实例管理").
		SetComponent("cmdb/cis/index").
		SetSort(1).
		SetTitle("CI实例管理").
		SetIcon("mdi:server").
		SetHideMenu(false).
		SetParams("").
		SetPermission("cmdb:ci:list").
		SetServiceName("Cmdb").
		SetTenantID(1),
	)

	// CI实例权限菜单 ID 102-108
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(102).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(101).
		SetPath("/cmdb/ci/query").
		SetName("查询CI").
		SetSort(1).
		SetTitle("查询CI").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("cmdb:ci:query").
		SetServiceName("Cmdb").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(103).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(101).
		SetPath("/cmdb/ci/create").
		SetName("创建CI").
		SetSort(2).
		SetTitle("创建CI").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("cmdb:ci:create").
		SetServiceName("Cmdb").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(104).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(101).
		SetPath("/cmdb/ci/update").
		SetName("更新CI").
		SetSort(3).
		SetTitle("更新CI").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("cmdb:ci:update").
		SetServiceName("Cmdb").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(105).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(101).
		SetPath("/cmdb/ci/delete").
		SetName("删除CI").
		SetSort(4).
		SetTitle("删除CI").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("cmdb:ci:delete").
		SetServiceName("Cmdb").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(106).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(101).
		SetPath("/cmdb/ci/export").
		SetName("导出CI").
		SetSort(5).
		SetTitle("导出CI").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("cmdb:ci:export").
		SetServiceName("Cmdb").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(107).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(101).
		SetPath("/cmdb/ci/import").
		SetName("导入CI").
		SetSort(6).
		SetTitle("导入CI").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("cmdb:ci:import").
		SetServiceName("Cmdb").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(108).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(101).
		SetPath("/cmdb/ci/batch").
		SetName("批量操作CI").
		SetSort(7).
		SetTitle("批量操作CI").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("cmdb:ci:batch").
		SetServiceName("Cmdb").
		SetTenantID(1),
	)

	// CI类型管理 ID 110
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(110).
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(100).
		SetPath("ci_types").
		SetName("CI类型管理").
		SetComponent("cmdb/ci_types/index").
		SetSort(2).
		SetTitle("CI类型管理").
		SetIcon("mdi:file-tree").
		SetHideMenu(false).
		SetParams("").
		SetPermission("cmdb:ci_type:list").
		SetServiceName("Cmdb").
		SetTenantID(1),
	)

	// CI类型权限菜单 ID 111-118
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(111).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(110).
		SetPath("/cmdb/ci_type/query").
		SetName("查询CI类型").
		SetSort(1).
		SetTitle("查询CI类型").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("cmdb:ci_type:query").
		SetServiceName("Cmdb").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(112).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(110).
		SetPath("/cmdb/ci_type/create").
		SetName("创建CI类型").
		SetSort(2).
		SetTitle("创建CI类型").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("cmdb:ci_type:create").
		SetServiceName("Cmdb").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(113).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(110).
		SetPath("/cmdb/ci_type/update").
		SetName("更新CI类型").
		SetSort(3).
		SetTitle("更新CI类型").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("cmdb:ci_type:update").
		SetServiceName("Cmdb").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(114).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(110).
		SetPath("/cmdb/ci_type/delete").
		SetName("删除CI类型").
		SetSort(4).
		SetTitle("删除CI类型").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("cmdb:ci_type:delete").
		SetServiceName("Cmdb").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(115).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(110).
		SetPath("/cmdb/ci_type/export").
		SetName("导出CI类型").
		SetSort(5).
		SetTitle("导出CI类型").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("cmdb:ci_type:export").
		SetServiceName("Cmdb").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(116).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(110).
		SetPath("/cmdb/ci_type/attribute").
		SetName("管理属性").
		SetSort(6).
		SetTitle("管理属性").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("cmdb:ci_type:attribute").
		SetServiceName("Cmdb").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(117).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(110).
		SetPath("/cmdb/ci_type/group").
		SetName("管理分组").
		SetSort(7).
		SetTitle("管理分组").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("cmdb:ci_type:group").
		SetServiceName("Cmdb").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(118).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(110).
		SetPath("/cmdb/ci_type/relation").
		SetName("管理关系").
		SetSort(8).
		SetTitle("管理关系").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("cmdb:ci_type:relation").
		SetServiceName("Cmdb").
		SetTenantID(1),
	)

	// CI权限管理 ID 120
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(120).
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(100).
		SetPath("ci_permission").
		SetName("CI权限管理").
		SetComponent("cmdb/ci_permission/index").
		SetSort(3).
		SetTitle("CI权限管理").
		SetIcon("mdi:shield-account").
		SetHideMenu(false).
		SetParams("").
		SetPermission("cmdb:permission:list").
		SetServiceName("Cmdb").
		SetTenantID(1),
	)

	// CI权限管理权限菜单 ID 121-125
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(121).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(120).
		SetPath("/cmdb/ci_permission/query").
		SetName("查询CI权限").
		SetSort(1).
		SetTitle("查询CI权限").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("cmdb:permission:query").
		SetServiceName("Cmdb").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(122).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(120).
		SetPath("/cmdb/ci_permission/create").
		SetName("创建CI权限").
		SetSort(2).
		SetTitle("创建CI权限").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("cmdb:permission:create").
		SetServiceName("Cmdb").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(123).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(120).
		SetPath("/cmdb/ci_permission/update").
		SetName("更新CI权限").
		SetSort(3).
		SetTitle("更新CI权限").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("cmdb:permission:update").
		SetServiceName("Cmdb").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(124).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(120).
		SetPath("/cmdb/ci_permission/delete").
		SetName("删除CI权限").
		SetSort(4).
		SetTitle("删除CI权限").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("cmdb:permission:delete").
		SetServiceName("Cmdb").
		SetTenantID(1),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetID(125).
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(120).
		SetPath("/cmdb/ci_permission/assign").
		SetName("分配权限").
		SetSort(5).
		SetTitle("分配权限").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("cmdb:permission:assign").
		SetServiceName("Cmdb").
		SetTenantID(1),
	)

	err := l.svcCtx.DB.Menu.CreateBulk(menus...).Exec(ctxWithTenant)
	if err != nil {
		logx.Errorw(err.Error())
		return errorx.NewInternalError(err.Error())
	} else {
		return nil
	}
}
