package base

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bsm/redislock"
	"github.com/coder-lulu/newbee-common/v2/config"

	"github.com/coder-lulu/newbee-core/rpc/ent/casbinrule"
	"github.com/coder-lulu/newbee-core/rpc/ent/role"
	"github.com/coder-lulu/newbee-core/rpc/ent/tenant"
	"github.com/coder-lulu/newbee-core/rpc/ent/user"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/redisfunc"

	"entgo.io/ent/dialect/sql/schema"
	"github.com/coder-lulu/newbee-common/v2/enum/common"
	"github.com/coder-lulu/newbee-common/v2/i18n"
	"github.com/coder-lulu/newbee-common/v2/msg/logmsg"
	"github.com/coder-lulu/newbee-common/v2/utils/encrypt"
	"github.com/zeromicro/go-zero/core/errorx"

	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/coder-lulu/newbee-common/v2/orm/ent/hooks"
	"github.com/zeromicro/go-zero/core/logx"
)

type InitDatabaseLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewInitDatabaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitDatabaseLogic {
	return &InitDatabaseLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *InitDatabaseLogic) InitDatabase(_ *core.Empty) (*core.BaseResp, error) {
	// If your mysql speed is high, comment the code below.
	// Because the context deadline will reach if the database is too slow
	l.ctx = context.Background()

	// 使用系统上下文绕过租户隔离限制
	systemCtx := hooks.NewSystemContext(l.ctx)

	// add lock to avoid duplicate initialization
	locker := redislock.New(l.svcCtx.Redis)

	lock, err := locker.Obtain(l.ctx, "INIT:DATABASE:LOCK", 10*time.Minute, nil)
	if errors.Is(err, redislock.ErrNotObtained) {
		logx.Error("last initialization is running")
		return nil, errorx.NewInternalError("i18n.InitRunning")
	} else if err != nil {
		logx.Errorw(logmsg.RedisError, logx.Field("detail", err.Error()))
		return nil, errorx.NewInternalError("failed to get redis lock")
	}

	defer lock.Release(l.ctx)

	// initialize table structure
	if err = l.svcCtx.DB.Schema.Create(systemCtx, schema.WithForeignKeys(false), schema.WithDropColumn(true),
		schema.WithDropIndex(true)); err != nil {
		logx.Errorw(logmsg.DatabaseError, logx.Field("detail", err.Error()))
		_ = l.svcCtx.Redis.Set(l.ctx, "INIT:DATABASE:ERROR", err.Error(), 300*time.Second).Err()
		return nil, errorx.NewInternalError(err.Error())
	}

	// remove all data perm cache
	err = redisfunc.RemoveAllKeyByPrefix(l.ctx, config.RedisDataPermissionPrefix, l.svcCtx.Redis)
	if err != nil {
		logx.Errorw(logmsg.RedisError, logx.Field("detail", err.Error()))
		return nil, errorx.NewInternalError(i18n.RedisError)
	}

	// judge if the initialization had been done
	check, err := l.svcCtx.DB.API.Query().Count(systemCtx)

	if check != 0 {
		err = l.svcCtx.Redis.Set(l.ctx, "INIT:DATABASE:STATE", "1", 24*time.Hour).Err()
		if err != nil {
			logx.Errorw(logmsg.RedisError, logx.Field("detail", err.Error()))
			return nil, errorx.NewInternalError(i18n.RedisError)
		}

		// 🔥 Phase 3: data_scope field removed - data permission now managed via sys_casbin_rules
		// No longer need to reset data_scope in sys_roles table

		return &core.BaseResp{Msg: i18n.AlreadyInit}, nil
	}

	// set default state value
	_ = l.svcCtx.Redis.Set(l.ctx, "INIT:DATABASE:ERROR", "", 300*time.Second)
	_ = l.svcCtx.Redis.Set(l.ctx, "INIT:DATABASE:STATE", "0", 300*time.Second)

	errHandler := func(err error) (*core.BaseResp, error) {
		logx.Errorw(logmsg.DatabaseError, logx.Field("detail", err.Error()))
		_ = l.svcCtx.Redis.Set(l.ctx, "INIT:DATABASE:ERROR", err.Error(), 300*time.Second)
		return nil, errorx.NewInternalError(err.Error())
	}

	// 🔥 修复顺序: 先创建租户、部门、职位,再创建依赖它们的用户和角色

	// 第一步: 创建租户
	err = l.insertTenantData(systemCtx)
	if err != nil {
		return errHandler(err)
	}

	// 第二步: 创建部门 (用户需要部门)
	err = l.insertDepartmentData(systemCtx)
	if err != nil {
		return errHandler(err)
	}

	// 第三步: 创建职位 (用户需要职位)
	err = l.insertPositionData(systemCtx)
	if err != nil {
		return errHandler(err)
	}

	// 第四步: 创建菜单 (角色需要关联菜单)
	err = l.insertMenuData(systemCtx)
	if err != nil {
		return errHandler(err)
	}

	// 第五步: 创建角色 (用户需要角色)
	err = l.insertRoleData(systemCtx)
	if err != nil {
		return errHandler(err)
	}

	// 第六步: 创建用户 (现在所有依赖都已就绪)
	err = l.insertUserData(systemCtx)
	if err != nil {
		return errHandler(err)
	}

	// 第七步: 创建API数据
	err = l.insertApiData(systemCtx)
	if err != nil {
		return errHandler(err)
	}

	// 第八步: 关联角色和菜单权限
	err = l.insertRoleMenuAuthorityData(systemCtx)
	if err != nil {
		return errHandler(err)
	}

	// 第九步: 其他数据
	err = l.insertProviderData(systemCtx)
	if err != nil {
		return errHandler(err)
	}

	err = l.insertCasbinPoliciesData(systemCtx)
	if err != nil {
		return errHandler(err)
	}

	// 🔥 Phase 2: 创建数据权限规则到sys_casbin_rules
	err = l.insertDataPermRules(systemCtx)
	if err != nil {
		return errHandler(err)
	}

	err = l.insertDictData(systemCtx)
	if err != nil {
		return errHandler(err)
	}

	err = l.insertOAuthEnhancedData(systemCtx)
	if err != nil {
		return errHandler(err)
	}

	_ = l.svcCtx.Redis.Set(l.ctx, "INIT:DATABASE:STATE", "1", 24*time.Hour)
	return &core.BaseResp{Msg: i18n.Success}, nil
}

