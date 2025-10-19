package dictionarydetail

import (
	"context"
	"fmt"

	"github.com/coder-lulu/newbee-common/v2/i18n"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateDictionaryDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateDictionaryDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDictionaryDetailLogic {
	return &UpdateDictionaryDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateDictionaryDetailLogic) UpdateDictionaryDetail(in *core.DictionaryDetailInfo) (*core.BaseResp, error) {
	fmt.Println("init here")
	err := l.svcCtx.DB.DictionaryDetail.UpdateOneID(*in.Id).
		//SetNillableStatus(pointy.GetStatusPointer(in.Status)).
		SetNotNilTitle(in.Title).
		SetNotNilValue(in.Value).
		SetNotNilSort(in.Sort).
		SetNotNilCSSClass(in.CssClass).
		SetNotNilListClass(in.ListClass).
		SetNotNilDictionaryID(in.DictionaryId).
		Exec(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &core.BaseResp{Msg: i18n.UpdateSuccess}, nil
}
