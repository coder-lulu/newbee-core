// Copyright 2024 The NewBee Authors. All Rights Reserved.

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/mixins"
)

// CasbinRule holds the schema definition for the CasbinRule entity.
// Casbin权限规则实体，支持租户隔离和企业级权限管理
type CasbinRule struct {
	ent.Schema
}

// Mixin 定义实体的混入，严格遵循CLAUDE.md规则
func (CasbinRule) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},     // ID字段
		mixins.StatusMixin{}, // 状态字段
		mixins.TenantMixin{}, // 🔥 必须包含租户隔离
	}
}

// Fields 定义CasbinRule的字段
func (CasbinRule) Fields() []ent.Field {
	return []ent.Field{
		// Casbin标准字段 - 与Casbin原生适配器兼容
		field.String("ptype").
			Comment("策略类型: p(策略规则), g(角色继承), g2(资源继承)等"),

		field.String("v0").
			Optional().
			Comment("主体: 用户ID、角色代码等"),

		field.String("v1").
			Optional().
			Comment("资源: 资源路径、API端点等"),

		field.String("v2").
			Optional().
			Comment("操作: read, write, delete, create等"),

		field.String("v3").
			Optional().
			Comment("效果: allow, deny"),

		field.Text("v4").
			Optional().
			Comment("条件表达式: JSON格式的复杂条件"),

		field.String("v5").
			Optional().
			Comment("优先级: 数值越大优先级越高"),

		// 业务扩展字段 - 支持企业级功能
		field.String("service_name").
			Comment("服务名称: core, cmdb, workflow等"),

		field.String("rule_name").
			Optional().
			Comment("规则名称: 便于管理和识别"),

		field.Text("description").
			Optional().
			Comment("规则描述: 详细说明规则用途"),

		field.String("category").
			Default("custom").
			Comment("规则分类: system, business, custom等"),

		field.String("version").
			Default("1.0.0").
			Comment("规则版本: 支持规则版本管理"),

		// 审批流程字段 - 企业级权限管理
		field.Bool("require_approval").
			Default(false).
			Comment("是否需要审批: 敏感权限需要审批"),

		field.Enum("approval_status").
			Values("pending", "approved", "rejected").
			Default("approved").
			Comment("审批状态: 权限审批工作流状态"),

		field.Uint64("approved_by").
			Optional().
			Comment("审批人ID: 审批该权限的用户"),

		field.Time("approved_at").
			Optional().
			Comment("审批时间: 权限审批的时间戳"),

		// 时间控制字段 - 临时权限支持
		field.Time("effective_from").
			Optional().
			Comment("生效开始时间: 权限生效的开始时间"),

		field.Time("effective_to").
			Optional().
			Comment("生效结束时间: 权限失效的时间"),

		field.Bool("is_temporary").
			Default(false).
			Comment("是否为临时权限: 支持临时权限管理"),

		// 元数据字段 - 便于管理和监控
		field.Text("metadata").
			Optional().
			Comment("元数据: 扩展属性，JSON格式字符串"),

		field.Text("tags").
			Optional().
			Comment("标签: 权限规则标签，JSON格式字符串"),

		field.Int64("usage_count").
			Default(0).
			Comment("使用次数: 权限规则的使用统计"),

		field.Time("last_used_at").
			Optional().
			Comment("最后使用时间: 权限规则最后一次使用时间"),
	}
}

// Indexes 定义索引，优化查询性能
func (CasbinRule) Indexes() []ent.Index {
	return []ent.Index{
		// 租户+服务查询优化 - 最常用的查询模式
		index.Fields("tenant_id", "service_name", "status"),

		// Casbin策略查询优化 - 权限检查核心索引
		index.Fields("ptype", "v0", "v1"),

		// 权限主体查询 - 按用户/角色查询权限
		index.Fields("v0", "tenant_id", "status"),

		// 审批状态查询 - 审批流程管理
		index.Fields("approval_status", "require_approval"),

		// 时间范围查询 - 临时权限管理
		index.Fields("effective_from", "effective_to"),

		// 规则分类查询 - 按类别管理规则
		index.Fields("category", "service_name"),

		// 复合查询优化 - 权限验证时的完整查询
		index.Fields("tenant_id", "ptype", "v0", "v1", "status"),
	}
}

// Edges 定义关联关系（暂时为空，后续可扩展）
func (CasbinRule) Edges() []ent.Edge {
	return nil
}

// Annotations 定义注解
func (CasbinRule) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// 表注释
		entsql.Annotation{
			Table:   "sys_casbin_rules",
			Options: "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Casbin权限规则表'",
		},
	}
}
