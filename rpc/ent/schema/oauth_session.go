package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	uuid "github.com/gofrs/uuid/v5"

	"github.com/coder-lulu/newbee-common/v2/orm/ent/mixins"
)

type OauthSession struct {
	ent.Schema
}

func (OauthSession) Fields() []ent.Field {
	return []ent.Field{
		field.String("session_id").MaxLen(128).Unique().
			Comment("Unique session identifier | 唯一会话标识符"),
		field.String("state").MaxLen(256).
			Comment("OAuth state parameter | OAuth state参数"),
		field.Uint64("provider_id").
			Comment("OAuth provider ID | OAuth提供商ID"),
		field.UUID("user_id", uuid.UUID{}).Optional().
			Comment("Associated user ID (if logged in) | 关联的用户ID（如果已登录）"),
		field.String("redirect_uri").MaxLen(500).
			Comment("OAuth redirect URI | OAuth回调地址"),
		field.String("scope").MaxLen(500).Optional().
			Comment("Requested OAuth scopes | 请求的OAuth权限范围"),
		// PKCE支持字段
		field.String("code_challenge").MaxLen(128).Optional().
			Comment("PKCE code challenge | PKCE代码挑战"),
		field.String("code_challenge_method").MaxLen(10).Optional().
			Comment("PKCE code challenge method | PKCE代码挑战方法"),
		field.String("code_verifier").MaxLen(128).Optional().
			Comment("PKCE code verifier | PKCE代码验证器"),
		// 会话状态管理使用StatusMixin提供的status字段
		field.Time("expires_at").
			Comment("Session expiration time | 会话过期时间"),
		field.String("client_ip").MaxLen(45).Optional().
			Comment("Client IP address | 客户端IP地址"),
		field.String("user_agent").MaxLen(500).Optional().
			Comment("Client user agent | 客户端用户代理"),
		// OAuth回调数据
		field.String("authorization_code").MaxLen(512).Optional().
			Comment("OAuth authorization code | OAuth授权码"),
		field.Time("code_received_at").Optional().
			Comment("Authorization code received time | 授权码接收时间"),
		field.JSON("callback_data", map[string]interface{}{}).Optional().
			Comment("OAuth callback additional data | OAuth回调额外数据"),
		// 错误信息
		field.String("error_code").MaxLen(50).Optional().
			Comment("OAuth error code | OAuth错误码"),
		field.String("error_description").MaxLen(500).Optional().
			Comment("OAuth error description | OAuth错误描述"),
		field.Int("retry_count").Default(0).
			Comment("Retry attempts count | 重试次数"),
		field.Uint64("department_id").Optional().Default(0).
			Comment("Department ID when session was created | 会话创建时的部门ID"),
	}
}

func (OauthSession) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.StatusMixin{},
		mixins.TenantMixin{}, // 必须包含TenantMixin遵循编码规范
	}
}

func (OauthSession) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("provider", OauthProvider.Type).
			Ref("oauth_sessions").
			Field("provider_id").
			Unique().
			Required(),
		edge.From("user", User.Type).
			Ref("oauth_sessions").
			Field("user_id").
			Unique(),
	}
}

func (OauthSession) Indexes() []ent.Index {
	return []ent.Index{
		// 会话ID唯一索引
		index.Fields("session_id").Unique(),
		// State参数索引（用于CSRF防护验证）
		index.Fields("state", "tenant_id"),
		// Provider会话查询索引
		index.Fields("provider_id", "tenant_id"),
		// 用户会话查询索引
		index.Fields("user_id", "tenant_id"),
		// 状态查询索引
		index.Fields("status", "tenant_id"),
		// 过期时间索引（用于清理过期会话）
		index.Fields("expires_at", "tenant_id"),
		// 授权码索引
		index.Fields("authorization_code", "tenant_id"),
		// IP地址索引（用于安全分析）
		index.Fields("client_ip", "tenant_id"),
		// 复合索引：活跃会话查询
		index.Fields("status", "expires_at", "tenant_id"),
		// 部门级数据权限查询索引
		index.Fields("department_id", "tenant_id", "status"),
		// 部门+过期时间复合索引（用于清理）
		index.Fields("department_id", "expires_at", "tenant_id"),
	}
}

func (OauthSession) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("OAuth Session Management Table | OAuth会话管理表"),
		entsql.Annotation{Table: "sys_oauth_sessions"},
	}
}
