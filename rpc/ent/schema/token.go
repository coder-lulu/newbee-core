package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gofrs/uuid/v5"

	"github.com/coder-lulu/newbee-common/v2/orm/ent/mixins"
	commonMixins "github.com/coder-lulu/newbee-common/v2/orm/ent/mixins"
)

type Token struct {
	ent.Schema
}

func (Token) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("uuid", uuid.UUID{}).
			Comment(" User's UUID | 用户的UUID"),
		field.String("username").
			Comment("Username | 用户名").
			Default("unknown"),
		field.String("token").
			Comment("Token string | Token 字符串").
			SchemaType(map[string]string{
				"mysql": "varchar(1000)",
			}),
		field.String("source").
			Comment("Log in source such as GitHub | Token 来源 （本地为core, 第三方如github等）"),
		field.Time("expired_at").
			Comment(" Expire time | 过期时间"),
		field.Uint64("department_id").Optional().Default(0).
			Comment("Department ID when token was issued | Token签发时的部门ID"),
	}
}

func (Token) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.UUIDMixin{},
		mixins.StatusMixin{},
		commonMixins.TenantMixin{},
	}
}

func (Token) Edges() []ent.Edge {
	return nil
}

func (Token) Indexes() []ent.Index {
	return []ent.Index{
		// UUID查询索引
		index.Fields("uuid", "tenant_id"),
		// 用户Token查询索引（带状态）
		index.Fields("uuid", "tenant_id", "status"),
		// 过期时间索引（用于清理过期Token）
		index.Fields("expired_at", "tenant_id"),
		// 部门级数据权限查询索引
		index.Fields("department_id", "tenant_id", "status"),
		// 注意：Token字段(varchar(1000))太长，无法创建完整唯一索引
		// Token的唯一性由JWT生成算法保证（包含timestamp+UUID+signature）
		// 查询时通过完整匹配，性能可接受
	}
}

func (Token) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("Token Log Table | 令牌信息表"),
		entsql.Annotation{Table: "sys_tokens"},
	}
}
