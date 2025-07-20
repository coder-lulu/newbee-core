package schema

import (
	"context"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/suyuan32/simple-admin-common/orm/ent/entctx/datapermctx"
	"github.com/suyuan32/simple-admin-common/orm/ent/entctx/deptctx"
	"github.com/suyuan32/simple-admin-common/orm/ent/entctx/userctx"
	"github.com/suyuan32/simple-admin-common/orm/ent/entenum"
	"github.com/suyuan32/simple-admin-common/orm/ent/mixins"
	"github.com/zeromicro/go-zero/core/errorx"

	"github.com/coder-lulu/newbee-core/rpc/ent/intercept"
	mixins2 "github.com/coder-lulu/newbee-core/rpc/ent/schema/mixins"
	"github.com/coder-lulu/newbee-core/rpc/ent/user"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").Unique().
			Comment("User's login name | 登录名"),
		field.String("password").
			Comment("Password | 密码"),
		field.String("nickname").Unique().
			Comment("Nickname | 昵称"),
		field.String("description").Optional().
			Comment("The description of user | 用户的描述信息"),
		field.String("home_path").Default("/dashboard").
			Comment("The home page that the user enters after logging in | 用户登陆后进入的首页"),
		field.String("mobile").Optional().
			Comment("Mobile number | 手机号"),
		field.String("email").Optional().
			Comment("Email | 邮箱号"),
		field.String("avatar").
			SchemaType(map[string]string{dialect.MySQL: "varchar(512)"}).
			Optional().
			Comment("Avatar | 头像路径"),
		field.Uint64("department_id").Optional().Default(1).
			Comment("Department ID | 部门ID"),
	}
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.UUIDMixin{},
		mixins.StatusMixin{},
		mixins2.SoftDeleteMixin{},
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("departments", Department.Type).Unique().Field("department_id"),
		edge.To("positions", Position.Type),
		edge.To("roles", Role.Type),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("username", "email").
			Unique(),
	}
}

func (User) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		intercept.Func(func(ctx context.Context, q intercept.Query) error {
			dataScope, err := datapermctx.GetScopeFromCtx(ctx)
			if err != nil {
				return err
			}

			switch dataScope {
			case entenum.DataPermAll:
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
		}),
	}
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("User Table | 用户信息表"),
		entsql.Annotation{Table: "sys_users"},
	}
}
