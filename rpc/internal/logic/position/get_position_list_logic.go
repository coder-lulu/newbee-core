package position

import (
	"context"
	"entgo.io/ent/dialect/sql"
	dept "github.com/coder-lulu/newbee-core/rpc/ent/department"

	"github.com/coder-lulu/newbee-common/v2/utils/pointy"

	"github.com/coder-lulu/newbee-core/rpc/ent/position"
	"github.com/coder-lulu/newbee-core/rpc/ent/predicate"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPositionListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPositionListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPositionListLogic {
	return &GetPositionListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetPositionListLogic) GetPositionList(in *core.PositionListReq) (*core.PositionListResp, error) {
	var predicates []predicate.Position
	if in.Name != nil {
		predicates = append(predicates, position.NameContains(*in.Name))
	}
	if in.Code != nil {
		predicates = append(predicates, position.CodeContains(*in.Code))
	}
	if in.Remark != nil {
		predicates = append(predicates, position.RemarkContains(*in.Remark))
	}
	if in.BelongDeptId != nil {
		var lists []uint64
		queue := []uint64{*in.BelongDeptId} // 使用队列代替递归

		for len(queue) > 0 {
			currentDeptId := queue[0]
			queue = queue[1:]

			// 添加当前部门 ID
			lists = append(lists, currentDeptId)

			// 查询当前部门的子部门
			result, err := l.svcCtx.DB.Department.Query().
				Where(dept.ParentID(currentDeptId)).
				All(l.ctx)
			if err != nil {
				return nil, err
			}

			// 将子部门 ID 加入队列
			for _, v := range result {
				queue = append(queue, v.ID)
			}
		}
		predicates = append(predicates, position.DeptIDIn(lists...))

	}

	result, err := l.svcCtx.DB.Position.Query().Where(predicates...).Order(position.BySort(sql.OrderAsc())).Page(l.ctx, in.Page, in.PageSize)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	resp := &core.PositionListResp{}
	resp.Total = result.PageDetails.Total

	for _, v := range result.List {
		resp.Data = append(resp.Data, &core.PositionInfo{
			Id:        &v.ID,
			CreatedAt: pointy.GetPointer(v.CreatedAt.UnixMilli()),
			UpdatedAt: pointy.GetPointer(v.UpdatedAt.UnixMilli()),
			Status:    pointy.GetPointer(uint32(v.Status)),
			Sort:      &v.Sort,
			Name:      &v.Name,
			Code:      &v.Code,
			Remark:    &v.Remark,
			DeptId:    &v.DeptID,
		})
	}

	return resp, nil
}