// insert initial tenant data
func (l *InitDatabaseLogic) insertTenantData(ctx context.Context) error {
	// 直接使用传入的系统上下文，避免租户隔离限制

	// 🔥 检查是否已存在默认租户
	existing, err := l.svcCtx.DB.Tenant.Query().
		Where(tenant.CodeEQ("default")).
		First(ctx)

	if err == nil && existing != nil {
		// 租户已存在，记录日志并跳过
		logx.Infow("Default tenant already exists, skipping creation",
			logx.Field("tenant_id", existing.ID),
			logx.Field("tenant_code", existing.Code),
			logx.Field("tenant_name", existing.Name))
		return nil
	}

	// 创建默认租户
	defaultConfig := map[string]interface{}{
		"max_users":        1000,
		"storage_limit_gb": 100,
		"features":         []string{"all"},
		"is_default":       true,
	}

	tenant, err := l.svcCtx.DB.Tenant.Create().
		SetName("默认租户").
		SetCode("default").
		SetDescription("系统默认租户，用于超级管理员和系统管理").
		SetStatus(1).
		SetExpiredAt(time.Now().AddDate(10, 0, 0)). // 10年后过期
		SetConfig(defaultConfig).
		SetCreatedBy(0). // 系统创建
		Save(ctx)

	if err != nil {
		logx.Errorw("Failed to create default tenant", logx.Field("error", err.Error()))
		return errorx.NewInternalError(err.Error())
	}

	logx.Infow("✅ Default tenant created successfully",
		logx.Field("tenant_id", tenant.ID),
		logx.Field("tenant_code", tenant.Code))
	return nil
}

