// Copyright 2024 The NewBee Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");

package svc

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	commoncasbin "github.com/coder-lulu/newbee-common/casbin"
	commonadapter "github.com/coder-lulu/newbee-common/casbin/adapter"
	"github.com/coder-lulu/newbee-common/i18n"
	"github.com/coder-lulu/newbee-common/middleware/integration"
	"github.com/coder-lulu/newbee-common/middleware/keys"
	"github.com/coder-lulu/newbee-common/orm/ent/hooks"
	"github.com/coder-lulu/newbee-common/utils/captcha"
	apicasbin "github.com/coder-lulu/newbee-core/api/internal/casbin"
	"github.com/coder-lulu/newbee-core/api/internal/config"
	i18n2 "github.com/coder-lulu/newbee-core/api/internal/i18n"
	"github.com/coder-lulu/newbee-core/rpc/coreclient"
	"github.com/coder-lulu/newbee-core/rpc/types/core"
	"github.com/mojocn/base64Captcha"
	"github.com/redis/go-redis/v9"
	"github.com/suyuan32/simple-admin-job/jobclient"
	"github.com/suyuan32/simple-admin-message-center/mcmsclient"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext 统一的服务上下文
type ServiceContext struct {
	Config config.Config

	// 核心业务依赖
	ContextManager *keys.ContextManager
	CoreRpc        coreclient.Core
	JobRpc         jobclient.Job
	McmsRpc        mcmsclient.Mcms
	Redis          redis.UniversalClient
	Casbin         *casbin.Enforcer
	Trans          *i18n.Translator
	Captcha        *base64Captcha.Captcha

	// 统一中间件链
	ManagedMiddlewareChain []rest.Middleware

	// 统一集成结果
	IntegrationResult *integration.Result

	// RPC健康状态监控
	coreRpcHealthy int64 // 1 for healthy, 0 for unhealthy (atomic operations)
}

// GetCoreRpcClient 实现audit.AuditSvcProvider接口，带降级机制
func (svc *ServiceContext) GetCoreRpcClient() interface{} {
	if atomic.LoadInt64(&svc.coreRpcHealthy) == 0 {
		logx.Error("Core RPC is unhealthy, returning disabled client")
		return &DisabledCoreRpcClient{}
	}
	return svc.CoreRpc
}

// GetCasbinEnforcer implements permission.EnforcerProvider implicitly.
// It allows the RBAC middleware plugin to fetch the Casbin enforcer from service context.
func (svc *ServiceContext) GetCasbinEnforcer() interface{} { return svc.Casbin }

// SetCoreRpcHealthy 设置Core RPC健康状态
func (svc *ServiceContext) SetCoreRpcHealthy(healthy bool) {
	if healthy {
		atomic.StoreInt64(&svc.coreRpcHealthy, 1)
	} else {
		atomic.StoreInt64(&svc.coreRpcHealthy, 0)
	}
}

// IsCoreRpcHealthy 检查Core RPC是否健康
func (svc *ServiceContext) IsCoreRpcHealthy() bool {
	return atomic.LoadInt64(&svc.coreRpcHealthy) == 1
}

// DisabledCoreRpcClient 禁用状态的Core RPC客户端（降级实现）
type DisabledCoreRpcClient struct{}

// CreateAuditLog 禁用状态下的审计日志创建（静默失败）
func (d *DisabledCoreRpcClient) CreateAuditLog(ctx context.Context, info *core.AuditLogInfo) (*core.BaseResp, error) {
	logx.Error("Core RPC is disabled, audit log creation silently failed")
	// 返回成功响应，避免影响主业务流程
	return &core.BaseResp{Msg: "Core RPC disabled"}, nil
}

// GetApiList 禁用状态下的API列表获取（返回空列表）
func (d *DisabledCoreRpcClient) GetApiList(ctx context.Context, req *core.ApiListReq) (*core.ApiListResp, error) {
	logx.Error("Core RPC is disabled, returning empty API list")
	return &core.ApiListResp{
		Total: 0,
		Data:  []*core.ApiInfo{},
	}, nil
}

// 实现其他必要的方法，根据具体使用情况添加
// 这里只实现了审计相关的关键方法

// RpcApiResourceProvider implements framework.ApiResourceProvider interface
// 使用缓存机制优化性能，避免每次请求都查询数据库
type RpcApiResourceProvider struct {
	coreRpc     coreclient.Core
	cache       map[string]string // key: "METHOD:PATH", value: resource name
	mu          sync.RWMutex
	lastRefresh time.Time
	cacheTTL    time.Duration
}

