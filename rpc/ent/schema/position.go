package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/coder-lulu/newbee-common/v2/orm/ent/mixins"
	commonMixins "github.com/coder-lulu/newbee-common/v2/orm/ent/mixins"
)

type Position struct {
	ent.Schema
}

func (Position) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("Position Name | 职位名称"),
		field.String("code").
			Comment("The code of position | 职位编码"),
		field.String("remark").Optional().
			Comment("Remark | 备注"),
		field.Uint64("dept_id").Default(1).Optional().
			Comment("deptId | 所属部门ID"),
	}
}

func (Position) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.StatusMixin{},
		mixins.SortMixin{},
		commonMixins.TenantMixin{},
	}
}

func (Position) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("users", User.Type).Ref("positions"),
		edge.From("departments", Department.Type).Ref("posts").Unique().Field("dept_id"),
	}
}

func (Position) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code", "dept_id", "tenant_id").Unique(),
	}
}

func (Position) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("Position Table | 职位信息表"),
		entsql.Annotation{Table: "sys_positions"},
	}
}
