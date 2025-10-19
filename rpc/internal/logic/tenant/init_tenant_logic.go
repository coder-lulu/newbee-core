package tenant

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/ent/configuration"
	"github.com/coder-lulu/newbee-core/rpc/ent/dictionary"
	"github.com/coder-lulu/newbee-core/rpc/ent/menu"
	"github.com/coder-lulu/newbee-core/rpc/ent/position"
	"github.com/coder-lulu/newbee-core/rpc/ent/tenant"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/redisfunc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/coder-lulu/newbee-common/v2/i18n"
	"github.com/coder-lulu/newbee-common/v2/middleware/keys"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/hooks"
	"github.com/coder-lulu/newbee-common/v2/utils/encrypt"

	"github.com/gofrs/uuid/v5"
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
)

// TenantInitConfig 租户初始化配置（用于兼容性）
type TenantInitConfig struct {
	InitializedAt time.Time `json:"initialized_at"`
	Version       string    `json:"version"`
	Components    []string  `json:"components"`
	Status        string    `json:"status"`
}

type InitTenantLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	bypassPlugins bool
}

func NewInitTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitTenantLogic {
	return &InitTenantLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		bypassPlugins: false,
	}
}

// NewInitTenantLogicLegacy 构造仅使用旧版初始化逻辑的实例（供插件框架回退时使用）
func NewInitTenantLogicLegacy(ctx context.Context, svcCtx *svc.ServiceContext) *InitTenantLogic {
	return &InitTenantLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		bypassPlugins: true,
	}
}

// InitTenant 租户初始化逻辑
func (l *InitTenantLogic) InitTenant(in *core.TenantInitReq) (*core.BaseResp, error) {
	if !l.bypassPlugins {
		return NewInitTenantLogicV2(l.ctx, l.svcCtx).InitTenant(in)
	}

	return l.initTenantLegacy(in)
}