// NewRpcApiResourceProvider creates a new API resource provider with cache
func NewRpcApiResourceProvider(coreRpc coreclient.Core) *RpcApiResourceProvider {
	return &RpcApiResourceProvider{
		coreRpc:     coreRpc,
		cache:       make(map[string]string),
		cacheTTL:    5 * time.Minute, // 缓存5分钟
		lastRefresh: time.Time{},     // 零值，强制首次加载
	}
}

// GetApiResourceName gets resource name from database with cache optimization
func (p *RpcApiResourceProvider) GetApiResourceName(ctx context.Context, method, path string) (string, error) {
	cacheKey := fmt.Sprintf("%s:%s", method, path)

	logx.WithContext(ctx).Infow("Getting API resource name",
		logx.Field("method", method),
		logx.Field("path", path),
		logx.Field("cacheKey", cacheKey))

	// 检查缓存是否需要刷新
	p.mu.RLock()
	needRefresh := time.Since(p.lastRefresh) > p.cacheTTL || len(p.cache) == 0
	cacheSize := len(p.cache)
	p.mu.RUnlock()

	logx.WithContext(ctx).Infow("Cache status check",
		logx.Field("needRefresh", needRefresh),
		logx.Field("cacheSize", cacheSize),
		logx.Field("lastRefresh", p.lastRefresh))

	if needRefresh {
		// 需要刷新缓存
		if err := p.refreshCache(ctx); err != nil {
			logx.WithContext(ctx).Errorw("Failed to refresh API resource cache",
				logx.Field("error", err))
			// 刷新失败，尝试使用旧缓存或返回错误
			p.mu.RLock()
			if resourceName, exists := p.cache[cacheKey]; exists {
				p.mu.RUnlock()
				logx.WithContext(ctx).Infow("Using old cache after refresh failure",
					logx.Field("resourceName", resourceName))
				return resourceName, nil
			}
			p.mu.RUnlock()
			return "", fmt.Errorf("cache refresh failed and no cached value found: %w", err)
		}
	}

	// 从缓存中获取
	p.mu.RLock()
	defer p.mu.RUnlock()

	if resourceName, exists := p.cache[cacheKey]; exists {
		logx.WithContext(ctx).Infow("Found resource name in cache",
			logx.Field("resourceName", resourceName))
		return resourceName, nil
	}

	// 输出部分缓存内容用于调试
	logx.WithContext(ctx).Errorw("Resource name not found in cache",
		logx.Field("cacheKey", cacheKey),
		logx.Field("cacheSize", len(p.cache)))

	// 输出前5个缓存键用于调试
	count := 0
	for key := range p.cache {
		if count >= 5 {
			break
		}
		logx.WithContext(ctx).Infow("Sample cache key", logx.Field("key", key))
		count++
	}

	return "", fmt.Errorf("resource name not found for %s %s", method, path)
}

// refreshCache 刷新缓存，从数据库加载所有API信息
func (p *RpcApiResourceProvider) refreshCache(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 双重检查，避免并发刷新
	if time.Since(p.lastRefresh) <= p.cacheTTL && len(p.cache) > 0 {
		return nil
	}

	// 使用系统上下文查询所有API（不受租户限制）
	// API definitions are system-level data that should be accessible globally
	systemCtx := hooks.NewSystemContext(ctx)

	logx.WithContext(ctx).Infow("Attempting to refresh API cache via RPC",
		logx.Field("rpcClient", "coreRpc"))

	resp, err := p.coreRpc.GetApiList(systemCtx, &core.ApiListReq{
		Page:     1,
		PageSize: 10000, // 假设不会超过10000个API
	})

	if err != nil {
		// 记录详细的错误信息，但不要让缓存刷新失败，继续使用fallback
		logx.WithContext(ctx).Errorw("Failed to get API list from RPC, will use fallback naming",
			logx.Field("error", err.Error()),
			logx.Field("errorType", fmt.Sprintf("%T", err)))

		// 创建一个基本的缓存结构，避免频繁重试
		p.cache = make(map[string]string)
		p.lastRefresh = time.Now()
		return fmt.Errorf("failed to get API list from RPC: %w", err)
	}

	// 清空旧缓存并构建新缓存
	newCache := make(map[string]string)

	if resp.Data != nil {
		for _, api := range resp.Data {
			if api.Method != nil && api.Path != nil {
				method := *api.Method
				path := p.normalizePath(*api.Path)
				cacheKey := fmt.Sprintf("%s:%s", method, path)

				// 使用Description字段作为资源名称
				resourceName := ""
				if api.Description != nil && *api.Description != "" {
					resourceName = *api.Description
				} else {
					// 如果Description为空，根据路径生成通用名称
					resourceName = p.generateFallbackName(method, path)
				}

				newCache[cacheKey] = resourceName
			}
		}
	}

	p.cache = newCache
	p.lastRefresh = time.Now()

	logx.WithContext(ctx).Infow("API resource cache refreshed",
		logx.Field("count", len(p.cache)))

	return nil
}

