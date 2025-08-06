package tenant

import (
	"context"

	"github.com/coder-lulu/newbee-common/orm/ent/hooks"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type InitTenantDataLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewInitTenantDataLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitTenantDataLogic {
	return &InitTenantDataLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// InitTenantData 初始化租户数据（创建默认角色、部门等）
func (l *InitTenantDataLogic) InitTenantData(in *core.IDReq) (*core.BaseResp, error) {
	// 使用系统上下文来绕过租户检查
	systemCtx := hooks.NewSystemContext(l.ctx)

	tenantID := in.Id

	// 为新租户创建context
	tenantCtx := hooks.SetTenantIDToContext(systemCtx, tenantID)

	// 创建默认角色
	err := l.createDefaultRole(tenantCtx, tenantID)
	if err != nil {
		logx.Errorw("Failed to create default role for tenant",
			logx.Field("tenantId", tenantID),
			logx.Field("error", err))
		return nil, err
	}

	// 创建默认部门
	err = l.createDefaultDepartment(tenantCtx, tenantID)
	if err != nil {
		logx.Errorw("Failed to create default department for tenant",
			logx.Field("tenantId", tenantID),
			logx.Field("error", err))
		return nil, err
	}

	// 创建默认职位
	err = l.createDefaultPosition(tenantCtx, tenantID)
	if err != nil {
		logx.Errorw("Failed to create default position for tenant",
			logx.Field("tenantId", tenantID),
			logx.Field("error", err))
		return nil, err
	}

	logx.Infow("Tenant data initialized successfully",
		logx.Field("tenantId", tenantID))

	return &core.BaseResp{Msg: "Tenant data initialized successfully"}, nil
}

func (l *InitTenantDataLogic) createDefaultRole(ctx context.Context, tenantID uint64) error {
	_, err := l.svcCtx.DB.Role.Create().
		SetName("管理员").
		SetCode("admin").
		SetDefaultRouter("dashboard").
		SetRemark("租户管理员角色").
		SetSort(1).
		SetDataScope(1). // 全部数据权限
		SetTenantID(tenantID).
		Save(ctx)

	return err
}

func (l *InitTenantDataLogic) createDefaultDepartment(ctx context.Context, tenantID uint64) error {
	_, err := l.svcCtx.DB.Department.Create().
		SetName("总部").
		SetAncestors("0").
		SetLeader("admin").
		SetPhone("").
		SetEmail("").
		SetRemark("默认部门").
		SetSort(1).
		SetTenantID(tenantID).
		Save(ctx)

	return err
}

func (l *InitTenantDataLogic) createDefaultPosition(ctx context.Context, tenantID uint64) error {
	_, err := l.svcCtx.DB.Position.Create().
		SetName("管理员").
		SetCode("admin").
		SetRemark("默认管理员职位").
		SetSort(1).
		SetTenantID(tenantID).
		Save(ctx)

	return err
}