// initTenantLegacy 保留旧版初始化流程，供插件框架降级使用
func (l *InitTenantLogic) initTenantLegacy(in *core.TenantInitReq) (*core.BaseResp, error) {
	// 验证租户是否存在
	tenantInfo, err := l.svcCtx.DB.Tenant.Query().
		Where(tenant.IDEQ(in.TenantId)).
		Only(hooks.NewSystemContext(l.ctx))
	if err != nil {
		switch {
		case ent.IsNotFound(err):
			return nil, errorx.NewInvalidArgumentError("tenant.notFound")
		default:
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}
	}

	// 检查是否已经初始化过
	if tenantInfo.Config != nil {
		if status, exists := tenantInfo.Config["status"]; exists {
			if status == "completed" {
				logx.Infow("Tenant already initialized", logx.Field("tenant_id", in.TenantId))
				return &core.BaseResp{
					Msg: "租户已经初始化过",
				}, nil
			}
		}
	}

	// 创建租户专属的上下文,使用统一的ContextManager
	cm := keys.NewContextManager()
	tenantIDStr := strconv.FormatUint(in.TenantId, 10)
	tenantCtx := cm.SetTenantID(l.ctx, tenantIDStr)

	// 在事务中执行初始化
	tx, err := l.svcCtx.DB.Tx(tenantCtx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	// 记录初始化开始状态
	initConfig := map[string]interface{}{
		"initialized_at": time.Now().Format(time.RFC3339),
		"version":        "1.0.0",
		"components":     []string{},
		"status":         "initializing",
	}
	_, err = tx.Tenant.UpdateOneID(in.TenantId).
		SetConfig(initConfig).
		Save(hooks.NewSystemContext(tenantCtx))
	if err != nil {
		tx.Rollback()
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 第一阶段：初始化字典数据
	if err = l.initDictionaries(tenantCtx, tx, in.TenantId); err != nil {
		tx.Rollback()
		return nil, err
	}
	initConfig["components"] = append(initConfig["components"].([]string), "dictionaries")

	// 第二阶段：初始化系统配置
	if err = l.initConfigurations(tenantCtx, tx, in.TenantId); err != nil {
		tx.Rollback()
		return nil, err
	}
	initConfig["components"] = append(initConfig["components"].([]string), "configurations")

	// 第三阶段：初始化租户菜单（为租户创建菜单副本）
	if err = l.initTenantMenus(tenantCtx, tx, in.TenantId); err != nil {
		tx.Rollback()
		return nil, err
	}
	initConfig["components"] = append(initConfig["components"].([]string), "tenant_menus")

	// 第四阶段：API权限是系统级的，无需为每个租户单独初始化
	// if err := l.initAPIPermissions(tenantCtx, tx, in.TenantId); err != nil {
	// 	tx.Rollback()
	// 	return nil, err
	// }
	// initConfig.Components = append(initConfig.Components, "api_permissions")

	// 第五阶段：创建默认部门和职位
	dept, err := l.initDepartmentAndPositions(tenantCtx, tx, in.TenantId)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	initConfig["components"] = append(initConfig["components"].([]string), "departments", "positions")

	// 第六阶段：创建管理员角色和用户
	adminRole, adminUser, err := l.initAdminRoleAndUser(tenantCtx, tx, in, dept)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	initConfig["components"] = append(initConfig["components"].([]string), "admin_role", "admin_user")

	// 第七阶段：完成初始化状态更新
	initConfig["status"] = "completed"
	initConfig["completed_at"] = time.Now().Format(time.RFC3339)
	_, err = tx.Tenant.UpdateOneID(in.TenantId).
		SetConfig(initConfig).
		Save(hooks.NewSystemContext(tenantCtx))
	if err != nil {
		tx.Rollback()
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	logx.Infow("Tenant initialized successfully",
		logx.Field("tenant_id", in.TenantId),
		logx.Field("tenant_code", tenantInfo.Code),
		logx.Field("components", initConfig["components"]),
		logx.Field("admin_role_id", adminRole.ID),
		logx.Field("admin_user_id", adminUser.ID),
		logx.Field("department_id", dept.ID))

	if err := redisfunc.PublishCasbinReload(l.ctx, l.svcCtx.Redis, l.svcCtx.Config.RedisConf.Db, in.TenantId, "legacy_init"); err != nil {
		logx.Errorw("Failed to publish Casbin reload notification (legacy init)",
			logx.Field("tenant_id", in.TenantId),
			logx.Field("error", err.Error()))
	} else {
		logx.Infow("✅ Broadcasted Casbin reload notification (legacy init)",
			logx.Field("tenant_id", in.TenantId))
	}

	return &core.BaseResp{
		Msg: i18n.CreateSuccess,
	}, nil
}

// initDictionaries 初始化字典数据
func (l *InitTenantLogic) initDictionaries(ctx context.Context, tx *ent.Tx, tenantID uint64) error {
	// 检查是否已存在字典数据，避免重复创建
	existingCount, err := tx.Dictionary.Query().
		Where(dictionary.TenantIDEQ(tenantID)).
		Count(ctx)
	if err != nil {
		return dberrorhandler.DefaultEntError(l.Logger, err, nil)
	}
	if existingCount > 0 {
		logx.Infow("Dictionaries already exist, skipping", logx.Field("tenant_id", tenantID))
		return nil
	}

	// 创建基础字典
	dictionaries := []*ent.DictionaryCreate{
		tx.Dictionary.Create().
			SetTitle("用户性别").
			SetName("sys_user_sex").
			SetDesc("用户性别列表").
			SetStatus(1).
			SetTenantID(tenantID),
		tx.Dictionary.Create().
			SetTitle("系统状态").
			SetName("sys_common_status").
			SetDesc("通用状态列表").
			SetStatus(1).
			SetTenantID(tenantID),
		tx.Dictionary.Create().
			SetTitle("系统是否").
			SetName("sys_yes_no").
			SetDesc("系统是否列表").
			SetStatus(1).
			SetTenantID(tenantID),
		tx.Dictionary.Create().
			SetTitle("显示隐藏").
			SetName("sys_show_hide").
			SetDesc("菜单显示隐藏状态").
			SetStatus(1).
			SetTenantID(tenantID),
		tx.Dictionary.Create().
			SetTitle("操作类型").
			SetName("sys_oper_type").
			SetDesc("操作类型列表").
			SetStatus(1).
			SetTenantID(tenantID),
	}

	if err := tx.Dictionary.CreateBulk(dictionaries...).Exec(ctx); err != nil {
		return dberrorhandler.DefaultEntError(l.Logger, err, nil)
	}

	// 获取刚创建的字典记录，避免硬编码ID
	createdDicts, err := tx.Dictionary.Query().
		Where(dictionary.TenantIDEQ(tenantID)).
		Order(ent.Asc(dictionary.FieldID)).
		All(ctx)
	if err != nil {
		return dberrorhandler.DefaultEntError(l.Logger, err, nil)
	}

	// 创建字典名称到ID的映射
	dictMap := make(map[string]uint64)
	for _, dict := range createdDicts {
		dictMap[dict.Name] = dict.ID
	}

	// 为每个字典创建对应的详情数据
	var allDetails []*ent.DictionaryDetailCreate

	// 用户性别详情
	if sexDictID, exists := dictMap["sys_user_sex"]; exists {
		sexDetails := []*ent.DictionaryDetailCreate{
			tx.DictionaryDetail.Create().
				SetTitle("男").SetValue("0").SetDictionariesID(sexDictID).
				SetStatus(1).SetSort(1).SetIsDefault(1).SetTenantID(tenantID),
			tx.DictionaryDetail.Create().
				SetTitle("女").SetValue("1").SetDictionariesID(sexDictID).
				SetStatus(1).SetSort(2).SetIsDefault(0).SetTenantID(tenantID),
			tx.DictionaryDetail.Create().
				SetTitle("未知").SetValue("2").SetDictionariesID(sexDictID).
				SetStatus(1).SetSort(3).SetIsDefault(0).SetTenantID(tenantID),
		}
		allDetails = append(allDetails, sexDetails...)
	}

	// 系统状态详情
	if statusDictID, exists := dictMap["sys_common_status"]; exists {
		statusDetails := []*ent.DictionaryDetailCreate{
			tx.DictionaryDetail.Create().
				SetTitle("正常").SetValue("1").SetDictionariesID(statusDictID).
				SetStatus(1).SetSort(1).SetIsDefault(1).SetTenantID(tenantID),
			tx.DictionaryDetail.Create().
				SetTitle("停用").SetValue("0").SetDictionariesID(statusDictID).
				SetStatus(1).SetSort(2).SetIsDefault(0).SetTenantID(tenantID),
		}
		allDetails = append(allDetails, statusDetails...)
	}

	// 系统是否详情
	if yesNoDictID, exists := dictMap["sys_yes_no"]; exists {
		yesNoDetails := []*ent.DictionaryDetailCreate{
			tx.DictionaryDetail.Create().
				SetTitle("是").SetValue("1").SetDictionariesID(yesNoDictID).
				SetStatus(1).SetSort(1).SetIsDefault(0).SetTenantID(tenantID),
			tx.DictionaryDetail.Create().
				SetTitle("否").SetValue("0").SetDictionariesID(yesNoDictID).
				SetStatus(1).SetSort(2).SetIsDefault(1).SetTenantID(tenantID),
		}
		allDetails = append(allDetails, yesNoDetails...)
	}

	// 显示隐藏详情
	if showHideDictID, exists := dictMap["sys_show_hide"]; exists {
		showHideDetails := []*ent.DictionaryDetailCreate{
			tx.DictionaryDetail.Create().
				SetTitle("显示").SetValue("1").SetDictionariesID(showHideDictID).
				SetStatus(1).SetSort(1).SetIsDefault(1).SetTenantID(tenantID),
			tx.DictionaryDetail.Create().
				SetTitle("隐藏").SetValue("0").SetDictionariesID(showHideDictID).
				SetStatus(1).SetSort(2).SetIsDefault(0).SetTenantID(tenantID),
		}
		allDetails = append(allDetails, showHideDetails...)
	}

	// 操作类型详情
	if operTypeDictID, exists := dictMap["sys_oper_type"]; exists {
		operTypeDetails := []*ent.DictionaryDetailCreate{
			tx.DictionaryDetail.Create().
				SetTitle("新增").SetValue("1").SetDictionariesID(operTypeDictID).
				SetStatus(1).SetSort(1).SetIsDefault(0).SetTenantID(tenantID),
			tx.DictionaryDetail.Create().
				SetTitle("修改").SetValue("2").SetDictionariesID(operTypeDictID).
				SetStatus(1).SetSort(2).SetIsDefault(0).SetTenantID(tenantID),
			tx.DictionaryDetail.Create().
				SetTitle("删除").SetValue("3").SetDictionariesID(operTypeDictID).
				SetStatus(1).SetSort(3).SetIsDefault(0).SetTenantID(tenantID),
		}
		allDetails = append(allDetails, operTypeDetails...)
	}

	// 批量创建所有详情数据
	if len(allDetails) > 0 {
		if err := tx.DictionaryDetail.CreateBulk(allDetails...).Exec(ctx); err != nil {
			return dberrorhandler.DefaultEntError(l.Logger, err, nil)
		}
	}

	return nil
}

// initConfigurations 初始化系统配置
func (l *InitTenantLogic) initConfigurations(ctx context.Context, tx *ent.Tx, tenantID uint64) error {
	// 检查是否已存在配置数据
	existingCount, err := tx.Configuration.Query().
		Where(configuration.TenantIDEQ(tenantID)).
		Count(ctx)
	if err != nil {
		return dberrorhandler.DefaultEntError(l.Logger, err, nil)
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

	if err := tx.Configuration.CreateBulk(configs...).Exec(ctx); err != nil {
		return dberrorhandler.DefaultEntError(l.Logger, err, nil)
	}

	return nil
}

// initDepartmentAndPositions 初始化部门和职位
func (l *InitTenantLogic) initDepartmentAndPositions(ctx context.Context, tx *ent.Tx, tenantID uint64) (*ent.Department, error) {
	// 创建默认部门
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
		Save(ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, nil)
	}

	// 检查是否已存在职位数据
	existingPosCount, err := tx.Position.Query().
		Where(position.TenantIDEQ(tenantID)).
		Count(ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, nil)
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

		if err := tx.Position.CreateBulk(positions...).Exec(ctx); err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, nil)
		}
	}

	return dept, nil
}

