package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	uuid "github.com/gofrs/uuid/v5"

	"github.com/coder-lulu/newbee-common/orm/ent/mixins"
)

type OauthAccount struct {
	ent.Schema
}

func (OauthAccount) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}).
			Comment("Associated user ID | 关联的用户ID"),
		field.Uint64("provider_id").
			Comment("OAuth provider ID | OAuth提供商ID"),
		field.String("provider_type").MaxLen(20).
			Comment("Provider type (wechat, qq, github, google, facebook) | 提供商类型"),
		field.String("provider_user_id").MaxLen(100).
			Comment("User ID from OAuth provider | 第三方平台的用户ID"),
		field.String("provider_username").MaxLen(100).Optional().
			Comment("Username from OAuth provider | 第三方平台的用户名"),
		field.String("provider_nickname").MaxLen(100).Optional().
			Comment("Nickname from OAuth provider | 第三方平台的昵称"),
		field.String("provider_email").MaxLen(255).Optional().
			Comment("Email from OAuth provider | 第三方平台的邮箱"),
		field.String("provider_avatar").MaxLen(500).Optional().
			Comment("Avatar URL from OAuth provider | 第三方平台的头像URL"),
		field.String("access_token").MaxLen(2000).
			Comment("Access token (encrypted) | 访问令牌（加密存储）"),
		field.String("refresh_token").MaxLen(2000).Optional().
			Comment("Refresh token (encrypted) | 刷新令牌（加密存储）"),
		field.Time("token_expires_at").Optional().
			Comment("Token expiration time | 令牌过期时间"),
		field.JSON("extra_data", map[string]interface{}{}).Optional().
			Comment("Extra data from provider | 第三方平台的额外数据"),
		field.Time("last_login_at").Optional().
			Comment("Last login time | 最后登录时间"),
		field.String("last_login_ip").MaxLen(45).Optional().
			Comment("Last login IP address | 最后登录IP地址"),
		field.Uint32("login_count").Default(0).
			Comment("Login count | 登录次数"),
		field.Uint64("department_id").Optional().Default(0).
			Comment("Department ID of the bound user | 绑定用户的部门ID"),
	}
}

func (OauthAccount) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.StatusMixin{},
		mixins.TenantMixin{}, // 必须包含TenantMixin遵循编码规范
	}
}

func (OauthAccount) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("oauth_accounts").
			Field("user_id").
			Unique().
			Required(),
		edge.From("provider", OauthProvider.Type).
			Ref("oauth_accounts").
			Field("provider_id").
			Unique().
			Required(),
	}
}

func (OauthAccount) Indexes() []ent.Index {
	return []ent.Index{
		// 联合唯一索引：同一用户在同一Provider下只能有一个绑定账户
		index.Fields("user_id", "provider_id", "tenant_id").Unique(),
		// 联合唯一索引：同一Provider下同一第三方用户ID只能绑定一个本地用户
		index.Fields("provider_id", "provider_user_id", "tenant_id").Unique(),
		// 第三方平台用户查询索引
		index.Fields("provider_type", "provider_user_id", "tenant_id"),
		// 用户查询索引
		index.Fields("user_id", "tenant_id"),
		// 状态查询索引
		index.Fields("status", "tenant_id"),
		// 最后登录时间索引
		index.Fields("last_login_at", "tenant_id"),
		// 部门级数据权限查询索引
		index.Fields("department_id", "tenant_id", "status"),
	}
}

func (OauthAccount) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("OAuth Account Binding Table | OAuth账户绑定表"),
		entsql.Annotation{Table: "sys_oauth_accounts"},
	}
}
