package hook

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/coder-lulu/newbee-common/orm/ent/entctx/datapermctx"
	"github.com/coder-lulu/newbee-common/orm/ent/entctx/deptctx"
	"github.com/coder-lulu/newbee-common/orm/ent/entctx/userctx"
	"github.com/coder-lulu/newbee-common/orm/ent/entenum"
	"github.com/coder-lulu/newbee-common/orm/ent/hooks"
	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/ent/intercept"
	"github.com/coder-lulu/newbee-core/rpc/ent/user"
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
)

// DataPermissionInterceptor returns an interceptor that applies data permission filtering for user queries.
func DataPermissionInterceptor() ent.Interceptor {
	return intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		// Skip data permission check for system context
		if hooks.IsSystemContext(ctx) {
			logx.Infow("SystemContext bypassing data permission interceptor",
				logx.Field("query_type", q.Type()),
				logx.Field("action", "bypass_data_permission"))
			return nil
		}

		// Only apply to User queries
		if q.Type() != user.Table {
			return nil
		}

		dataScope, err := datapermctx.GetScopeFromCtx(ctx)
		if err != nil {
			return err
		}

		switch dataScope {
		case entenum.DataPermAll:
			// No filtering needed
		case entenum.DataPermCustomDept:
			customDeptIds, err := datapermctx.GetCustomDeptFromCtx(ctx)
			if err != nil {
				return err
			}
			q.WhereP(func(selector *sql.Selector) {
				sql.FieldIn(user.FieldDepartmentID, customDeptIds...)(selector)
			})
		case entenum.DataPermOwnDeptAndSub:
			subDeptIds, err := datapermctx.GetSubDeptFromCtx(ctx)
			if err != nil {
				return err
			}

			q.WhereP(func(selector *sql.Selector) {
				sql.FieldIn(user.FieldDepartmentID, subDeptIds...)(selector)
			})
		case entenum.DataPermOwnDept:
			deptId, err := deptctx.GetDepartmentIDFromCtx(ctx)
			if err != nil {
				return err
			}

			q.WhereP(func(selector *sql.Selector) {
				selector.Where(sql.EQ(user.FieldDepartmentID, deptId))
			})
		case entenum.DataPermSelf:
			fieldName, err := datapermctx.GetFilterFieldFromCtx(ctx)
			if err != nil {
				return err
			}

			userId, err := userctx.GetUserIDFromCtx(ctx)
			if err != nil {
				return err
			}

			q.WhereP(func(selector *sql.Selector) {
				selector.Where(sql.EQ(fieldName, userId))
			})
		default:
			return errorx.NewInvalidArgumentError("data scope not supported")
		}
		return nil
	})
}
