package position

import (
	"context"

	"github.com/coder-lulu/newbee-common/utils/pointy"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/coder-lulu/newbee-common/i18n"
)

type CreatePositionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreatePositionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePositionLogic {
	return &CreatePositionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreatePositionLogic) CreatePosition(in *core.PositionInfo) (*core.BaseIDResp, error) {
	result, err := l.svcCtx.DB.Position.Create().
		SetNotNilStatus(pointy.GetStatusPointer(in.Status)).
		SetNotNilSort(in.Sort).
		SetNotNilName(in.Name).
		SetNotNilCode(in.Code).
		SetNotNilRemark(in.Remark).
		SetNotNilDeptID(in.DeptId).
		Save(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &core.BaseIDResp{Id: result.ID, Msg: i18n.CreateSuccess}, nil
}
