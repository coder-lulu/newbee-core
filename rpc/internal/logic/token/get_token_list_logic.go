package token

import (
	"context"
	"strconv"

	"github.com/coder-lulu/newbee-common/middleware/keys"
	"github.com/coder-lulu/newbee-common/utils/pointy"
	"github.com/coder-lulu/newbee-common/utils/uuidx"
	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/ent/predicate"
	"github.com/coder-lulu/newbee-core/rpc/ent/token"
	"github.com/coder-lulu/newbee-core/rpc/ent/user"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
)

type GetTokenListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTokenListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTokenListLogic {
	return &GetTokenListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetTokenListLogic) GetTokenList(in *core.TokenListReq) (*core.TokenListResp, error) {
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

	var tokens *ent.TokenPageList
	if in.Username == nil && in.Uuid == nil && in.Nickname == nil && in.Email == nil {
		tokens, err = l.svcCtx.DB.Token.Query().
			Where(token.TenantIDEQ(tenantID)).
			Page(l.ctx, in.Page, in.PageSize)

		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}
	} else {
		var predicates []predicate.User

		if in.Uuid != nil {
			predicates = append(predicates, user.IDEQ(uuidx.ParseUUIDString(*in.Uuid)))
		}

		if in.Username != nil {
			predicates = append(predicates, user.Username(*in.Username))
		}

		if in.Email != nil {
			predicates = append(predicates, user.EmailEQ(*in.Email))
		}

		if in.Nickname != nil {
			predicates = append(predicates, user.NicknameEQ(*in.Nickname))
		}

		predicates = append(predicates, user.TenantIDEQ(tenantID))

		u, err := l.svcCtx.DB.User.Query().Where(predicates...).First(l.ctx)
		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}

		tokens, err = l.svcCtx.DB.Token.Query().
			Where(token.UUIDEQ(u.ID), token.TenantIDEQ(tenantID)).
			Page(l.ctx, in.Page, in.PageSize)

		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}
	}

	resp := &core.TokenListResp{}
	resp.Total = tokens.PageDetails.Total

	for _, v := range tokens.List {
		resp.Data = append(resp.Data, &core.TokenInfo{
			Id:        pointy.GetPointer(v.ID.String()),
			Uuid:      pointy.GetPointer(v.UUID.String()),
			Token:     &v.Token,
			Status:    pointy.GetPointer(uint32(v.Status)),
			Source:    &v.Source,
			Username:  &v.Username,
			ExpiredAt: pointy.GetPointer(v.ExpiredAt.UnixMilli()),
			CreatedAt: pointy.GetPointer(v.CreatedAt.UnixMilli()),
		})
	}

	return resp, nil
}
