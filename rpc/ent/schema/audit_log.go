// Copyright 2024 The NewBee Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/coder-lulu/newbee-common/orm/ent/mixins"
)

// AuditLog holds the schema definition for the AuditLog entity.
type AuditLog struct {
	ent.Schema
}

// Fields of the AuditLog.
func (AuditLog) Fields() []ent.Field {
	return []ent.Field{
		field.String("tenant_id").
			Comment("Tenant ID | 租户ID"),
		field.String("user_id").
			Comment("User ID who performed the operation | 执行操作的用户ID"),
		field.String("user_name").Optional().
			Comment("User name who performed the operation | 执行操作的用户名"),
		field.Enum("operation_type").
			Values("CREATE", "READ", "UPDATE", "DELETE").
			Comment("Operation type | 操作类型"),
		field.String("resource_type").
			Comment("Resource type that was operated on | 被操作的资源类型"),
		field.String("resource_id").Optional().
			Comment("Resource ID that was operated on | 被操作的资源ID"),
		field.String("request_method").
			Comment("HTTP request method | HTTP请求方法"),
		field.String("request_path").
			Comment("HTTP request path | HTTP请求路径"),
		field.Text("request_data").Optional().
			SchemaType(map[string]string{dialect.MySQL: "TEXT"}).
			Comment("Request data (JSON format, sensitive data filtered) | 请求数据(JSON格式，已过滤敏感数据)"),
		field.Int("response_status").
			Comment("HTTP response status code | HTTP响应状态码"),
		field.Text("response_data").Optional().
			SchemaType(map[string]string{dialect.MySQL: "TEXT"}).
			Comment("Response data (JSON format, optional) | 响应数据(JSON格式，可选)"),
		field.String("ip_address").
			Comment("Client IP address | 客户端IP地址"),
		field.String("user_agent").Optional().
			SchemaType(map[string]string{dialect.MySQL: "varchar(512)"}).
			Comment("User agent string | 用户代理字符串"),
		field.Int64("duration_ms").Default(0).
			Comment("Request processing duration in milliseconds | 请求处理耗时(毫秒)"),
		field.Text("error_message").Optional().
			Comment("Error message if operation failed | 操作失败时的错误信息"),
		field.JSON("metadata", map[string]interface{}{}).Optional().
			Comment("Additional metadata in JSON format | 额外的元数据(JSON格式)"),
	}
}

// Mixin of the AuditLog.
func (AuditLog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.UUIDMixin{},
		mixins.StatusMixin{},
	}
}

// Edges of the AuditLog.
func (AuditLog) Edges() []ent.Edge {
	return nil
}

// Indexes of the AuditLog.
func (AuditLog) Indexes() []ent.Index {
	return []ent.Index{
		// 复合索引：租户ID + 用户ID + 创建时间（最常用的查询组合）
		index.Fields("tenant_id", "user_id", "created_at").
			Annotations(entsql.DescColumns("created_at")),

		// 操作类型索引
		index.Fields("tenant_id", "operation_type", "created_at").
			Annotations(entsql.DescColumns("created_at")),

		// 资源类型索引
		index.Fields("tenant_id", "resource_type", "created_at").
			Annotations(entsql.DescColumns("created_at")),

		// 资源ID索引（用于查询特定资源的操作历史）
		index.Fields("tenant_id", "resource_type", "resource_id"),

		// IP地址索引（用于安全分析）
		index.Fields("tenant_id", "ip_address", "created_at").
			Annotations(entsql.DescColumns("created_at")),

		// 响应状态索引（用于错误分析）
		index.Fields("tenant_id", "response_status", "created_at").
			Annotations(entsql.DescColumns("created_at")),

		// 时间范围查询索引
		index.Fields("tenant_id", "created_at").
			Annotations(entsql.DescColumns("created_at")),
	}
}

// Annotations of the AuditLog.
func (AuditLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("Audit Log Table | 审计日志表"),
		entsql.Annotation{Table: "sys_audit_logs"},
	}
}
