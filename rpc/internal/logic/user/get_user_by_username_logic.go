package user

import (
	"context"
	"strconv"

	"github.com/coder-lulu/newbee-common/v2/middleware/keys"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/entctx/datapermctx"
	"github.com/coder-lulu/newbee-common/v2/utils/pointy"
	"github.com/coder-lulu/newbee-core/rpc/ent/user"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
)

type GetUserByUsernameLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserByUsernameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserByUsernameLogic {
	return &GetUserByUsernameLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserByUsernameLogic) GetUserByUsername(in *core.UsernameReq) (*core.UserInfo, error) {
	tenantIDStr, ok := l.ctx.Value(keys.TenantIDKey).(string)
	if !ok || tenantIDStr == "" {
		if md, mdOK := metadata.FromIncomingContext(l.ctx); mdOK {
			if vals := md.Get(keys.TenantIDKey.String()); len(vals) > 0 {
				tenantIDStr = vals[0]
			}
		}
	}

	if tenantIDStr == "" {
		return nil, errorx.NewInvalidArgumentError("tenant.missingContext")
	}

	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 64)
	if err != nil {
		return nil, errorx.NewInvalidArgumentError("tenant.invalidContext")
	}

	queryCtx := datapermctx.WithFilterFieldContext(l.ctx, user.FieldID)

	result, err := l.svcCtx.DB.User.Query().
		Where(user.UsernameEQ(in.Username), user.TenantIDEQ(tenantID)).
		WithRoles().
		First(queryCtx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &core.UserInfo{
		Nickname:     &result.Nickname,
		Avatar:       &result.Avatar,
		Password:     &result.Password,
		RoleIds:      GetRoleIds(result.Edges.Roles),
		RoleCodes:    GetRoleCodes(result.Edges.Roles),
		Mobile:       &result.Mobile,
		Email:        &result.Email,
		Status:       pointy.GetPointer(uint32(result.Status)),
		Id:           pointy.GetPointer(result.ID.String()),
		Username:     &result.Username,
		HomePath:     &result.HomePath,
		Description:  &result.Description,
		DepartmentId: &result.DepartmentID,
		TenantId:     &result.TenantID,
		CreatedAt:    pointy.GetPointer(result.CreatedAt.UnixMilli()),
		UpdatedAt:    pointy.GetPointer(result.UpdatedAt.UnixMilli()),
	}, nil
}
