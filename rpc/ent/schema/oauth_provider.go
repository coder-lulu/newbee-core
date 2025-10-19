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

type OauthProvider struct {
	ent.Schema
}

func (OauthProvider) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").MaxLen(50).
			Comment("The provider's name | 提供商名称"),
		field.String("display_name").MaxLen(100).Optional().
			Comment("Display name for UI | UI显示名称"),
		field.String("type").MaxLen(20).
			Comment("The provider type (wechat, qq, github, google, facebook) | 提供商类型"),
		field.String("provider_type").MaxLen(20).Default("oauth2").
			Comment("Provider type: oauth2, oidc, saml | 认证协议类型"),
		field.String("client_id").MaxLen(255).
			Comment("The client id | 客户端 ID"),
		field.String("client_secret").MaxLen(1000).
			Comment("The client secret (original) | 客户端密钥（原始）"),
		field.String("encrypted_secret").MaxLen(1000).Optional().
			Comment("Encrypted client secret | 加密后的客户端密钥"),
		field.String("encryption_key_id").MaxLen(50).Optional().
			Comment("Encryption key ID | 加密密钥ID"),
		field.String("redirect_url").MaxLen(500).
			Comment("The redirect url | 回调地址"),
		field.String("scopes").MaxLen(500).Optional().
			Comment("The scopes | 权限范围"),
		// 兼容性字段：保持与现有逻辑的兼容
		field.String("auth_url").MaxLen(500).
			Comment("OAuth authorization URL | OAuth授权URL"),
		field.String("token_url").MaxLen(500).
			Comment("OAuth token exchange URL | OAuth令牌交换URL"),
		field.String("info_url").MaxLen(500).
			Comment("User info URL | 用户信息获取URL"),
		field.Int("auth_style").Default(2).
			Comment("OAuth auth style (1=params, 2=header) | OAuth认证方式"),
		field.JSON("extra_config", map[string]interface{}{}).Optional().
			Comment("Provider specific configuration | 提供商特定配置"),
		field.Bool("enabled").Default(true).
			Comment("Whether the provider is enabled | 是否启用"),
		field.Uint32("sort").Default(0).
			Comment("Sort order | 排序"),
		field.String("remark").MaxLen(200).Optional().
			Comment("Remark | 备注"),
		field.Bool("support_pkce").Default(false).
			Comment("Whether support PKCE | 是否支持PKCE"),
		field.String("icon_url").MaxLen(500).Optional().
			Comment("Provider icon URL | 提供商图标URL"),
		// 性能优化字段
		field.Int("cache_ttl").Default(3600).
			Comment("Configuration cache TTL in seconds | 配置缓存TTL(秒)"),
		field.String("webhook_url").MaxLen(500).Optional().
			Comment("Webhook URL for configuration updates | 配置更新webhook地址"),
		// 监控字段
		field.Int("success_count").Default(0).
			Comment("Successful OAuth attempts count | 成功登录次数"),
		field.Int("failure_count").Default(0).
			Comment("Failed OAuth attempts count | 失败登录次数"),
		field.Time("last_used_at").Optional().
			Comment("Last used timestamp | 最后使用时间"),
	}
}

func (OauthProvider) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.StatusMixin{},
		mixins.TenantMixin{}, // 必须包含TenantMixin遵循编码规范
	}
}

func (OauthProvider) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("oauth_accounts", OauthAccount.Type),
		edge.To("oauth_sessions", OauthSession.Type),
	}
}

func (OauthProvider) Indexes() []ent.Index {
	return []ent.Index{
		// 名称唯一索引（全局唯一）
		index.Fields("name").Unique(),
		// 联合唯一索引：同一租户下同一类型的Provider只能有一个
		index.Fields("type", "tenant_id").Unique(),
		// 启用状态索引，用于快速查询启用的Provider
		index.Fields("enabled", "tenant_id"),
		// 状态索引，用于快速查询正常状态的Provider
		index.Fields("status", "tenant_id"),
		// 认证协议类型索引
		index.Fields("provider_type", "tenant_id"),
		// 排序索引
		index.Fields("sort", "tenant_id"),
		// 最后使用时间索引，用于监控和清理
		index.Fields("last_used_at", "tenant_id"),
		// 复合索引：启用且正常状态的Provider
		index.Fields("enabled", "status", "tenant_id"),
	}
}

func (OauthProvider) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("OAuth Provider Configuration Table | OAuth第三方登录提供商配置表"),
		entsql.Annotation{Table: "sys_oauth_providers"},
	}
}
