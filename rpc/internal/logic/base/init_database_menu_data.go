package base

import (
	"github.com/suyuan32/simple-admin-common/enum/common"
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/coder-lulu/newbee-core/rpc/ent"
)

// insert initial menu data
func (l *InitDatabaseLogic) insertMenuData() error {
	var menus []*ent.MenuCreate

	// 系统管理
	menus = append(menus, l.svcCtx.DB.Menu.Create().
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
		SetServiceName("Core"),
	)

	// 用户管理

	menus = append(menus, l.svcCtx.DB.Menu.Create().
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
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(2).
		SetName("查询用户").
		SetSort(1).
		SetTitle("查询用户").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:user:query").
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(2).
		SetName("创建用户").
		SetSort(2).
		SetTitle("创建用户").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:user:create").
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(2).
		SetName("修改用户").
		SetSort(3).
		SetTitle("修改用户").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:user:update").
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(2).
		SetName("删除用户").
		SetSort(4).
		SetTitle("删除用户").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:user:delete").
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(2).
		SetName("导出用户").
		SetSort(5).
		SetTitle("导出用户").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:user:export").
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(2).
		SetName("导入用户").
		SetSort(6).
		SetTitle("导入用户").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:user:import").
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(2).
		SetName("重置用户密码").
		SetSort(7).
		SetTitle("重置用户密码").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:user:resetPwd").
		SetServiceName("Core"),
	)

	// 角色管理  ID 10
	menus = append(menus, l.svcCtx.DB.Menu.Create().
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
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(10).
		SetName("查询角色").
		SetSort(1).
		SetTitle("查询角色").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:role:query").
		SetServiceName("Core"),
	)
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(10).
		SetName("创建角色").
		SetSort(2).
		SetTitle("创建角色").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:role:create").
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(10).
		SetName("更新角色").
		SetSort(3).
		SetTitle("更新角色").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:role:update").
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(10).
		SetName("删除角色").
		SetSort(4).
		SetTitle("删除角色").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:role:delete").
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(10).
		SetName("导出角色").
		SetSort(5).
		SetTitle("导出角色").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:role:export").
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(10).
		SetName("修改角色数据权限").
		SetSort(6).
		SetTitle("修改角色数据权限").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:role:changeDataScope").
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(10).
		SetName("修改角色接口权限").
		SetSort(7).
		SetTitle("修改角色接口权限").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:role:changeApiAuth").
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(10).
		SetName("分配角色").
		SetSort(8).
		SetTitle("分配角色").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:role:assignUser").
		SetServiceName("Core"),
	)

	//   ID 19
	menus = append(menus, l.svcCtx.DB.Menu.Create().
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
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(19).
		SetName("查询菜单").
		SetSort(1).
		SetTitle("查询菜单").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:menu:query").
		SetServiceName("Core"),
	)
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(19).
		SetName("创建菜单").
		SetSort(2).
		SetTitle("创建菜单").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:menu:create").
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(19).
		SetName("更新菜单").
		SetSort(3).
		SetTitle("更新菜单").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:menu:update").
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(19).
		SetName("删除菜单").
		SetSort(4).
		SetTitle("删除菜单").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:menu:delete").
		SetServiceName("Core"),
	)

	// 部门管理 ID 24

	menus = append(menus, l.svcCtx.DB.Menu.Create().
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
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(24).
		SetName("查询部门").
		SetSort(1).
		SetTitle("查询部门").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:dept:query").
		SetServiceName("Core"),
	)
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(24).
		SetName("创建部门").
		SetSort(2).
		SetTitle("创建部门").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:dept:create").
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(24).
		SetName("更新部门").
		SetSort(3).
		SetTitle("更新部门").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:dept:update").
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(24).
		SetName("删除部门").
		SetSort(4).
		SetTitle("删除部门").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:dept:delete").
		SetServiceName("Core"),
	)

	// 岗位管理 ID 29
	menus = append(menus, l.svcCtx.DB.Menu.Create().
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
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(29).
		SetName("查询岗位").
		SetSort(1).
		SetTitle("查询岗位").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:post:query").
		SetServiceName("Core"),
	)
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(29).
		SetName("创建岗位").
		SetSort(2).
		SetTitle("创建岗位").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:post:create").
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(29).
		SetName("更新岗位").
		SetSort(3).
		SetTitle("更新岗位").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:post:update").
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(29).
		SetName("删除岗位").
		SetSort(4).
		SetTitle("删除岗位").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:post:delete").
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(29).
		SetName("导出岗位").
		SetSort(5).
		SetTitle("导出岗位").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:post:export").
		SetServiceName("Core"),
	)

	// 字典管理 ID 35

	menus = append(menus, l.svcCtx.DB.Menu.Create().
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
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(35).
		SetName("查询字典").
		SetSort(1).
		SetTitle("查询字典").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:dict:query").
		SetServiceName("Core"),
	)
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(35).
		SetName("创建字典").
		SetSort(2).
		SetTitle("创建字典").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:dict:create").
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(35).
		SetName("更新字典").
		SetSort(3).
		SetTitle("更新字典").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:dict:update").
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(35).
		SetName("删除字典").
		SetSort(4).
		SetTitle("删除字典").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:dict:delete").
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(35).
		SetName("导出字典").
		SetSort(5).
		SetTitle("导出字典").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:dict:export").
		SetServiceName("Core"),
	)

	//参数设置 ID 41

	menus = append(menus, l.svcCtx.DB.Menu.Create().
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
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(41).
		SetName("查询参数").
		SetSort(1).
		SetTitle("查询参数").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:config:query").
		SetServiceName("Core"),
	)
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(41).
		SetName("创建参数").
		SetSort(2).
		SetTitle("创建参数").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:config:create").
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(41).
		SetName("更新参数").
		SetSort(3).
		SetTitle("更新参数").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:config:update").
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(41).
		SetName("删除参数").
		SetSort(4).
		SetTitle("删除参数").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:config:delete").
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(41).
		SetName("导出参数").
		SetSort(5).
		SetTitle("导出参数").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:config:export").
		SetServiceName("Core"),
	)

	//接口管理 ID 47

	menus = append(menus, l.svcCtx.DB.Menu.Create().
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
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(47).
		SetName("查询接口").
		SetSort(1).
		SetTitle("查询接口").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:api:query").
		SetServiceName("Core"),
	)
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(47).
		SetName("创建接口").
		SetSort(2).
		SetTitle("创建接口").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:api:create").
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(47).
		SetName("更新接口").
		SetSort(3).
		SetTitle("更新接口").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:api:update").
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(47).
		SetName("删除接口").
		SetSort(4).
		SetTitle("删除接口").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:api:delete").
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(47).
		SetName("导出接口").
		SetSort(5).
		SetTitle("导出接口").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:api:export").
		SetServiceName("Core"),
	)

	//三方登录  ID 53

	menus = append(menus, l.svcCtx.DB.Menu.Create().
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
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(53).
		SetName("查询三方登录").
		SetSort(1).
		SetTitle("查询三方登录").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:oauth:query").
		SetServiceName("Core"),
	)
	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(47).
		SetName("创建三方登录").
		SetSort(2).
		SetTitle("创建三方登录").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:oauth:create").
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(47).
		SetName("更新三方登录").
		SetSort(3).
		SetTitle("更新三方登录").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:oauth:update").
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(53).
		SetName("删除三方登录").
		SetSort(4).
		SetTitle("删除三方登录").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:oauth:delete").
		SetServiceName("Core"),
	)

	// token管理 ID 58

	menus = append(menus, l.svcCtx.DB.Menu.Create().
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
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(58).
		SetName("查询令牌").
		SetSort(1).
		SetTitle("查询令牌").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:token:query").
		SetServiceName("Core"),
	)

	menus = append(menus, l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(58).
		SetName("删除令牌").
		SetSort(2).
		SetTitle("删除令牌").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:token:delete").
		SetServiceName("Core"),
	)

	err := l.svcCtx.DB.Menu.CreateBulk(menus...).Exec(l.ctx)
	if err != nil {
		logx.Errorw(err.Error())
		return errorx.NewInternalError(err.Error())
	} else {
		return nil
	}
}
