package token

import (
	"context"
	"strconv"

	"github.com/coder-lulu/newbee-common/v2/i18n"
	"github.com/coder-lulu/newbee-common/v2/middleware/keys"
	"github.com/coder-lulu/newbee-common/v2/utils/pointy"
	"github.com/coder-lulu/newbee-common/v2/utils/uuidx"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
)

type CreateTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTokenLogic {
	return &CreateTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateTokenLogic) CreateToken(in *core.TokenInfo) (*core.BaseUUIDResp, error) {
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

	if in.TenantId != nil {
		if *in.TenantId != tenantID {
			return nil, errorx.NewInvalidArgumentError("tenant.mismatch")
		}
	} else {
		tenantIDCopy := tenantID
		in.TenantId = &tenantIDCopy
	}

	tokenCreate := l.svcCtx.DB.Token.Create().
		SetNotNilStatus(pointy.GetStatusPointer(in.Status)).
		SetNotNilUUID(uuidx.ParseUUIDStringToPointer(in.Uuid)).
		SetNotNilToken(in.Token).
		SetNotNilSource(in.Source).
		SetNotNilUsername(in.Username).
		SetNotNilExpiredAt(pointy.GetTimeMilliPointer(in.ExpiredAt)).
		SetTenantID(tenantID)

	result, err := tokenCreate.Save(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &core.BaseUUIDResp{Id: result.ID.String(), Msg: i18n.CreateSuccess}, nil
}
