package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/coder-lulu/newbee-common/orm/ent/mixins"
	commonMixins "github.com/coder-lulu/newbee-common/orm/ent/mixins"
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
		// 🔥 Phase 3.2: data_scope field removed - now managed via sys_casbin_rules (ptype='d')
		// Data permission scope is determined by querying casbin rules at runtime
		field.JSON("custom_dept_ids", []uint64{}).
			Optional().
			Comment("Custom department setting for data permission | 自定义部门数据权限"),
	}
}

func (Role) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.StatusMixin{},
		commonMixins.TenantMixin{},
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
		index.Fields("code", "tenant_id").Unique(),
	}
}

func (Role) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("Role Table | 角色信息表"),
		entsql.Annotation{Table: "sys_roles"},
	}
}