// initAdminRoleAndUser 初始化管理员角色和用户
func (l *InitTenantLogic) initAdminRoleAndUser(ctx context.Context, tx *ent.Tx, in *core.TenantInitReq, dept *ent.Department) (*ent.Role, *ent.User, error) {
	// 创建管理员角色
	adminRole, err := tx.Role.Create().
		SetName("超级管理员").
		SetCode("admin").
		SetDefaultRouter("/dashboard").
		SetRemark("租户超级管理员角色").
		SetStatus(1).
		SetSort(1).
		// 🔥 Phase 3: data_scope field removed - now managed via sys_casbin_rules
		SetTenantID(in.TenantId).
		Save(ctx)
	if err != nil {
		return nil, nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 为角色分配基础菜单权限（使用租户上下文而非系统上下文）
	// 注意：这里应该为租户创建专属的菜单副本，而不是共享全局菜单
	menus, err := l.svcCtx.DB.Menu.Query().
		Where(menu.DisabledEQ(false)).
		All(ctx) // 使用租户上下文，确保租户隔离
	if err != nil {
		// 如果租户没有菜单数据，需要先创建租户的菜单副本
		logx.Infow("No tenant menus found, tenant may need menu initialization",
			logx.Field("tenant_id", in.TenantId))
		// 这里可以选择跳过菜单权限分配，或者实现菜单复制逻辑
		return nil, nil, fmt.Errorf("租户菜单数据未初始化，请先初始化租户菜单")
	}

	// 为角色关联菜单（注意：这假设每个租户都有自己的菜单副本）
	if len(menus) > 0 {
		menuIDs := make([]uint64, len(menus))
		for i, m := range menus {
			menuIDs[i] = m.ID
		}

		_, err = tx.Role.UpdateOneID(adminRole.ID).
			AddMenuIDs(menuIDs...).
			Save(ctx)
		if err != nil {
			return nil, nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}

		logx.Infow("Assigned menus to admin role",
			logx.Field("tenant_id", in.TenantId),
			logx.Field("role_id", adminRole.ID),
			logx.Field("menu_count", len(menuIDs)))
	} else {
		logx.Infow("No menus available for role assignment",
			logx.Field("tenant_id", in.TenantId),
			logx.Field("role_id", adminRole.ID))
	}

	// 创建管理员用户
	username := "admin"
	if in.AdminUsername != nil && *in.AdminUsername != "" {
		username = *in.AdminUsername
	}

	password := "123456"
	if in.AdminPassword != nil && *in.AdminPassword != "" {
		password = *in.AdminPassword
	}

	// 获取租户信息以构建默认邮箱
	tenantInfo, err := l.svcCtx.DB.Tenant.Get(hooks.NewSystemContext(l.ctx), in.TenantId)
	if err != nil {
		return nil, nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	email := "admin@" + tenantInfo.Code + ".com"
	if in.AdminEmail != nil && *in.AdminEmail != "" {
		email = *in.AdminEmail
	}

	encryptedPassword := encrypt.BcryptEncrypt(password)
	userUUID := uuid.Must(uuid.NewV4())

	adminUser, err := tx.User.Create().
		SetID(userUUID).
		SetUsername(username).
		SetPassword(encryptedPassword).
		SetNickname("超级管理员").
		SetDescription("租户超级管理员").
		SetHomePath("/dashboard").
		SetEmail(email).
		SetStatus(1).
		SetDepartmentID(dept.ID).
		SetTenantID(in.TenantId).
		Save(ctx)
	if err != nil {
		return nil, nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 为用户分配管理员角色
	_, err = tx.User.UpdateOneID(adminUser.ID).
		AddRoleIDs(adminRole.ID).
		Save(ctx)
	if err != nil {
		return nil, nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return adminRole, adminUser, nil
}

// initTenantMenus 为租户初始化菜单副本
func (l *InitTenantLogic) initTenantMenus(ctx context.Context, tx *ent.Tx, tenantID uint64) error {
	// 检查是否已存在租户菜单数据
	existingCount, err := tx.Menu.Query().
		Where(menu.TenantIDEQ(tenantID)).
		Count(ctx)
	if err != nil {
		return dberrorhandler.DefaultEntError(l.Logger, err, nil)
	}
	if existingCount > 0 {
		logx.Infow("Tenant menus already exist, skipping", logx.Field("tenant_id", tenantID))
		return nil
	}

	// 获取基础菜单模板（使用系统上下文获取全局菜单模板）
	baseMenus, err := l.svcCtx.DB.Menu.Query().
		Where(menu.DisabledEQ(false)).
		Order(ent.Asc(menu.FieldParentID), ent.Asc(menu.FieldSort)).
		All(hooks.NewSystemContext(l.ctx))
	if err != nil || len(baseMenus) == 0 {
		logx.Infow("No base menus found for tenant initialization",
			logx.Field("tenant_id", tenantID))
		// 如果没有基础菜单，创建基础菜单结构
		return l.createBasicMenus(ctx, tx, tenantID)
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
		return dberrorhandler.DefaultEntError(l.Logger, err, nil)
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
					return dberrorhandler.DefaultEntError(l.Logger, err, nil)
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
func (l *InitTenantLogic) createBasicMenus(ctx context.Context, tx *ent.Tx, tenantID uint64) error {
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
		{1, 0, "/dashboard", "Dashboard", "LAYOUT", "仪表盘", "lucide:area-chart", 0, 1},
		{2, 1, "/dashboard/workbench", "DashboardWorkbench", "/dashboard/workbench/index", "工作台", "lucide:square-chart-gantt", 1, 1},
		{1, 0, "/system", "System", "LAYOUT", "系统管理", "lucide:computer", 0, 2},
		{2, 1, "/system/user", "SystemUser", "/system/user/index", "用户管理", "lucide:circle-user-round", 2, 1},
		{2, 1, "/system/role", "SystemRole", "/system/role/index", "角色管理", "lucide:circle-user", 2, 2},
		{2, 1, "/system/department", "SystemDepartment", "/system/department/index", "部门管理", "lucide:git-branch-plus", 2, 3},
		{2, 1, "/system/menu", "SystemMenu", "/system/menu/index", "菜单管理", "lucide:menu", 2, 4},
		{2, 1, "/system/tenant", "SystemTenant", "/system/tenant/index", "租户管理", "lucide:building", 2, 5},
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
				return dberrorhandler.DefaultEntError(l.Logger, err, nil)
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
					return dberrorhandler.DefaultEntError(l.Logger, err, nil)
				}
			}
		}
	}

	logx.Infow("Created basic tenant menus", logx.Field("tenant_id", tenantID))
	return nil
}
