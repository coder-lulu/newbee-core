package dictionarydetail

import (
	"context"
	"github.com/coder-lulu/newbee-common/utils/pointy"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/coder-lulu/newbee-common/i18n"
)

type CreateDictionaryDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateDictionaryDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateDictionaryDetailLogic {
	return &CreateDictionaryDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateDictionaryDetailLogic) CreateDictionaryDetail(in *core.DictionaryDetailInfo) (*core.BaseIDResp, error) {
	result, err := l.svcCtx.DB.DictionaryDetail.Create().
		SetNotNilStatus(pointy.GetStatusPointer(in.Status)).
		SetNotNilTitle(in.Title).
		SetNotNilValue(in.Value).
		SetNotNilSort(in.Sort).
		SetNotNilIsDefault(in.IsDefault).
		SetNotNilCSSClass(in.CssClass).
		SetNotNilListClass(in.ListClass).
		SetNotNilDictionaryID(in.DictionaryId).
		Save(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &core.BaseIDResp{Id: result.ID, Msg: i18n.CreateSuccess}, nil
}
