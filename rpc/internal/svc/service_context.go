package svc

import (
	"github.com/redis/go-redis/v9"
	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/internal/config"
	casbinMgr "github.com/coder-lulu/newbee-core/rpc/internal/casbin"
	"github.com/coder-lulu/newbee-core/rpc/internal/oauth"
	oauthSvc "github.com/coder-lulu/newbee-core/rpc/internal/svc/oauth"
	"github.com/coder-lulu/newbee-core/rpc/internal/encryption"

	"github.com/coder-lulu/newbee-common/v2/orm/ent/hooks"
	"github.com/zeromicro/go-zero/core/logx"

	_ "github.com/coder-lulu/newbee-core/rpc/ent/runtime"
)

type ServiceContext struct {
	Config          config.Config
	DB              *ent.Client
	Redis           redis.UniversalClient
	OAuthManager    *oauth.OAuthManager
	StateManager    *oauthSvc.StateManager
	PKCEManager     *oauthSvc.PKCEManager
	EnforcerManager *casbinMgr.EnforcerManager
	// 🔥 新增文档要求的Casbin组件
	PolicyManager     *casbinMgr.PolicyManager     // 策略管理器
	PermissionChecker *casbinMgr.PermissionChecker // 权限检查器
	// 🔐 OAuth Provider加密服务
	EncryptionService *encryption.ProviderEncryptionService
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := ent.NewClient(
		ent.Log(logx.Error), // logger
		ent.Driver(c.DatabaseConf.NewNoCacheDriver()),
	)

	// 🎯 使用统一Hook系统 - 一键设置租户和部门Hook
	// 配置Core服务的租户过滤规则 - 添加Core特有的系统表
	hooks.AddExcludedTable("sys_apis")        // API接口表是系统级数据
	hooks.AddExcludedTable("sys_audit_logs")  // 审计日志表是系统级数据
	hooks.AddExcludedTable("oauth_providers")  // OAuth提供商表是系统级数据

	// 一键设置：初始化配置 + 注册所有hooks (租户Hook + 部门Hook)
	if err := hooks.QuickSetup(db); err != nil {
		logx.Errorw("Failed to setup unified hooks", logx.Field("error", err.Error()))
		panic("统一Hook初始化失败: " + err.Error())
	}
	logx.Infow("✅ Core service: Unified hooks initialized successfully")

	// RPC服务层不注册数据权限拦截器
	// 数据权限控制应该在API层通过中间件处理，RPC层作为数据访问层不承担权限职责
	// 这样可以保持清晰的层次分离，避免跨服务的上下文传递问题

	rds := c.RedisConf.MustNewUniversalRedis()

	// Initialize OAuth Manager
	oauthManagerConfig := oauth.DefaultOAuthManagerConfig()
	oauth.InitGlobalOAuthManager(db, oauthManagerConfig)
	oauthManager := oauth.GetGlobalOAuthManager()

	// Initialize State Manager with default secret key (should be configurable)
	secretKey := []byte("default-secret-key-should-be-configurable") // TODO: Move to config
	oauthSvc.InitGlobalStateManager(rds, secretKey)
	stateManager := oauthSvc.GetGlobalStateManager()

	// Initialize PKCE Manager
	oauthSvc.InitGlobalPKCEManager(rds)
	pkceManager := oauthSvc.GetGlobalPKCEManager()

	// Initialize Casbin components
	enforcerManager := casbinMgr.NewEnforcerManager(db, rds, logx.WithContext(nil))
	
	// 🔥 初始化策略管理器
	policyManager := casbinMgr.NewPolicyManager(db, rds, enforcerManager, logx.WithContext(nil))
	
	// 🔥 初始化权限检查器
	permissionChecker := casbinMgr.NewPermissionChecker(db, rds, enforcerManager, policyManager, logx.WithContext(nil))

	// 🔐 初始化Provider加密服务
	encryption.InitGlobalEncryption()
	encMgr := encryption.GetGlobalEncryptionManager()
	
	// 配置加密密钥
	encryptionKey := c.EncryptionKey
	if encryptionKey == "" {
		// 使用默认密钥(生产环境必须配置)
		encryptionKey = "default-32-byte-encryption-key"
		logx.Infow("Using default encryption key - MUST configure in production!")
	}
	
	// 添加加密密钥(32字节用于AES-256-GCM)
	keyBytes := []byte(encryptionKey)
	if len(keyBytes) < 32 {
		// 如果密钥不足32字节,填充到32字节
		padded := make([]byte, 32)
		copy(padded, keyBytes)
		keyBytes = padded
	} else if len(keyBytes) > 32 {
		// 如果超过32字节,截取前32字节
		keyBytes = keyBytes[:32]
	}
	
	if err := encMgr.AddKey("prod-2024", keyBytes, encryption.AlgorithmAES256GCM); err != nil {
		logx.Errorw("Failed to add encryption key", logx.Field("error", err))
	}
	if err := encMgr.SetActiveKey("prod-2024"); err != nil {
		logx.Errorw("Failed to set active encryption key", logx.Field("error", err))
	}
	
	encryptionService := encryption.GetGlobalProviderEncryptionService()

	return &ServiceContext{
		Config:            c,
		DB:                db,
		Redis:             rds,
		OAuthManager:      oauthManager,
		StateManager:      stateManager,
		PKCEManager:       pkceManager,
		EnforcerManager:   enforcerManager,
		PolicyManager:     policyManager,
		PermissionChecker: permissionChecker,
		EncryptionService: encryptionService,
	}
}
