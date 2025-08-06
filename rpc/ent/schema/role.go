package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/coder-lulu/newbee-common/orm/ent/mixins"
)

type Role struct {
	ent.Schema
}

func (Role) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("Role name | 角色名"),
		field.String("code").
			Comment("Role code for permission control in front end | 角色码，用于前端权限控制"),
		field.String("default_router").Default("dashboard").
			Comment("Default menu : dashboard | 默认登录页面"),
		field.String("remark").Default("").
			Comment("Remark | 备注"),
		field.Uint32("sort").Default(0).
			Comment("Order number | 排序编号"),
		field.Uint8("data_scope").
			Comment("Data scope 1 - all data 2 - custom dept data 3 - own dept and sub dept data 4 - own dept data  5 - your own data | 数据权限范围 1 - 所有数据 2 - 自定义部门数据 3 - 您所在部门及下属部门数据 4 - 您所在部门数据 5 - 本人数据").
			Default(1),
		field.JSON("custom_dept_ids", []uint64{}).
			Optional().
			Comment("Custom department setting for data permission | 自定义部门数据权限"),
	}
}

func (Role) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.StatusMixin{},
		mixins.TenantMixin{},
	}
}

func (Role) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("menus", Menu.Type),
		edge.From("users", User.Type).Ref("roles"),
	}
}

func (Role) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").Unique(),
	}
}

func (Role) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("Role Table | 角色信息表"),
		entsql.Annotation{Table: "sys_roles"},
	}
}
