// Copyright 2024 The NewBee Authors. All Rights Reserved.

package contextx

import (
	"context"

	"github.com/coder-lulu/newbee-common/middleware/keys"
	"google.golang.org/grpc/metadata"
)

// ExtractFromContext 从 context 或 gRPC metadata 中提取指定 key 的值
// 优先从 context 中获取，如果没有则从 gRPC metadata 中提取
func ExtractFromContext(ctx context.Context, key keys.ContextKey) string {
	// 1. 优先从 context 中获取
	if val, ok := ctx.Value(key).(string); ok && val != "" {
		return val
	}

	// 2. 从 gRPC metadata 中提取
	if md, mdOK := metadata.FromIncomingContext(ctx); mdOK {
		// 尝试原始key
		if vals := md.Get(string(key)); len(vals) > 0 {
			return vals[0]
		}
		// 尝试小写key（gRPC会自动转小写）
		if vals := md.Get(string(key)); len(vals) > 0 {
			return vals[0]
		}
	}

	return ""
}

// ExtractTenantID 提取租户ID
func ExtractTenantID(ctx context.Context) string {
	return ExtractFromContext(ctx, keys.TenantIDKey)
}

// ExtractUserID 提取用户ID
func ExtractUserID(ctx context.Context) string {
	return ExtractFromContext(ctx, keys.UserIDKey)
}

// ExtractDeptID 提取部门ID
func ExtractDeptID(ctx context.Context) string {
	return ExtractFromContext(ctx, keys.DeptIDKey)
}

// ExtractRoleCodes 提取角色代码
func ExtractRoleCodes(ctx context.Context) string {
	return ExtractFromContext(ctx, keys.RoleCodesKey)
}

// ExtractDataScope 提取数据权限范围
func ExtractDataScope(ctx context.Context) string {
	return ExtractFromContext(ctx, keys.DataScopeKey)
}

// ExtractAllAuthInfo 提取所有认证信息
func ExtractAllAuthInfo(ctx context.Context) map[string]string {
	return map[string]string{
		"tenant_id":  ExtractTenantID(ctx),
		"user_id":    ExtractUserID(ctx),
		"dept_id":    ExtractDeptID(ctx),
		"role_codes": ExtractRoleCodes(ctx),
		"data_scope": ExtractDataScope(ctx),
	}
}
