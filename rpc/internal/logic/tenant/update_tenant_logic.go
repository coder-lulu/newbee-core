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

type UpdateTenantLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTenantLogic {
	return &UpdateTenantLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateTenantLogic) UpdateTenant(in *core.TenantInfo) (*core.BaseResp, error) {
	if in.Id == nil {
		return nil, errorx.NewInvalidArgumentError("tenant.idRequired")
	}

	// 验证租户是否存在
	exist, err := l.svcCtx.DB.Tenant.Query().Where(tenant.IDEQ(*in.Id)).Exist(hooks.NewSystemContext(l.ctx))
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	if !exist {
		return nil, errorx.NewInvalidArgumentError("tenant.notFound")
	}

	// 验证租户编码唯一性（排除当前租户）
	if in.Code != nil {
		checkCode, err := l.svcCtx.DB.Tenant.Query().
			Where(tenant.CodeEQ(*in.Code), tenant.IDNEQ(*in.Id)).
			Exist(hooks.NewSystemContext(l.ctx))
		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}

		if checkCode {
			return nil, errorx.NewInvalidArgumentError("tenant.codeExist")
		}
	}

	// 验证租户名称唯一性（排除当前租户）
	if in.Name != nil {
		checkName, err := l.svcCtx.DB.Tenant.Query().
			Where(tenant.NameEQ(*in.Name), tenant.IDNEQ(*in.Id)).
			Exist(hooks.NewSystemContext(l.ctx))
		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}

		if checkName {
			return nil, errorx.NewInvalidArgumentError("tenant.nameExist")
		}
	}

	// 构建更新请求
	builder := l.svcCtx.DB.Tenant.UpdateOneID(*in.Id)

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
	}

	if in.ExpiredAt != nil {
		builder.SetExpiredAt(time.Unix(*in.ExpiredAt, 0))
	}

	if in.Config != nil {
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(*in.Config), &config); err != nil {
			return nil, errorx.NewInvalidArgumentError("tenant.invalidConfig")
		}
		builder.SetConfig(config)
	}

	if in.CreatedBy != nil {
		builder.SetCreatedBy(*in.CreatedBy)
	}

	// 使用系统上下文更新租户
	_, err = builder.Save(hooks.NewSystemContext(l.ctx))
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	logx.Infow("Tenant updated successfully",
		logx.Field("tenant_id", *in.Id))

	return &core.BaseResp{
		Msg: i18n.UpdateSuccess,
	}, nil
}