// insert init user data
func (l *InitDatabaseLogic) insertUserData(ctx context.Context) error {
	// 🔥 检查admin用户是否已存在
	tenantID := uint64(1)
	ctxWithTenant := hooks.SetTenantIDToContext(context.Background(), tenantID)
	existingUser, err := l.svcCtx.DB.User.Query().
		Where(user.UsernameEQ("admin"), user.TenantIDEQ(tenantID)).
		First(ctxWithTenant)

	if err == nil && existingUser != nil {
		// 用户已存在，记录日志并跳过
		logx.Infow("Admin user already exists, skipping creation",
			logx.Field("user_id", existingUser.ID),
			logx.Field("username", existingUser.Username),
			logx.Field("tenant_id", existingUser.TenantID))
		return nil
	}

	// 创建admin用户
	var users []*ent.UserCreate
	users = append(users, l.svcCtx.DB.User.Create().
		SetUsername("admin").
		SetNickname("admin").
		SetPassword(encrypt.BcryptEncrypt("123456")).
		SetEmail("530077128@qq.com").
		AddRoleIDs(1).
		SetDepartmentID(1).
		SetHomePath("/workspace").
		AddPositionIDs(1).
		SetTenantID(tenantID), // 🔥 修复: 显式设置租户ID为1
	)

	// 🔥 修复: 使用Save()而不是Exec(),以便获取创建的记录并验证tenant_id
	createdUsers, err := l.svcCtx.DB.User.CreateBulk(users...).Save(ctxWithTenant)
	if err != nil {
		logx.Errorw("❌ Failed to create admin user", logx.Field("error", err.Error()))
		return errorx.NewInternalError(err.Error())
	}

	// 🔥 修复: 验证创建的用户tenant_id是否正确
	if len(createdUsers) > 0 {
		adminUser := createdUsers[0]
		if adminUser.TenantID == 0 {
			logx.Errorw("🚨 Critical: Admin user created with tenant_id=0! This is a bug!",
				logx.Field("user_id", adminUser.ID),
				logx.Field("username", adminUser.Username),
				logx.Field("tenant_id", adminUser.TenantID))
			return errorx.NewInternalError("admin user created with invalid tenant_id=0")
		}
		logx.Infow("✅ Admin user created successfully",
			logx.Field("user_id", adminUser.ID),
			logx.Field("username", adminUser.Username),
			logx.Field("tenant_id", adminUser.TenantID),
			logx.Field("department_id", adminUser.DepartmentID))
	}

	return nil
}

// insert initial role data
func (l *InitDatabaseLogic) insertRoleData(ctx context.Context) error {
	// 🔥 检查角色是否已存在
	tenantID := uint64(1)
	ctxWithTenant := hooks.SetTenantIDToContext(context.Background(), tenantID)

	// 检查超级管理员角色
	superAdminExists, _ := l.svcCtx.DB.Role.Query().
		Where(role.CodeEQ("superadmin"), role.TenantIDEQ(tenantID)).
		Exist(ctxWithTenant)

	// 检查普通用户角色
	userExists, _ := l.svcCtx.DB.Role.Query().
		Where(role.CodeEQ("user"), role.TenantIDEQ(tenantID)).
		Exist(ctxWithTenant)

	var roles []*ent.RoleCreate

	// 只创建不存在的角色
	if !superAdminExists {
		roles = append(roles, l.svcCtx.DB.Role.Create().
			SetName("超级管理员").
			SetCode("superadmin").
			SetRemark("超级管理员").
			SetDefaultRouter("workspace").
			// 🔥 Phase 3: data_scope field removed - now managed via sys_casbin_rules
			SetSort(1).
			SetTenantID(tenantID),
		)
		logx.Info("Will create superadmin role")
	} else {
		logx.Info("Superadmin role already exists, skipping")
	}

	if !userExists {
		roles = append(roles, l.svcCtx.DB.Role.Create().
			SetName("普通用户").
			SetCode("user").
			SetRemark("普通员工").
			SetDefaultRouter("workspace").
			SetSort(2).
			SetTenantID(tenantID),
		)
		logx.Info("Will create user role")
	} else {
		logx.Info("User role already exists, skipping")
	}

	// 如果有需要创建的角色，执行批量创建
	if len(roles) > 0 {
		err := l.svcCtx.DB.Role.CreateBulk(roles...).Exec(ctxWithTenant)
		if err != nil {
			logx.Errorw("Failed to create roles", logx.Field("error", err.Error()))
			return errorx.NewInternalError(err.Error())
		}
		logx.Infof("✅ Created %d roles", len(roles))
	} else {
		logx.Info("All roles already exist, no creation needed")
	}

	return nil
}

