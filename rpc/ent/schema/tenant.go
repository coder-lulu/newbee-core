package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/coder-lulu/newbee-common/orm/ent/mixins"
)

type Tenant struct {
	ent.Schema
}

func (Tenant) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Comment("Tenant name | 租户名称"),
		field.String("code").
			Unique().
			NotEmpty().
			Comment("Tenant code for identification | 租户标识码"),
		field.String("description").
			Optional().
			Comment("Tenant description | 租户描述"),
		field.Time("expired_at").
			Optional().
			Comment("Tenant expiration time | 租户过期时间"),
		field.JSON("config", map[string]interface{}{}).
			Optional().
			Comment("Tenant configuration | 租户配置"),
		field.Uint64("created_by").
			Optional().
			Comment("Creator user ID | 创建者用户ID"),
	}
}

func (Tenant) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.StatusMixin{},
		// 注意：租户表本身不需要TenantMixin，因为它是租户管理的根实体
	}
}

func (Tenant) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").Unique(),
		index.Fields("status"),
		index.Fields("expired_at"),
	}
}

func (Tenant) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("Tenant Table | 租户信息表"),
		entsql.Annotation{Table: "sys_tenants"},
	}
}
