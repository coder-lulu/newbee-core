package tenant

import (
	"context"
	"encoding/json"

	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/ent/tenant"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/coder-lulu/newbee-common/orm/ent/hooks"
	"github.com/coder-lulu/newbee-common/utils/pointy"

	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenantByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTenantByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantByIdLogic {
	return &GetTenantByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetTenantByIdLogic) GetTenantById(in *core.IDReq) (*core.TenantInfo, error) {
	// 权限检查：只有角色为superadmin的用户才允许查看
	// roleCodes, ok := l.ctx.Value(keys.RoleCodesKey).(string)
	// if !ok || roleCodes == "" {
	// 	return nil, errorx.NewCodeInvalidArgumentError(i18n.PermissionDeny)
	// }

	// 检查是否包含superadmin角色
	// roleList := strings.Split(roleCodes, ",")
	// hasSuperAdminRole := false
	// for _, role := range roleList {
	// 	if strings.TrimSpace(role) == "superadmin" {
	// 		hasSuperAdminRole = true
	// 		break
	// 	}
	// }

	// if !hasSuperAdminRole {
	// 	return nil, errorx.NewCodeInvalidArgumentError(i18n.PermissionDeny)
	// }

	// 使用系统上下文查询租户（不受租户隔离限制）
	result, err := l.svcCtx.DB.Tenant.Query().
		Where(tenant.IDEQ(in.Id)).
		Only(hooks.NewSystemContext(l.ctx))

	if err != nil {
		switch {
		case ent.IsNotFound(err):
			return nil, errorx.NewInvalidArgumentError("tenant.notFound")
		default:
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}
	}

	var configStr string
	if result.Config != nil {
		configBytes, _ := json.Marshal(result.Config)
		configStr = string(configBytes)
	}

	return &core.TenantInfo{
		Id:          pointy.GetPointer(result.ID),
		CreatedAt:   pointy.GetPointer(result.CreatedAt.Unix()),
		UpdatedAt:   pointy.GetPointer(result.UpdatedAt.Unix()),
		Status:      pointy.GetPointer(uint32(result.Status)),
		Name:        pointy.GetPointer(result.Name),
		Code:        pointy.GetPointer(result.Code),
		Description: pointy.GetPointer(result.Description),
		ExpiredAt:   pointy.GetPointer(result.ExpiredAt.Unix()),
		Config:      pointy.GetPointer(configStr),
		CreatedBy:   pointy.GetPointer(result.CreatedBy),
	}, nil
}