// insert initial admin menu authority data
func (l *InitDatabaseLogic) insertRoleMenuAuthorityData(ctx context.Context) error {
	ctxWithTenant := hooks.SetTenantIDToContext(context.Background(), 1)
	count, err := l.svcCtx.DB.Menu.Query().Count(ctxWithTenant)
	if err != nil {
		logx.Errorw(err.Error())
		return errorx.NewInternalError(err.Error())
	}

	// Super admin gets all menus
	var allMenuIds []uint64
	allMenuIds = make([]uint64, count)
	for i := range allMenuIds {
		allMenuIds[i] = uint64(i + 1)
	}

	err = l.svcCtx.DB.Role.Update().AddMenuIDs(allMenuIds...).Where(role.IDEQ(1)).Exec(ctxWithTenant)
	if err != nil {
		logx.Errorw(err.Error())
		return errorx.NewInternalError(err.Error())
	}

	// Normal user gets limited menus (only basic user-related functionality)
	// These correspond to basic user profile and personal settings
	var userMenuIds []uint64
	// No system management menus for normal users - they only get basic access
	// The specific menu IDs would need to be determined based on actual menu structure
	// For now, we give them minimal access - they can manage their own profile

	if count > 0 {
		// Only add very basic menus that don't require system admin permissions
		// This is intentionally limited for security
		userMenuIds = []uint64{} // Empty for now - normal users have no menu access by default
	}

	if len(userMenuIds) > 0 {
		err = l.svcCtx.DB.Role.Update().AddMenuIDs(userMenuIds...).Where(role.IDEQ(2)).Exec(ctxWithTenant)
		if err != nil {
			logx.Errorw(err.Error())
			return errorx.NewInternalError(err.Error())
		}
	}

	return nil
}

