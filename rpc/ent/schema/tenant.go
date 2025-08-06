package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/coder-lulu/newbee-common/orm/ent/mixins"
	mixins2 "github.com/coder-lulu/newbee-core/rpc/ent/schema/mixins"
	"github.com/coder-lulu/newbee-core/rpc/ent/schema/types"
)

type Tenant struct {
	ent.Schema
}

func (Tenant) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("Tenant name | 租户名称"),
		field.String("code").
			Unique().
			Comment("Tenant code | 租户代码"),
		field.String("domain").
			Optional().
			Comment("Tenant domain | 租户域名"),
		field.String("contact_person").
			Optional().
			Comment("Contact person | 联系人"),
		field.String("contact_phone").
			Optional().
			Comment("Contact phone | 联系电话"),
		field.String("contact_email").
			Optional().
			Comment("Contact email | 联系邮箱"),
		field.Text("description").
			Optional().
			Comment("Tenant description | 租户描述"),
		field.Time("expire_time").
			Optional().
			Comment("Expire time | 过期时间"),
		field.Uint32("max_users").
			Default(100).
			Comment("Maximum users | 最大用户数"),
		field.JSON("settings", []types.TenantSettings{}).
			Optional().
			Comment("Tenant settings | 租户配置"),
	}
}



func (Tenant) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.StatusMixin{},
		mixins2.SoftDeleteMixin{},
	}
}

func (Tenant) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").Unique(),
		index.Fields("domain").Unique(),
		index.Fields("status"),
	}
}

func (Tenant) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("Tenant Table | 租户表"),
		entsql.Annotation{Table: "sys_tenants"},
	}
}
