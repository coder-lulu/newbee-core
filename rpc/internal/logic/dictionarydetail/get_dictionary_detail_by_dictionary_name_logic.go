package dictionarydetail

import (
	"context"
	"github.com/coder-lulu/newbee-common/enum/common"

	"github.com/coder-lulu/newbee-common/utils/pointy"

	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/ent/dictionary"
	"github.com/coder-lulu/newbee-core/rpc/ent/dictionarydetail"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDictionaryDetailByDictionaryNameLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDictionaryDetailByDictionaryNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDictionaryDetailByDictionaryNameLogic {
	return &GetDictionaryDetailByDictionaryNameLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetDictionaryDetailByDictionaryNameLogic) GetDictionaryDetailByDictionaryName(in *core.BaseMsg) (*core.DictionaryDetailListResp, error) {
	dictionaryData, err := l.svcCtx.DB.Dictionary.Query().Where(dictionary.NameEQ(in.Msg)).First(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	result, err := l.svcCtx.DB.DictionaryDetail.Query().Where(dictionarydetail.DictionaryID(dictionaryData.ID), dictionarydetail.StatusEQ(common.StatusNormal)).
		Page(l.ctx, 1, 10000, func(pager *ent.DictionaryDetailPager) {
			pager.Order = ent.Asc(dictionarydetail.FieldSort)
		})
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	resp := &core.DictionaryDetailListResp{}
	resp.Total = result.PageDetails.Total

	for _, v := range result.List {
		resp.Data = append(resp.Data, &core.DictionaryDetailInfo{
			Id:           &v.ID,
			CreatedAt:    pointy.GetPointer(v.CreatedAt.UnixMilli()),
			UpdatedAt:    pointy.GetPointer(v.UpdatedAt.UnixMilli()),
			Status:       pointy.GetPointer(uint32(v.Status)),
			Title:        &v.Title,
			Value:        &v.Value,
			DictionaryId: &v.DictionaryID,
			Sort:         &v.Sort,
			IsDefault:    &v.IsDefault,
			CssClass:     &v.CSSClass,
			ListClass:    &v.ListClass,
		})
	}

	return resp, nil
}