// normalizePath 标准化路径
func (p *RpcApiResourceProvider) normalizePath(path string) string {
	// 移除查询参数
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}

	// 确保以/开头
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// 移除结尾的/
	path = strings.TrimSuffix(path, "/")

	return path
}

// generateFallbackName 生成回退的资源名称
func (p *RpcApiResourceProvider) generateFallbackName(method, path string) string {
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	if len(pathParts) >= 2 {
		resource := pathParts[0]
		action := pathParts[1]

		// 资源类型映射
		resourceTypeMap := map[string]string{
			"user":          "用户",
			"role":          "角色",
			"menu":          "菜单",
			"api":           "接口",
			"department":    "部门",
			"position":      "职位",
			"dictionary":    "字典",
			"tenant":        "租户",
			"token":         "令牌",
			"task":          "任务",
			"captcha":       "验证码",
			"oauth":         "OAuth",
			"sms":           "短信",
			"email":         "邮件",
			"configuration": "配置",
			"audit-log":     "审计日志",
			"authority":     "权限",
		}

		// 操作类型映射
		actionMap := map[string]string{
			"create":   "创建",
			"update":   "更新",
			"delete":   "删除",
			"list":     "列表",
			"detail":   "详情",
			"login":    "登录",
			"logout":   "退出",
			"register": "注册",
			"send":     "发送",
		}

		resourceType := resourceTypeMap[resource]
		if resourceType == "" {
			resourceType = resource
		}

		actionType := actionMap[action]
		if actionType == "" {
			// 基于HTTP方法推断操作类型
			switch method {
			case "POST":
				actionType = "操作"
			case "GET":
				actionType = "查询"
			case "PUT", "PATCH":
				actionType = "更新"
			case "DELETE":
				actionType = "删除"
			default:
				actionType = "访问"
			}
		}

		return actionType + resourceType
	}

	// 单级路径的处理
	return "访问" + path
}

