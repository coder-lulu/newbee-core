package dictionarydetail

import (
	"context"

	"github.com/coder-lulu/newbee-common/v2/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/ent/dictionarydetail"
	"github.com/coder-lulu/newbee-core/rpc/ent/predicate"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"
)

type GetDictionaryDetailListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDictionaryDetailListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDictionaryDetailListLogic {
	return &GetDictionaryDetailListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetDictionaryDetailListLogic) GetDictionaryDetailList(in *core.DictionaryDetailListReq) (*core.DictionaryDetailListResp, error) {
	var predicates []predicate.DictionaryDetail
	if in.DictionaryId != nil {
		predicates = append(predicates, dictionarydetail.DictionaryIDEQ(*in.DictionaryId))
	}
	if in.Title != nil {
		predicates = append(predicates, dictionarydetail.TitleContains(*in.Title))
	}
	result, err := l.svcCtx.DB.DictionaryDetail.Query().Where(predicates...).Page(l.ctx, in.Page, in.PageSize, func(pager *ent.DictionaryDetailPager) {
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