// insert initial Casbin policies
func (l *InitDatabaseLogic) insertCasbinPoliciesData(ctx context.Context) error {
	tenantID := uint64(1)
	ctxWithTenant := hooks.SetTenantIDToContext(context.Background(), tenantID)
	apis, err := l.svcCtx.DB.API.Query().All(ctxWithTenant)
	if err != nil {
		logx.Errorw(err.Error())
		return errorx.NewInternalError(err.Error())
	}

	logx.Info("apis", logx.Field("apis", apis))

	// 🔥 架构修复: 直接使用ent client写入sys_casbin_rules表
	// 原问题: 之前使用gormadapter会写入错误的casbin_rule表
	// 新方案: 直接创建CasbinRule实体，确保写入sys_casbin_rules表

	// Super admin gets all API permissions
	var superAdminRoleCode string
	if adminData, roleErr := l.svcCtx.DB.Role.Query().Where(role.NameEQ("超级管理员")).First(ctxWithTenant); roleErr == nil {
		superAdminRoleCode = adminData.Code
	} else {
		superAdminRoleCode = "superadmin"
	}

	// Normal user gets limited API permissions (only basic required APIs)
	var normalUserRoleCode string
	if userData, roleErr := l.svcCtx.DB.Role.Query().Where(role.NameEQ("普通用户")).First(ctxWithTenant); roleErr == nil {
		normalUserRoleCode = userData.Code
	} else {
		normalUserRoleCode = "user"
	}

	// Define basic APIs that normal users should have access to
	basicAPIs := map[string][]string{
		"/user/login":           {"POST"},
		"/user/info":            {"GET"},
		"/user/change_password": {"POST"},
		"/user/profile":         {"GET", "POST"},
		"/user/perm":            {"GET"},
		"/user/logout":          {"GET"},
		"/captcha":              {"GET"},
		"/oauth/login":          {"POST"},
		"/user/refresh_token":   {"GET"},
		"/user/access_token":    {"GET"},
	}

	// Clear old policies for both roles (using ent)
	_, err = l.svcCtx.DB.CasbinRule.Delete().
		Where(
			casbinrule.V0EQ(superAdminRoleCode),
			casbinrule.TenantIDEQ(tenantID),
		).
		Exec(ctxWithTenant)
	if err != nil {
		logx.Errorw("failed to delete old super admin policies", logx.Field("error", err.Error()))
		return errorx.NewInternalError(err.Error())
	}

	_, err = l.svcCtx.DB.CasbinRule.Delete().
		Where(
			casbinrule.V0EQ(normalUserRoleCode),
			casbinrule.TenantIDEQ(tenantID),
		).
		Exec(ctxWithTenant)
	if err != nil {
		logx.Errorw("failed to delete old normal user policies", logx.Field("error", err.Error()))
		return errorx.NewInternalError(err.Error())
	}

	// Prepare bulk insert for super admin policies
	var superAdminBulk []*ent.CasbinRuleCreate
	for _, api := range apis {
		// RBAC with Domains 格式: [sub, domain, obj, act, eft]
		// Ptype="p", V0=角色code, V1=租户ID(domain), V2=路径(obj), V3=方法(act), V4=效果(eft)
		superAdminBulk = append(superAdminBulk, l.svcCtx.DB.CasbinRule.Create().
			SetPtype("p").
			SetV0(superAdminRoleCode).          // subject: 角色code
			SetV1(fmt.Sprintf("%d", tenantID)). // domain: 租户ID
			SetV2(api.Path).                    // object: API路径
			SetV3(api.Method).                  // action: HTTP方法
			SetV4("allow").                     // effect: 效果
			SetServiceName("core").
			SetRuleName(fmt.Sprintf("Super admin access to %s %s", api.Method, api.Path)).
			SetCategory("system").
			SetVersion("1.0.0").
			SetRequireApproval(false).
			SetApprovalStatus("approved").
			SetStatus(1).
			SetTenantID(tenantID),
		)
	}

	// Prepare bulk insert for normal user policies
	var normalUserBulk []*ent.CasbinRuleCreate
	for path, methods := range basicAPIs {
		for _, method := range methods {
			normalUserBulk = append(normalUserBulk, l.svcCtx.DB.CasbinRule.Create().
				SetPtype("p").
				SetV0(normalUserRoleCode).          // subject: 角色code
				SetV1(fmt.Sprintf("%d", tenantID)). // domain: 租户ID
				SetV2(path).                        // object: API路径
				SetV3(method).                      // action: HTTP方法
				SetV4("allow").                     // effect: 效果
				SetServiceName("core").
				SetRuleName(fmt.Sprintf("Normal user access to %s %s", method, path)).
				SetCategory("system").
				SetVersion("1.0.0").
				SetRequireApproval(false).
				SetApprovalStatus("approved").
				SetStatus(1).
				SetTenantID(tenantID),
			)
		}
	}

	// Execute bulk insert for super admin policies
	if len(superAdminBulk) > 0 {
		_, err = l.svcCtx.DB.CasbinRule.CreateBulk(superAdminBulk...).Save(ctx)
		if err != nil {
			logx.Errorw("failed to insert super admin policies", logx.Field("error", err.Error()))
			return errorx.NewInternalError(err.Error())
		}
		logx.Infof("✅ Inserted %d super admin policies to sys_casbin_rules", len(superAdminBulk))
	}

	// Execute bulk insert for normal user policies
	if len(normalUserBulk) > 0 {
		_, err = l.svcCtx.DB.CasbinRule.CreateBulk(normalUserBulk...).Save(ctx)
		if err != nil {
			logx.Errorw("failed to insert normal user policies", logx.Field("error", err.Error()))
			return errorx.NewInternalError(err.Error())
		}
		logx.Infof("✅ Inserted %d normal user policies to sys_casbin_rules", len(normalUserBulk))
	}

	// 🔥 重要: 通过Redis发送更新通知，触发所有服务实例重新加载策略
	// 使用Casbin RedisWatcher的标准通知机制
	updateMsg := fmt.Sprintf("UpdatePolicy:tenant_%d", tenantID)
	err = l.svcCtx.Redis.Publish(ctx, "casbin_watcher", updateMsg).Err()
	if err != nil {
		logx.Errorw("failed to publish policy update notification", logx.Field("error", err.Error()))
		// 不返回错误，因为策略已经写入数据库，只是通知失败
	} else {
		logx.Info("✅ Published policy update notification to Redis")
	}

	return nil
}

// insert initial provider data
func (l *InitDatabaseLogic) insertProviderData(ctx context.Context) error {
	var providers []*ent.OauthProviderCreate

	providers = append(providers, l.svcCtx.DB.OauthProvider.Create().
		SetName("google").
		SetDisplayName("Google").
		SetType("google").
		SetClientID("your client id").
		SetClientSecret("your client secret").
		SetRedirectURL("http://localhost:3100/oauth/login/callback").
		SetScopes("email openid").
		SetAuthURL("https://accounts.google.com/o/oauth2/auth").
		SetTokenURL("https://oauth2.googleapis.com/token").
		SetAuthStyle(1).
		SetInfoURL("https://www.googleapis.com/oauth2/v2/userinfo?access_token=TOKEN"),
	)

	providers = append(providers, l.svcCtx.DB.OauthProvider.Create().
		SetName("github").
		SetDisplayName("GitHub").
		SetType("github").
		SetClientID("your client id").
		SetClientSecret("your client secret").
		SetRedirectURL("http://localhost:3100/oauth/login/callback").
		SetScopes("email openid").
		SetAuthURL("https://github.com/login/oauth/authorize").
		SetTokenURL("https://github.com/login/oauth/access_token").
		SetAuthStyle(2).
		SetInfoURL("https://api.github.com/user"),
	)

	tenantCtx := hooks.SetTenantIDToContext(context.Background(), 1)
	err := l.svcCtx.DB.OauthProvider.CreateBulk(providers...).Exec(tenantCtx)
	if err != nil {
		logx.Errorw(err.Error())
		return errorx.NewInternalError(err.Error())
	} else {
		return nil
	}
}