// NewServiceContext 创建服务上下文
func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化Redis - 使用go-zero框架方法，自动根据配置选择单机/哨兵/集群模式
	rds := c.RedisConf.MustNewUniversalRedis()

	// 初始化核心RPC客户端
	logx.Infof("Initializing Core RPC client with config: %+v", c.CoreRpc)

	// 创建RPC客户端，使用拦截器添加SystemContext和租户上下文传递
	rpcClient, err := zrpc.NewClient(c.CoreRpc,
		zrpc.WithUnaryClientInterceptor(hooks.SystemContextClientInterceptor()),
		zrpc.WithStreamClientInterceptor(hooks.SystemContextStreamClientInterceptor()))
	if err != nil {
		logx.Errorf("Failed to create Core RPC client: %v", err)
		panic(fmt.Sprintf("Core RPC client initialization failed: %v", err))
	}

	coreClient := coreclient.NewCore(rpcClient)

	// 初始化验证码
	captchaInstance := captcha.MustNewOriginalRedisCaptcha(c.Captcha, rds)

	// 🔥 初始化Casbin - 使用EntAdapter通过RPC查询规则
	logx.Info("Initializing Casbin with EntAdapter and RPC querier for all tenants")

	// 1. 创建RPC查询器
	rpcQuerier := apicasbin.NewRpcCasbinRuleQuerier(coreClient)

	// 2. 创建EntAdapter
	// 🔥 使用SystemContext绕过租户隔离Hook，加载所有租户的Casbin规则
	// 架构设计：
	// - RBAC with Domains模型：一个enforcer包含所有租户的规则
	// - 运行时通过domain参数（租户ID）实现租户隔离
	// - Casbin会自动过滤匹配 r.dom == p.dom 的规则
	// - 相比每租户一个enforcer，内存占用更小，性能更好
	systemCtx := hooks.NewSystemContext(context.Background())
	adapter := commonadapter.NewEntAdapter(rpcQuerier, systemCtx)

	// 3. 创建Casbin模型
	modelText := commoncasbin.GetDefaultRBACWithDomainsModel()
	m, err := model.NewModelFromString(modelText)
	if err != nil {
		logx.Errorf("Failed to create Casbin model: %v", err)
		panic(fmt.Sprintf("Casbin model creation failed: %v", err))
	}

	// 4. 创建Enforcer（使用普通Enforcer，Redis Watcher会提供同步功能）
	cbn, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		logx.Errorf("Failed to create Casbin enforcer: %v", err)
		panic(fmt.Sprintf("Casbin enforcer creation failed: %v", err))
	}

	// 5. 加载策略
	err = cbn.LoadPolicy()
	if err != nil {
		logx.Errorf("Failed to load Casbin policy: %v", err)
		panic(fmt.Sprintf("Casbin policy loading failed: %v", err))
	}

	// 6. 添加Redis Watcher（用于策略同步）
	// 🔥 使用自定义回调，调用LoadPolicy()重新加载所有租户的策略
	// ⚠️ 不使用DefaultUpdateCallback（它会调用SelfAddPolicies，需要BatchAdapter接口）
	w := c.CasbinConf.MustNewOriginalRedisWatcher(c.RedisConf, func(data string) {
		logx.Infow("📨 Received Casbin policy update notification from Redis",
			logx.Field("message", data))

		// 直接重新加载所有策略（包括新租户的策略）
		// LoadPolicy()会通过EntAdapter + SystemContext查询所有租户的规则
		err := cbn.LoadPolicy()
		if err != nil {
			logx.Errorw("❌ Failed to reload Casbin policy after Redis notification",
				logx.Field("error", err.Error()),
				logx.Field("message", data))
		} else {
			logx.Infow("✅ Successfully reloaded Casbin policy after Redis notification",
				logx.Field("message", data))
		}
	})
	err = cbn.SetWatcher(w)
	if err != nil {
		logx.Errorf("Failed to set Casbin watcher: %v", err)
		panic(fmt.Sprintf("Casbin watcher setup failed: %v", err))
	}

	logx.Info("✅ Casbin initialized with EntAdapter, connected to sys_casbin_rules table via RPC")
	logx.Info("✅ Redis Watcher configured with custom callback (LoadPolicy on update)")

	trans := i18n.NewTranslator(c.I18nConf, i18n2.LocaleFS)

	// ===========================================
	// 🎉 新版统一中间件框架集成 - 使用统一的 integration.Setup() API
	// ===========================================

	// 1. 创建服务上下文实例（需要先创建以便传递给审计插件）
	svcCtx := &ServiceContext{
		Config:         c,
		CoreRpc:        coreClient,
		McmsRpc:        mcmsclient.NewMcms(zrpc.NewClientIfEnable(c.McmsRpc)),
		JobRpc:         jobclient.NewJob(zrpc.NewClientIfEnable(c.JobRpc)),
		Redis:          rds,
		Casbin:         cbn,
		Trans:          trans,
		Captcha:        captchaInstance,
		coreRpcHealthy: 1, // 默认假设健康
	}

	// 3. 获取JWT密钥 - 优先使用Middleware配置，向后兼容Auth配置
	jwtSecret := c.Middleware.Auth.AccessSecret
	// 确保routes.go能正常工作
	svcCtx.Config.Auth.AccessSecret = jwtSecret

	// 4. 准备审计配置 - 核心服务需要自己处理审计日志
	middlewareConfig := c.Middleware
	if middlewareConfig.Audit != nil && middlewareConfig.Audit.Enabled {
		// 对于Core服务，使用NoOp写入器避免循环调用
		middlewareConfig.Audit.WriterType = "custom"
	}

	// 5. 🎯 使用标准化的统一集成API - 遵循框架最佳实践
	// 创建高性能审计写入器，使用适配器消除反射调用
	highPerfAuditWriter := NewHighPerformanceCoreAuditWriter(coreClient, svcCtx.IsCoreRpcHealthy)

	result, err := integration.Setup(&integration.Config{
		Redis:                  rds,
		JWTSecret:              jwtSecret,
		Mode:                   integration.Production,
		ApiResourceProvider:    NewRpcApiResourceProvider(coreClient),
		AuditWriter:            highPerfAuditWriter, // 使用高性能审计写入器
		TenantInfoProvider:     NewRpcTenantInfoProvider(coreClient),
		RbacProvider:           svcCtx,            // 提供 Casbin enforcer（仅当启用 permission 插件时生效）
		DataPermCasbinProvider: rpcQuerier,        // 🔥 提供数据权限Casbin查询器（UnifiedDataPermPlugin必需）
		Middleware:             &middlewareConfig, // 使用调整后的配置
	})
	if err != nil {
		panic("统一中间件集成失败: " + err.Error())
	}

	// 5. 应用集成结果到服务上下文
	svcCtx.ContextManager = result.ContextManager
	svcCtx.ManagedMiddlewareChain = result.Middlewares
	svcCtx.IntegrationResult = result

	return svcCtx
}
