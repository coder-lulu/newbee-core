package auditlog

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/coder-lulu/newbee-common/v2/middleware/keys"
	"github.com/coder-lulu/newbee-common/v2/utils/pointy"
	"github.com/coder-lulu/newbee-common/v2/utils/uuidx"
	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/ent/auditlog"
	"github.com/coder-lulu/newbee-core/rpc/ent/predicate"
	"github.com/coder-lulu/newbee-core/rpc/ent/user"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
)

type GetAuditLogListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAuditLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAuditLogListLogic {
	return &GetAuditLogListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetAuditLogListLogic) GetAuditLogList(in *core.AuditLogListReq) (*core.AuditLogListResp, error) {
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

	var predicates []predicate.AuditLog
	predicates = append(predicates, auditlog.TenantIDEQ(tenantIDStr))

	// 基本查询条件
	if in.UserId != nil {
		predicates = append(predicates, auditlog.UserIDContains(*in.UserId))
	}
	if in.UserName != nil {
		predicates = append(predicates, auditlog.UserNameContains(*in.UserName))
	}
	if in.OperationType != nil {
		predicates = append(predicates, auditlog.OperationTypeEQ(auditlog.OperationType(*in.OperationType)))
	}
	if in.ResourceType != nil {
		predicates = append(predicates, auditlog.ResourceTypeContains(*in.ResourceType))
	}
	if in.ResourceId != nil {
		predicates = append(predicates, auditlog.ResourceIDContains(*in.ResourceId))
	}
	if in.RequestMethod != nil {
		predicates = append(predicates, auditlog.RequestMethodEQ(*in.RequestMethod))
	}
	if in.RequestPath != nil {
		predicates = append(predicates, auditlog.RequestPathContains(*in.RequestPath))
	}
	if in.IpAddress != nil {
		predicates = append(predicates, auditlog.IPAddressEQ(*in.IpAddress))
	}
	if in.ResponseStatus != nil {
		predicates = append(predicates, auditlog.ResponseStatusEQ(int(*in.ResponseStatus)))
	}

	// 时间范围查询
	if in.StartTime != nil {
		predicates = append(predicates, auditlog.CreatedAtGTE(time.UnixMilli(*in.StartTime)))
	}
	if in.EndTime != nil {
		predicates = append(predicates, auditlog.CreatedAtLTE(time.UnixMilli(*in.EndTime)))
	}

	// 耗时范围查询
	if in.MinDuration != nil {
		predicates = append(predicates, auditlog.DurationMsGTE(*in.MinDuration))
	}
	if in.MaxDuration != nil {
		predicates = append(predicates, auditlog.DurationMsLTE(*in.MaxDuration))
	}

	result, err := l.svcCtx.DB.AuditLog.Query().Where(predicates...).Order(ent.Desc(auditlog.FieldCreatedAt)).Page(l.ctx, in.Page, in.PageSize)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	resp := &core.AuditLogListResp{}
	resp.Total = result.PageDetails.Total

	// 收集所有唯一的用户ID进行批量查询
	userIdSet := make(map[string]bool)
	for _, v := range result.List {
		if v.UserID != "" {
			userIdSet[v.UserID] = true
		}
	}

	// 批量查询用户信息（限制在当前租户）
	userMap := make(map[string]string)
	if len(userIdSet) > 0 {
		userIds := make([]string, 0, len(userIdSet))
		for userId := range userIdSet {
			userIds = append(userIds, userId)
		}

		users, err := l.svcCtx.DB.User.Query().
			Where(func() predicate.User {
				// 构建UUID查询条件
				var uuidPredicates []predicate.User
				for _, userId := range userIds {
					if uuid := uuidx.ParseUUIDString(userId); uuid.String() != "00000000-0000-0000-0000-000000000000" {
						uuidPredicates = append(uuidPredicates, user.IDEQ(uuid))
					}
				}
				if len(uuidPredicates) == 0 {
					return user.ID(uuidx.ParseUUIDString("00000000-0000-0000-0000-000000000000"))
				}
				return user.And(user.Or(uuidPredicates...), user.TenantIDEQ(tenantID))
			}()).
			All(l.ctx)

		if err != nil {
			l.Logger.Errorw("Failed to batch query users for audit log display",
				logx.Field("error", err),
				logx.Field("userCount", len(userIds)))
		} else {
			// 构建用户ID到用户名的映射
			for _, u := range users {
				username := u.Username
				if username == "" && u.Nickname != "" {
					username = u.Nickname
				}
				if username == "" && u.Email != "" {
					username = u.Email
				}
				if username == "" {
					username = "[用户名为空]"
				}
				userMap[u.ID.String()] = username
			}

			l.Logger.Infow("Successfully batched user lookup for audit display",
				logx.Field("queriedUsers", len(userIds)),
				logx.Field("foundUsers", len(users)),
				logx.Field("userMapSize", len(userMap)))
		}
	}

	for _, v := range result.List {
		// 查找对应的用户名
		userName := v.UserName // 使用数据库存储的用户名（通常为空）
		if v.UserID != "" {
			if foundUserName, exists := userMap[v.UserID]; exists {
				userName = foundUserName
			} else if v.UserID != "" {
				userName = "[用户不存在]" // 数据库中找不到对应用户
			}
		}

		resp.Data = append(resp.Data, &core.AuditLogInfo{
			Id:             pointy.GetPointer(v.ID.String()),
			CreatedAt:      pointy.GetPointer(v.CreatedAt.UnixMilli()),
			UpdatedAt:      pointy.GetPointer(v.UpdatedAt.UnixMilli()),
			Status:         pointy.GetPointer(uint32(v.Status)),
			TenantId:       &v.TenantID,
			UserId:         &v.UserID,
			UserName:       &userName, // 使用批量查询得到的用户名
			OperationType:  pointy.GetPointer(string(v.OperationType)),
			ResourceType:   &v.ResourceType,
			ResourceId:     &v.ResourceID,
			RequestMethod:  &v.RequestMethod,
			RequestPath:    &v.RequestPath,
			RequestData:    &v.RequestData,
			ResponseStatus: pointy.GetPointer(int64(v.ResponseStatus)),
			ResponseData:   &v.ResponseData,
			IpAddress:      &v.IPAddress,
			UserAgent:      &v.UserAgent,
			DurationMs:     &v.DurationMs,
			ErrorMessage:   &v.ErrorMessage,
			Metadata: func() *string {
				if v.Metadata != nil {
					if jsonData, err := json.Marshal(v.Metadata); err == nil {
						return pointy.GetPointer(string(jsonData))
					}
				}
				return pointy.GetPointer("")
			}(),
		})
	}

	return resp, nil
}
