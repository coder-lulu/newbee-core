package tenant

import (
	"context"
	"encoding/json"
	"time"

	"github.com/coder-lulu/newbee-core/rpc/ent/tenant"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/coder-lulu/newbee-common/v2/i18n"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/hooks"

	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTenantLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTenantLogic {
	return &CreateTenantLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateTenantLogic) CreateTenant(in *core.TenantInfo) (*core.BaseIDResp, error) {
	// 验证租户编码唯一性
	if in.Code != nil {
		checkCode, err := l.svcCtx.DB.Tenant.Query().Where(tenant.CodeEQ(*in.Code)).Exist(hooks.NewSystemContext(l.ctx))
		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}

		if checkCode {
			return nil, errorx.NewInvalidArgumentError("tenant.codeExist")
		}
	}

	// 验证租户名称唯一性
	if in.Name != nil {
		checkName, err := l.svcCtx.DB.Tenant.Query().Where(tenant.NameEQ(*in.Name)).Exist(hooks.NewSystemContext(l.ctx))
		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}

		if checkName {
			return nil, errorx.NewInvalidArgumentError("tenant.nameExist")
		}
	}

	// 构建租户创建请求
	builder := l.svcCtx.DB.Tenant.Create()

	if in.Name != nil {
		builder.SetName(*in.Name)
	}

	if in.Code != nil {
		builder.SetCode(*in.Code)
	}

	if in.Description != nil {
		builder.SetDescription(*in.Description)
	}

	if in.Status != nil {
		builder.SetStatus(uint8(*in.Status))
	} else {
		builder.SetStatus(1) // 默认状态为正常
	}

	// 处理过期时间
	if in.ExpiredAt != nil {
		builder.SetExpiredAt(time.Unix(*in.ExpiredAt, 0))
	} else {
		// 默认设置为一年后过期
		builder.SetExpiredAt(time.Now().AddDate(1, 0, 0))
	}

	// 处理配置信息
	if in.Config != nil {
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(*in.Config), &config); err != nil {
			return nil, errorx.NewInvalidArgumentError("tenant.invalidConfig")
		}
		builder.SetConfig(config)
	} else {
		// 设置默认配置
		defaultConfig := map[string]interface{}{
			"max_users":        100,
			"storage_limit_gb": 10,
			"features":         []string{"basic"},
		}
		builder.SetConfig(defaultConfig)
	}

	if in.CreatedBy != nil {
		builder.SetCreatedBy(*in.CreatedBy)
	}

	// 使用系统上下文创建租户（租户自身不受租户隔离限制）
	result, err := builder.Save(hooks.NewSystemContext(l.ctx))
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	logx.Infow("Tenant created successfully",
		logx.Field("tenant_id", result.ID),
		logx.Field("tenant_code", result.Code),
		logx.Field("tenant_name", result.Name))

	return &core.BaseIDResp{
		Id:  result.ID,
		Msg: i18n.CreateSuccess,
	}, nil
}
