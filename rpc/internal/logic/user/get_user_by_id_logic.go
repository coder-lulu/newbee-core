package user

import (
	"context"
	"strconv"
	"time"

	"github.com/coder-lulu/newbee-common/middleware/keys"
	hookshelper "github.com/coder-lulu/newbee-common/orm/ent/hooks"
	"github.com/coder-lulu/newbee-common/utils/pointy"
	"github.com/coder-lulu/newbee-common/utils/uuidx"

	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/ent/user"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserByIdLogic {
	return &GetUserByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserByIdLogic) GetUserById(in *core.UUIDReq) (*core.UserInfo, error) {
	// 移除数据权限过滤调用，因为RPC服务层不需要数据权限控制
	// 数据权限应该在API层通过中间件控制

	l.Infof("RPC GetUserById start - userId: %s", in.Id)
	start := time.Now()

	cm := keys.NewContextManager()
	queryCtx := l.ctx
	activeTenant := cm.GetTenantID(l.ctx)
	originalTenant := cm.GetOriginalTenantID(l.ctx)
	if originalTenant != "" && originalTenant != "0" && originalTenant != activeTenant {
		if originalTenantID, err := strconv.ParseUint(originalTenant, 10, 64); err == nil {
			queryCtx = hookshelper.SetTenantIDToContext(l.ctx, originalTenantID)
			l.Infof("RPC GetUserById using original tenant context - userId: %s, originalTenantId: %s", in.Id, originalTenant)
		} else {
			// 解析失败时保持当前上下文，避免回退至系统上下文造成越权
			l.Errorf("RPC GetUserById rejected invalid original tenant id - value: %s, error: %v", originalTenant, err)
			queryCtx = l.ctx
		}
	}

	result, err := l.svcCtx.DB.User.Query().Where(user.IDEQ(uuidx.ParseUUIDString(in.Id))).WithRoles().WithDepartments().WithPositions().First(queryCtx)
	queryDuration := time.Since(start)
	l.Infof("RPC GetUserById database query completed - duration: %v", queryDuration)

	if err != nil {
		l.Errorf("RPC GetUserById database query error - duration: %v, error: %v", queryDuration, err)
		return nil, dberrorhandler.AuthUserEntError(l.Logger, err, in)
	}

	dataStart := time.Now()
	data := &core.UserInfo{
		Nickname:       &result.Nickname,
		Avatar:         &result.Avatar,
		RoleIds:        GetRoleIds(result.Edges.Roles),
		RoleNames:      GetRoleNames(result.Edges.Roles),
		RoleCodes:      GetRoleCodes(result.Edges.Roles),
		PositionIds:    GetPositionIds(result.Edges.Positions),
		Mobile:         &result.Mobile,
		Email:          &result.Email,
		Status:         pointy.GetPointer(uint32(result.Status)),
		Id:             pointy.GetPointer(result.ID.String()),
		Username:       &result.Username,
		HomePath:       &result.HomePath,
		Password:       &result.Password,
		Description:    &result.Description,
		DepartmentId:   &result.DepartmentID,
		DepartmentName: &result.Edges.Departments.Name,
		CreatedAt:      pointy.GetPointer(result.CreatedAt.UnixMilli()),
		UpdatedAt:      pointy.GetPointer(result.UpdatedAt.UnixMilli()),
	}
	dataDuration := time.Since(dataStart)
	totalDuration := time.Since(start)
	l.Infof("RPC GetUserById completed - query: %v, data mapping: %v, total: %v", queryDuration, dataDuration, totalDuration)

	return data, nil
}

func GetRoleIds(data []*ent.Role) []uint64 {
	var ids []uint64
	for _, v := range data {
		ids = append(ids, v.ID)
	}
	return ids
}

func GetRoleNames(data []*ent.Role) []string {
	var codes []string
	for _, v := range data {
		codes = append(codes, v.Name)
	}
	return codes
}

func GetRoleCodes(data []*ent.Role) []string {
	var codes []string
	for _, v := range data {
		codes = append(codes, v.Code)
	}
	return codes
}

func GetPositionIds(data []*ent.Position) []uint64 {
	var ids []uint64
	for _, v := range data {
		ids = append(ids, v.ID)
	}
	return ids
}

func IsSuperAdmin(data []*ent.Role) bool {
	for _, v := range data {
		if v.Code == "superadmin" {
			return true
		}
	}
	return false
}