// insert initial department data
func (l *InitDatabaseLogic) insertDepartmentData(ctx context.Context) error {
	var departments []*ent.DepartmentCreate
	departments = append(departments, l.svcCtx.DB.Department.Create().
		SetName("根部门").
		SetAncestors("").
		SetLeader("admin").
		SetEmail("530077128@qq.com").
		SetPhone("18888888888").
		SetRemark("Super Administrator").
		SetSort(1).
		SetParentID(common.DefaultParentId).
		SetTenantID(1), // 设置默认租户ID
	)

	ctxWithTenant := hooks.SetTenantIDToContext(context.Background(), 1)
	err := l.svcCtx.DB.Department.CreateBulk(departments...).Exec(ctxWithTenant)
	if err != nil {
		logx.Errorw(err.Error())
		return errorx.NewInternalError(err.Error())
	} else {
		return nil
	}
}

// insert initial position data
func (l *InitDatabaseLogic) insertPositionData(ctx context.Context) error {
	var posts []*ent.PositionCreate
	posts = append(posts, l.svcCtx.DB.Position.Create().
		SetName("主管").
		SetRemark("CEO").SetCode("ceo").SetSort(1).
		SetTenantID(1), // 设置默认租户ID
	)

	ctxWithTenant := hooks.SetTenantIDToContext(context.Background(), 1)
	err := l.svcCtx.DB.Position.CreateBulk(posts...).Exec(ctxWithTenant)
	if err != nil {
		logx.Errorw(err.Error())
		return errorx.NewInternalError(err.Error())
	} else {
		return nil
	}
}

// insert OAuth enhanced menu and API data
func (l *InitDatabaseLogic) insertOAuthEnhancedData(ctx context.Context) error {
	ctxWithTenant := hooks.SetTenantIDToContext(context.Background(), 1)

	// Add OAuth statistics submenu (ID will be 67)
	_, err := l.svcCtx.DB.Menu.Create().
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(1).
		SetPath("oauth-statistics").
		SetName("OAuth统计分析").
		SetComponent("system/oauth/statistics").
		SetSort(12).
		SetTitle("OAuth统计分析").
		SetIcon("lucide:bar-chart-3").
		SetHideMenu(false).
		SetParams("").
		SetPermission("system:oauth:statistics:view").
		SetServiceName("Core").
		SetTenantID(1).
		Save(ctxWithTenant)

	if err != nil {
		logx.Errorw("Failed to create OAuth statistics menu", logx.Field("error", err.Error()))
		return errorx.NewInternalError(err.Error())
	}

	// Add OAuth provider test permission (parent: OAuth Provider menu ID 53)
	_, err = l.svcCtx.DB.Menu.Create().
		SetMenuLevel(3).
		SetMenuType(2).
		SetParentID(53).
		SetPath("/system/oauth/test").
		SetName("测试OAuth连接").
		SetSort(5).
		SetTitle("测试OAuth连接").
		SetIcon("#").
		SetHideMenu(true).
		SetPermission("system:oauth:test").
		SetServiceName("Core").
		SetTenantID(1).
		Save(ctxWithTenant)

	if err != nil {
		logx.Errorw("Failed to create OAuth test permission", logx.Field("error", err.Error()))
		return errorx.NewInternalError(err.Error())
	}

	logx.Info("OAuth enhanced menu data inserted successfully")
	return nil
}

