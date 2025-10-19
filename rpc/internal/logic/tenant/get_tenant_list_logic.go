package tenant

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/coder-lulu/newbee-core/rpc/ent/predicate"
	"github.com/coder-lulu/newbee-core/rpc/ent/tenant"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/contextx"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/coder-lulu/newbee-common/v2/i18n"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/hooks"
	"github.com/coder-lulu/newbee-common/v2/utils/pointy"

	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenantListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTenantListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantListLogic {
	return &GetTenantListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetTenantListLogic) GetTenantList(in *core.TenantListReq) (*core.TenantListResp, error) {
	// 权限检查：只有管理员角色才允许获取租户列表
	// 使用统一的 context 提取工具
	authInfo := contextx.ExtractAllAuthInfo(l.ctx)
	roleCodes := authInfo["role_codes"]

	logx.Infow("权限检查",
		logx.Field("tenant_id", authInfo["tenant_id"]),
		logx.Field("user_id", authInfo["user_id"]),
		logx.Field("dept_id", authInfo["dept_id"]),
		logx.Field("role_codes", roleCodes))

	if roleCodes == "" {
		return nil, errorx.NewCodeInvalidArgumentError(i18n.PermissionDeny)
	}

	// 检查是否包含管理员角色（superadmin 或 admin）
	roleList := strings.Split(roleCodes, ",")
	hasAdminRole := false
	for _, role := range roleList {
		roleCode := strings.TrimSpace(role)
		if roleCode == "superadmin" {
			hasAdminRole = true
			break
		}
	}

	if !hasAdminRole {
		return nil, errorx.NewCodeInvalidArgumentError(i18n.PermissionDeny)
	}

	var predicates []predicate.Tenant

	// 构建查询条件
	if in.Name != nil {
		predicates = append(predicates, tenant.NameContains(*in.Name))
	}

	if in.Code != nil {
		predicates = append(predicates, tenant.CodeContains(*in.Code))
	}

	if in.Status != nil {
		predicates = append(predicates, tenant.StatusEQ(uint8(*in.Status)))
	}

	if in.CreatedBy != nil {
		predicates = append(predicates, tenant.CreatedByEQ(*in.CreatedBy))
	}

	// 使用系统上下文查询所有租户（不受租户隔离限制）
	result, err := l.svcCtx.DB.Tenant.Query().
		Where(predicates...).
		Order(tenant.ByCreatedAt()).
		Offset(int((in.Page - 1) * in.PageSize)).
		Limit(int(in.PageSize)).
		All(hooks.NewSystemContext(l.ctx))

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 获取总数
	total, err := l.svcCtx.DB.Tenant.Query().
		Where(predicates...).
		Count(hooks.NewSystemContext(l.ctx))

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 转换结果
	resp := &core.TenantListResp{
		Total: uint64(total),
		Data:  make([]*core.TenantInfo, 0, len(result)),
	}

	for _, tenant := range result {
		var configStr string
		if tenant.Config != nil {
			configBytes, _ := json.Marshal(tenant.Config)
			configStr = string(configBytes)
		}

		tenantInfo := &core.TenantInfo{
			Id:          pointy.GetPointer(tenant.ID),
			CreatedAt:   pointy.GetPointer(tenant.CreatedAt.Unix()),
			UpdatedAt:   pointy.GetPointer(tenant.UpdatedAt.Unix()),
			Status:      pointy.GetPointer(uint32(tenant.Status)),
			Name:        pointy.GetPointer(tenant.Name),
			Code:        pointy.GetPointer(tenant.Code),
			Description: pointy.GetPointer(tenant.Description),
			ExpiredAt:   pointy.GetPointer(tenant.ExpiredAt.Unix()),
			Config:      pointy.GetPointer(configStr),
			CreatedBy:   pointy.GetPointer(tenant.CreatedBy),
		}

		resp.Data = append(resp.Data, tenantInfo)
	}

	return resp, nil
}
