package base

import (
	"context"

	"github.com/coder-lulu/newbee-common/v2/orm/ent/hooks"
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/coder-lulu/newbee-core/rpc/ent"
)

// insert initial api data
// 🔥 自动生成于: 2025-10-08
// 🔥 从core/api/desc目录的.api文件解析生成
// 🔥 总接口数: 165个 (Core: 133, Job: 10, MCMS: 22)
func (l *InitDatabaseLogic) insertApiData(ctx context.Context) error {
	ctxWithTenant := hooks.SetTenantIDToContext(context.Background(), 1)
	var apis []*ent.APICreate

	// ==================== CORE Service ====================
	// Api
	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/api").
		SetDescription("Get API by ID | 通过ID获取API").
		SetAPIGroup("api").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/api/create").
		SetDescription("Create API information | 创建API").
		SetAPIGroup("api").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/api/delete").
		SetDescription("Delete API information | 删除API信息").
		SetAPIGroup("api").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/api/list").
		SetDescription("Get API list | 获取API列表").
		SetAPIGroup("api").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/api/update").
		SetDescription("Update API information | 创建API").
		SetAPIGroup("api").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	// Auditlog
	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/audit-log").
		SetDescription("Get audit log by ID | 通过ID获取审计日志").
		SetAPIGroup("auditlog").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/audit-log/delete").
		SetDescription("Delete audit logs | 删除审计日志").
		SetAPIGroup("auditlog").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/audit-log/list").
		SetDescription("Get audit log list | 获取审计日志列表").
		SetAPIGroup("auditlog").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/audit-log/stats").
		SetDescription("Get audit log statistics | 获取审计日志统计").
		SetAPIGroup("auditlog").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	// Auth
	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/auth/tenant/list").
		SetDescription("Get public tenant list | 获取公开租户列表（无需认证）").
		SetAPIGroup("auth").
		SetMethod("GET").
		SetIsRequired(true).
		SetTenantID(1),
	)

	// Authority
	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/authority/api/create_or_update").
		SetDescription("Create or update API authorization information | 创建或更新API权限").
		SetAPIGroup("authority").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/authority/api/role").
		SetDescription("Get role's API authorization list | 获取角色api权限列表").
		SetAPIGroup("authority").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/authority/menu/create_or_update").
		SetDescription("Create or update menu authorization information | 创建或更新菜单权限").
		SetAPIGroup("authority").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/authority/menu/role").
		SetDescription("Get role's menu authorization list | 获取角色菜单权限列表").
		SetAPIGroup("authority").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	// Captcha
	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/captcha").
		SetDescription("Get captcha | 获取验证码").
		SetAPIGroup("captcha").
		SetMethod("GET").
		SetIsRequired(true).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/captcha/email").
		SetDescription("Get Email Captcha | 获取邮箱验证码").
		SetAPIGroup("captcha").
		SetMethod("POST").
		SetIsRequired(true).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/captcha/sms").
		SetDescription("Get SMS Captcha | 获取短信验证码").
		SetAPIGroup("captcha").
		SetMethod("POST").
		SetIsRequired(true).
		SetTenantID(1),
	)

	// Casbin
	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/casbin/permission/batch/check").
		SetDescription("Batch check permission | 批量权限检查").
		SetAPIGroup("casbin").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/casbin/permission/check").
		SetDescription("Check permission | 权限检查").
		SetAPIGroup("casbin").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/casbin/permission/summary").
		SetDescription("Get user permission summary | 获取用户权限摘要").
		SetAPIGroup("casbin").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/casbin/rules").
		SetDescription("Create Casbin rule | 创建权限规则").
		SetAPIGroup("casbin").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/casbin/rules").
		SetDescription("Update Casbin rule | 更新权限规则").
		SetAPIGroup("casbin").
		SetMethod("PUT").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/casbin/rules").
		SetDescription("Delete Casbin rule | 删除权限规则").
		SetAPIGroup("casbin").
		SetMethod("DELETE").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/casbin/rules/batch/create").
		SetDescription("Batch create Casbin rules | 批量创建权限规则").
		SetAPIGroup("casbin").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/casbin/rules/batch/delete").
		SetDescription("Batch delete Casbin rules | 批量删除权限规则").
		SetAPIGroup("casbin").
		SetMethod("DELETE").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/casbin/rules/batch/update").
		SetDescription("Batch update Casbin rules | 批量更新权限规则").
		SetAPIGroup("casbin").
		SetMethod("PUT").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/casbin/rules/detail").
		SetDescription("Get Casbin rule by ID | 根据ID获取权限规则").
		SetAPIGroup("casbin").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/casbin/rules/list").
		SetDescription("Get Casbin rule list | 获取权限规则列表").
		SetAPIGroup("casbin").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/casbin/rules/validate").
		SetDescription("Validate Casbin rule | 验证权限规则").
		SetAPIGroup("casbin").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/casbin/system/cache/refresh").
		SetDescription("Refresh Casbin cache | 刷新权限缓存").
		SetAPIGroup("casbin").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/casbin/system/sync").
		SetDescription("Sync Casbin rules | 同步权限规则").
		SetAPIGroup("casbin").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	// Configuration
	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/configuration").
		SetDescription("Get configuration by ID | 通过ID获取参数配置").
		SetAPIGroup("configuration").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/configuration/create").
		SetDescription("Create configuration information | 创建参数配置").
		SetAPIGroup("configuration").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/configuration/delete").
		SetDescription("Delete configuration information | 删除参数配置信息").
		SetAPIGroup("configuration").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/configuration/list").
		SetDescription("Get configuration list | 获取参数配置列表").
		SetAPIGroup("configuration").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/configuration/refreshCache").
		SetDescription("Refresh configuration cache | 刷新参数配置缓存").
		SetAPIGroup("configuration").
		SetMethod("DELETE").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/configuration/update").
		SetDescription("Update configuration information | 更新参数配置").
		SetAPIGroup("configuration").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	// Department
	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/department").
		SetDescription("Get Department by ID | 通过ID获取部门").
		SetAPIGroup("department").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/department/create").
		SetDescription("Create department information | 创建部门").
		SetAPIGroup("department").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/department/delete").
		SetDescription("Delete department information | 删除部门信息").
		SetAPIGroup("department").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/department/list").
		SetDescription("Get department list | 获取部门列表").
		SetAPIGroup("department").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/department/update").
		SetDescription("Update department information | 更新部门").
		SetAPIGroup("department").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	// Dictionary
	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/dictionary").
		SetDescription("Get Dictionary by ID | 通过ID获取字典").
		SetAPIGroup("dictionary").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/dictionary/create").
		SetDescription("Create dictionary information | 创建字典").
		SetAPIGroup("dictionary").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/dictionary/delete").
		SetDescription("Delete dictionary information | 删除字典信息").
		SetAPIGroup("dictionary").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/dictionary/list").
		SetDescription("Get dictionary list | 获取字典列表").
		SetAPIGroup("dictionary").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/dictionary/update").
		SetDescription("Update dictionary information | 更新字典").
		SetAPIGroup("dictionary").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	// Dictionarydetail
	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/dict/:name").
		SetDescription("Get dictionary detail by dictionary name | 通过字典名称获取字典内容").
		SetAPIGroup("dictionarydetail").
		SetMethod("GET").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/dictionary_detail").
		SetDescription("Get dictionary detail by ID | 通过ID获取字典键值").
		SetAPIGroup("dictionarydetail").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/dictionary_detail/create").
		SetDescription("Create dictionary detail information | 创建字典键值").
		SetAPIGroup("dictionarydetail").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/dictionary_detail/delete").
		SetDescription("Delete dictionary detail information | 删除字典键值信息").
		SetAPIGroup("dictionarydetail").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/dictionary_detail/list").
		SetDescription("Get dictionary detail list | 获取字典键值列表").
		SetAPIGroup("dictionarydetail").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/dictionary_detail/update").
		SetDescription("Update dictionary detail information | 更新字典键值").
		SetAPIGroup("dictionarydetail").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	// Menu
	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/menu/create").
		SetDescription("Create menu information | 创建菜单").
		SetAPIGroup("menu").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/menu/delete").
		SetDescription("Delete menu information | 删除菜单信息").
		SetAPIGroup("menu").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/menu/detail").
		SetDescription("Get menu detail | 获取菜单信息").
		SetAPIGroup("menu").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/menu/list").
		SetDescription("Get menu list | 获取菜单列表").
		SetAPIGroup("menu").
		SetMethod("GET").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/menu/role/list").
		SetDescription("Get menu list by role | 获取菜单列表").
		SetAPIGroup("menu").
		SetMethod("GET").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/menu/update").
		SetDescription("Update menu information | 更新菜单").
		SetAPIGroup("menu").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	// Oauthaccount
	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/oauth/accounts").
		SetDescription("Get user OAuth accounts | 获取用户OAuth账户").
		SetAPIGroup("oauthaccount").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/oauth/bind").
		SetDescription("Bind OAuth account | 绑定OAuth账户").
		SetAPIGroup("oauthaccount").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/oauth/unbind").
		SetDescription("Unbind OAuth account | 解绑OAuth账户").
		SetAPIGroup("oauthaccount").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	// Oauthprovider
	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/oauth/callback").
		SetDescription("Enhanced OAuth callback with parameters | 增强的OAuth回调处理").
		SetAPIGroup("oauthprovider").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/oauth/login").
		SetDescription("Oauth log in | Oauth 登录").
		SetAPIGroup("oauthprovider").
		SetMethod("POST").
		SetIsRequired(true).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/oauth/login/callback").
		SetDescription("Oauth log in callback route | Oauth 登录返回调接口").
		SetAPIGroup("oauthprovider").
		SetMethod("GET").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/oauth/providers").
		SetDescription("Get available OAuth providers for users | 获取用户可用的OAuth提供商").
		SetAPIGroup("oauthprovider").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/oauth/statistics").
		SetDescription("Get OAuth statistics | 获取OAuth统计数据").
		SetAPIGroup("oauthprovider").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/oauth_account").
		SetDescription("Get oauth account by ID | 通过ID获取OAuth账户").
		SetAPIGroup("oauthprovider").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/oauth_account/create").
		SetDescription("Create oauth account | 创建OAuth账户").
		SetAPIGroup("oauthprovider").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/oauth_account/delete").
		SetDescription("Delete oauth account | 删除OAuth账户").
		SetAPIGroup("oauthprovider").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/oauth_account/list").
		SetDescription("Get oauth account list | 获取OAuth账户列表").
		SetAPIGroup("oauthprovider").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/oauth_account/update").
		SetDescription("Update oauth account | 更新OAuth账户").
		SetAPIGroup("oauthprovider").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/oauth_provider").
		SetDescription("Get oauth provider by ID | 通过ID获取第三方").
		SetAPIGroup("oauthprovider").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/oauth_provider/create").
		SetDescription("Create oauth provider information | 创建第三方").
		SetAPIGroup("oauthprovider").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/oauth_provider/delete").
		SetDescription("Delete oauth provider information | 删除第三方信息").
		SetAPIGroup("oauthprovider").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/oauth_provider/list").
		SetDescription("Get oauth provider list | 获取第三方列表").
		SetAPIGroup("oauthprovider").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/oauth_provider/test").
		SetDescription("Test oauth provider connection | 测试第三方提供商连接").
		SetAPIGroup("oauthprovider").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/oauth_provider/update").
		SetDescription("Update oauth provider information | 更新第三方").
		SetAPIGroup("oauthprovider").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	// Position
	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/position").
		SetDescription("Get position by ID | 通过ID获取职位").
		SetAPIGroup("position").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/position/create").
		SetDescription("Create position information | 创建职位").
		SetAPIGroup("position").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/position/delete").
		SetDescription("Delete position information | 删除职位信息").
		SetAPIGroup("position").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/position/list").
		SetDescription("Get position list | 获取职位列表").
		SetAPIGroup("position").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/position/update").
		SetDescription("Update position information | 更新职位").
		SetAPIGroup("position").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	// Publicapi
	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/configuration/system/list").
		SetDescription("Get public system configuration list | 获取公开系统参数列表").
		SetAPIGroup("publicapi").
		SetMethod("GET").
		SetIsRequired(false).
		SetTenantID(1),
	)

	// Publicuser
	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/user/login").
		SetDescription("Log in | 登录").
		SetAPIGroup("publicuser").
		SetMethod("POST").
		SetIsRequired(true).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/user/login_by_email").
		SetDescription("Log in by email | 邮箱登录").
		SetAPIGroup("publicuser").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/user/login_by_sms").
		SetDescription("Log in by SMS | 短信登录").
		SetAPIGroup("publicuser").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/user/logout").
		SetDescription("Log out | 退出登陆 (无需认证)").
		SetAPIGroup("publicuser").
		SetMethod("GET").
		SetIsRequired(true).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/user/register").
		SetDescription("Register | 注册").
		SetAPIGroup("publicuser").
		SetMethod("POST").
		SetIsRequired(true).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/user/register_by_email").
		SetDescription("Register by Email | 邮箱注册").
		SetAPIGroup("publicuser").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/user/register_by_sms").
		SetDescription("Register by SMS | 短信注册").
		SetAPIGroup("publicuser").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/user/reset_password_by_email").
		SetDescription("Reset password by Email | 通过邮箱重置密码").
		SetAPIGroup("publicuser").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/user/reset_password_by_sms").
		SetDescription("Reset password by Sms | 通过短信重置密码").
		SetAPIGroup("publicuser").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	// Role
	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/role").
		SetDescription("Get Role by ID | 通过ID获取角色").
		SetAPIGroup("role").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/role/cancelAuthUser").
		SetDescription("Cancel User Role Auth | 取消用户角色授权").
		SetAPIGroup("role").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/role/changeRoleStatus").
		SetDescription("Change role Status | 更新角色状态").
		SetAPIGroup("role").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/role/create").
		SetDescription("Create role information | 创建角色").
		SetAPIGroup("role").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/role/dataScope").
		SetDescription("Assign Role DataScope | 授权数据权限").
		SetAPIGroup("role").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/role/delete").
		SetDescription("Delete role information | 删除角色信息").
		SetAPIGroup("role").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/role/list").
		SetDescription("Get role list | 获取角色列表").
		SetAPIGroup("role").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/role/multiAuthUser").
		SetDescription("Auth User Role | 用户角色授权").
		SetAPIGroup("role").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/role/update").
		SetDescription("Update role information | 更新角色").
		SetAPIGroup("role").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	// Tenant
	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/tenant").
		SetDescription("Get Tenant by ID | 通过ID获取租户").
		SetAPIGroup("tenant").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/tenant/code").
		SetDescription("Get Tenant by Code | 通过租户码获取租户").
		SetAPIGroup("tenant").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/tenant/create").
		SetDescription("Create tenant | 创建租户").
		SetAPIGroup("tenant").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/tenant/current").
		SetDescription("Get current active tenant | 获取当前激活租户").
		SetAPIGroup("tenant").
		SetMethod("GET").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/tenant/delete").
		SetDescription("Delete tenant | 删除租户").
		SetAPIGroup("tenant").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/tenant/dynamic/clear").
		SetDescription("Clear tenant switch | 清除租户切换").
		SetAPIGroup("tenant").
		SetMethod("GET").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/tenant/init").
		SetDescription("Initialize tenant | 初始化租户").
		SetAPIGroup("tenant").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/tenant/list").
		SetDescription("Get tenant list | 获取租户列表").
		SetAPIGroup("tenant").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/tenant/status").
		SetDescription("Update tenant status | 更新租户状态").
		SetAPIGroup("tenant").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/tenant/switch").
		SetDescription("Switch tenant for super admin | 超级管理员切换租户").
		SetAPIGroup("tenant").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/tenant/update").
		SetDescription("Update tenant | 更新租户").
		SetAPIGroup("tenant").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	// Token
	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/token").
		SetDescription("Get Token by ID | 通过ID获取令牌").
		SetAPIGroup("token").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/token/create").
		SetDescription("Create token information | 创建令牌").
		SetAPIGroup("token").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/token/delete").
		SetDescription("Delete token information | 删除令牌信息").
		SetAPIGroup("token").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/token/list").
		SetDescription("Get token list | 获取令牌列表").
		SetAPIGroup("token").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/token/logout").
		SetDescription("Force logging out by user UUID | 根据UUID强制用户退出").
		SetAPIGroup("token").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/token/update").
		SetDescription("Update token information | 更新令牌").
		SetAPIGroup("token").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	// User
	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/user").
		SetDescription("Get User by ID | 通过ID获取用户").
		SetAPIGroup("user").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/user/access_token").
		SetDescription("Access token | 获取短期 token").
		SetAPIGroup("user").
		SetMethod("GET").
		SetIsRequired(true).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/user/change_password").
		SetDescription("Change Password | 修改密码").
		SetAPIGroup("user").
		SetMethod("POST").
		SetIsRequired(true).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/user/create").
		SetDescription("Create user information | 创建用户").
		SetAPIGroup("user").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/user/delete").
		SetDescription("Delete user information | 删除用户信息").
		SetAPIGroup("user").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/user/info").
		SetDescription("Get user basic information | 获取用户基本信息").
		SetAPIGroup("user").
		SetMethod("GET").
		SetIsRequired(true).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/user/list").
		SetDescription("Get user list | 获取用户列表").
		SetAPIGroup("user").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/user/perm").
		SetDescription("Get user's permission code | 获取用户权限码").
		SetAPIGroup("user").
		SetMethod("GET").
		SetIsRequired(true).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/user/profile").
		SetDescription("Get user's profile | 获取用户个人信息").
		SetAPIGroup("user").
		SetMethod("GET").
		SetIsRequired(true).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/user/profile").
		SetDescription("Update user's profile | 更新用户个人信息").
		SetAPIGroup("user").
		SetMethod("POST").
		SetIsRequired(true).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/user/refresh_token").
		SetDescription("Refresh token | 获取刷新 token").
		SetAPIGroup("user").
		SetMethod("GET").
		SetIsRequired(true).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/user/resetPwd").
		SetDescription("Reset password | 管理员后台重置密码").
		SetAPIGroup("user").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/user/unallocatedList").
		SetDescription("UnallocatedUserList | 获取未授权给当前角色的用户列表").
		SetAPIGroup("user").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Core").
		SetPath("/user/update").
		SetDescription("Update user information | 更新用户").
		SetAPIGroup("user").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	// ==================== JOB Service ====================
	// Task
	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Job").
		SetPath("/task").
		SetDescription("Get task by ID | 通过ID获取定时任务").
		SetAPIGroup("task").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Job").
		SetPath("/task/create").
		SetDescription("Create task information | 创建定时任务").
		SetAPIGroup("task").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Job").
		SetPath("/task/delete").
		SetDescription("Delete task information | 删除定时任务信息").
		SetAPIGroup("task").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Job").
		SetPath("/task/list").
		SetDescription("Get task list | 获取定时任务列表").
		SetAPIGroup("task").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Job").
		SetPath("/task/update").
		SetDescription("Update task information | 更新定时任务").
		SetAPIGroup("task").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	// Tasklog
	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Job").
		SetPath("/task_log").
		SetDescription("Get task log by ID | 通过ID获取任务日志").
		SetAPIGroup("tasklog").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Job").
		SetPath("/task_log/create").
		SetDescription("Create task log information | 创建任务日志").
		SetAPIGroup("tasklog").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Job").
		SetPath("/task_log/delete").
		SetDescription("Delete task log information | 删除任务日志信息").
		SetAPIGroup("tasklog").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Job").
		SetPath("/task_log/list").
		SetDescription("Get task log list | 获取任务日志列表").
		SetAPIGroup("tasklog").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Job").
		SetPath("/task_log/update").
		SetDescription("Update task log information | 更新任务日志").
		SetAPIGroup("tasklog").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	// ==================== MCMS Service ====================
	// Emaillog
	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Mcms").
		SetPath("/email_log").
		SetDescription("Get email log by ID | 通过ID获取电子邮件日志").
		SetAPIGroup("emaillog").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Mcms").
		SetPath("/email_log/create").
		SetDescription("Create email log information | 创建电子邮件日志").
		SetAPIGroup("emaillog").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Mcms").
		SetPath("/email_log/delete").
		SetDescription("Delete email log information | 删除电子邮件日志信息").
		SetAPIGroup("emaillog").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Mcms").
		SetPath("/email_log/list").
		SetDescription("Get email log list | 获取电子邮件日志列表").
		SetAPIGroup("emaillog").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Mcms").
		SetPath("/email_log/update").
		SetDescription("Update email log information | 更新电子邮件日志").
		SetAPIGroup("emaillog").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	// Emailprovider
	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Mcms").
		SetPath("/email_provider").
		SetDescription("Get email provider by ID | 通过ID获取邮箱服务配置").
		SetAPIGroup("emailprovider").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Mcms").
		SetPath("/email_provider/create").
		SetDescription("Create email provider information | 创建邮箱服务配置").
		SetAPIGroup("emailprovider").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Mcms").
		SetPath("/email_provider/delete").
		SetDescription("Delete email provider information | 删除邮箱服务配置信息").
		SetAPIGroup("emailprovider").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Mcms").
		SetPath("/email_provider/list").
		SetDescription("Get email provider list | 获取邮箱服务配置列表").
		SetAPIGroup("emailprovider").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Mcms").
		SetPath("/email_provider/update").
		SetDescription("Update email provider information | 更新邮箱服务配置").
		SetAPIGroup("emailprovider").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	// Messagesender
	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Mcms").
		SetPath("/email/send").
		SetDescription("Send email message | 发送电子邮件").
		SetAPIGroup("messagesender").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Mcms").
		SetPath("/sms/send").
		SetDescription("Send sms message | 发送短信").
		SetAPIGroup("messagesender").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	// Smslog
	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Mcms").
		SetPath("/sms_log").
		SetDescription("Get sms log by ID | 通过ID获取短信日志").
		SetAPIGroup("smslog").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Mcms").
		SetPath("/sms_log/create").
		SetDescription("Create sms log information | 创建短信日志").
		SetAPIGroup("smslog").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Mcms").
		SetPath("/sms_log/delete").
		SetDescription("Delete sms log information | 删除短信日志信息").
		SetAPIGroup("smslog").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Mcms").
		SetPath("/sms_log/list").
		SetDescription("Get sms log list | 获取短信日志列表").
		SetAPIGroup("smslog").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Mcms").
		SetPath("/sms_log/update").
		SetDescription("Update sms log information | 更新短信日志").
		SetAPIGroup("smslog").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	// Smsprovider
	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Mcms").
		SetPath("/sms_provider").
		SetDescription("Get sms provider by ID | 通过ID获取短信配置").
		SetAPIGroup("smsprovider").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Mcms").
		SetPath("/sms_provider/create").
		SetDescription("Create sms provider information | 创建短信配置").
		SetAPIGroup("smsprovider").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Mcms").
		SetPath("/sms_provider/delete").
		SetDescription("Delete sms provider information | 删除短信配置信息").
		SetAPIGroup("smsprovider").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Mcms").
		SetPath("/sms_provider/list").
		SetDescription("Get sms provider list | 获取短信配置列表").
		SetAPIGroup("smsprovider").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	apis = append(apis, l.svcCtx.DB.API.Create().
		SetServiceName("Mcms").
		SetPath("/sms_provider/update").
		SetDescription("Update sms provider information | 更新短信配置").
		SetAPIGroup("smsprovider").
		SetMethod("POST").
		SetIsRequired(false).
		SetTenantID(1),
	)

	err := l.svcCtx.DB.API.CreateBulk(apis...).Exec(ctxWithTenant)
	if err != nil {
		logx.Errorw(err.Error())
		return errorx.NewInternalError(err.Error())
	}
	return nil
}