// 🔥 Phase 2: 创建默认数据权限规则到sys_casbin_rules
// insert initial data permission rules for default roles
func (l *InitDatabaseLogic) insertDataPermRules(ctx context.Context) error {
	tenantID := uint64(1) // Default tenant
	ctxWithTenant := hooks.SetTenantIDToContext(context.Background(), tenantID)

	// 定义默认数据权限规则
	type DataPermRule struct {
		RoleCode    string
		RoleName    string
		DataScope   string   // all, custom_dept, own_dept_and_sub, own_dept, own
		CustomDepts []string // 自定义部门ID列表（仅custom_dept时使用）
	}

	defaultRules := []DataPermRule{
		{
			RoleCode:  "superadmin",
			RoleName:  "超级管理员",
			DataScope: "all",
		},
		{
			RoleCode:  "user",
			RoleName:  "普通用户",
			DataScope: "own_dept_and_sub",
		},
	}

	// 清理旧的数据权限规则（ptype=d）
	_, err := l.svcCtx.DB.CasbinRule.Delete().
		Where(
			casbinrule.PtypeEQ("d"),
			casbinrule.TenantIDEQ(tenantID),
		).
		Exec(ctxWithTenant)
	if err != nil {
		logx.Errorw("failed to delete old data perm rules", logx.Field("error", err.Error()))
		return errorx.NewInternalError(err.Error())
	}

	// 批量创建数据权限规则
	var bulk []*ent.CasbinRuleCreate
	for _, rule := range defaultRules {
		// 构造v4字段（自定义部门列表JSON）
		v4 := ""
		if rule.DataScope == "custom_dept" && len(rule.CustomDepts) > 0 {
			// 将部门ID列表转换为JSON数组
			v4 = fmt.Sprintf("[\"%s\"]", rule.CustomDepts[0])
			for i := 1; i < len(rule.CustomDepts); i++ {
				v4 = v4[:len(v4)-1] + fmt.Sprintf(",\"%s\"]", rule.CustomDepts[i])
			}
		}

		// 创建数据权限规则
		// ptype=d: 数据权限规则
		// v0: 角色代码
		// v1: 租户ID（domain）
		// v2: 资源类型（* 表示所有资源）
		// v3: 数据权限范围
		// v4: 自定义部门ID列表（JSON数组）
		bulk = append(bulk, l.svcCtx.DB.CasbinRule.Create().
			SetPtype("d").                                     // 数据权限规则类型
			SetV0(rule.RoleCode).                              // subject: 角色代码
			SetV1(fmt.Sprintf("%d", tenantID)).                // domain: 租户ID
			SetV2("*").                                        // object: 资源类型（* 表示所有）
			SetV3(rule.DataScope).                             // action: 数据权限范围
			SetV4(v4).                                         // effect: 自定义部门列表
			SetServiceName("core").                            // 服务名称
			SetRuleName(fmt.Sprintf("%s数据权限", rule.RoleName)). // 规则名称
			SetDescription(fmt.Sprintf("角色%s的默认数据权限规则，数据范围：%s", rule.RoleName, rule.DataScope)).
			SetCategory("data_permission"). // 规则分类
			SetVersion("1.0.0").
			SetRequireApproval(false).
			SetApprovalStatus("approved").
			SetStatus(1).
			SetTenantID(tenantID),
		)
	}

	// 执行批量创建
	if len(bulk) > 0 {
		_, err = l.svcCtx.DB.CasbinRule.CreateBulk(bulk...).Save(ctxWithTenant)
		if err != nil {
			logx.Errorw("failed to insert data perm rules", logx.Field("error", err.Error()))
			return errorx.NewInternalError(err.Error())
		}
		logx.Infof("✅ Inserted %d data permission rules to sys_casbin_rules", len(bulk))

		// 输出详细的规则信息
		for _, rule := range defaultRules {
			logx.Infow("Data permission rule created",
				logx.Field("role", rule.RoleCode),
				logx.Field("role_name", rule.RoleName),
				logx.Field("tenant_id", tenantID),
				logx.Field("data_scope", rule.DataScope),
				logx.Field("custom_depts", rule.CustomDepts))
		}
	}

	// 🔥 通过Redis发送更新通知，触发所有服务实例重新加载策略
	updateMsg := fmt.Sprintf("UpdatePolicy:tenant_%d:data_perm", tenantID)
	err = l.svcCtx.Redis.Publish(ctx, "casbin_watcher", updateMsg).Err()
	if err != nil {
		logx.Errorw("failed to publish data perm policy update notification", logx.Field("error", err.Error()))
		// 不返回错误，因为策略已经写入数据库，只是通知失败
	} else {
		logx.Info("✅ Published data permission policy update notification to Redis")
	}

	return nil
}
