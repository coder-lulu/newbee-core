// Copyright 2024 The NewBee Authors. All Rights Reserved.

package svc

import (
	"context"
	"strings"

	"github.com/coder-lulu/newbee-common/middleware/audit"
	"github.com/coder-lulu/newbee-common/middleware/framework"
	"github.com/coder-lulu/newbee-core/rpc/coreclient"
	"github.com/zeromicro/go-zero/core/logx"
)

// HighPerformanceCoreAuditWriter 高性能Core审计写入器
// 使用适配器消除反射调用，提升性能10-100倍
type HighPerformanceCoreAuditWriter struct {
	auditClient   audit.AuditRPCClient
	healthChecker func() bool
}

// NewHighPerformanceCoreAuditWriter 创建高性能Core审计写入器
func NewHighPerformanceCoreAuditWriter(coreClient coreclient.Core, healthChecker func() bool) *HighPerformanceCoreAuditWriter {
	return &HighPerformanceCoreAuditWriter{
		auditClient:   audit.NewCoreClientAdapter(coreClient),
		healthChecker: healthChecker,
	}
}

// WriteAuditLog 实现framework.AuditWriter接口
// 使用适配器进行直接调用，避免反射带来的性能损失
func (w *HighPerformanceCoreAuditWriter) WriteAuditLog(ctx context.Context, auditData framework.AuditLogData) error {
	// 1. 循环调用检测：检查是否是审计相关的API调用
	if w.isAuditRelatedCall(auditData.Path) {
		logx.WithContext(ctx).Infow("Skipping audit log for audit-related API call (preventing recursion)",
			logx.Field("path", auditData.Path),
			logx.Field("method", auditData.Method))
		return nil // 静默跳过，防止递归调用
	}

	// 2. 检查RPC健康状态
	if w.healthChecker != nil && !w.healthChecker() {
		logx.WithContext(ctx).Errorw("Core RPC is unhealthy, audit log will be lost",
			logx.Field("path", auditData.Path),
			logx.Field("method", auditData.Method))
		return nil // RPC不健康时静默失败
	}

	// 3. 审计客户端检查
	if w.auditClient == nil {
		logx.WithContext(ctx).Error("Audit client is nil, audit log will be lost")
		return nil
	}

	// 4. 转换中间件格式为审计客户端格式
	resourceType := auditData.ResourceType
	if resourceType == "" {
		resourceType = auditData.ResourceName
	}
	if resourceType == "" {
		resourceType = auditData.Path
	}

	resourceID := auditData.ResourceID
	if resourceID == "" {
		resourceID = auditData.Path
	}

	userName := auditData.UserName
	if userName == "" {
		userName = auditData.UserID
	}

	auditInfo := &audit.AuditLogInfo{
		TenantID:     auditData.TenantID,
		UserID:       auditData.UserID,
		UserName:     userName,
		Method:       auditData.Method,
		Path:         auditData.Path,
		ResourceName: auditData.ResourceName,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		StatusCode:   auditData.StatusCode,
		DurationMs:   auditData.DurationMs,
		UserAgent:    auditData.UserAgent,
		ClientIP:     auditData.ClientIP,
		RequestData:  auditData.RequestData,
		ResponseData: auditData.ResponseData,
		Metadata:     auditData.Metadata,
	}

	// 5. 直接调用审计客户端（无反射，高性能）
	result, err := w.auditClient.CreateAuditLog(ctx, auditInfo)
	if err != nil {
		logx.Errorw("High-performance audit log creation failed",
			logx.Field("error", err),
			logx.Field("tenant_id", auditData.TenantID),
			logx.Field("user_id", auditData.UserID),
			logx.Field("path", auditData.Path))
		return err
	}

	// 6. 检查RPC调用结果
	if !result.Success {
		logx.WithContext(ctx).Errorw("RPC audit log creation returned failure",
			logx.Field("message", result.Message),
			logx.Field("path", auditData.Path))
		return nil // 不返回错误，避免影响主业务
	}

	logx.Infow("High-performance audit log created successfully",
		logx.Field("audit_id", result.AuditID),
		logx.Field("tenant_id", auditData.TenantID),
		logx.Field("user_id", auditData.UserID),
		logx.Field("path", auditData.Path))

	return nil
}

// isAuditRelatedCall 检查是否是审计相关的API调用
// 防止审计API调用自身时产生无限递归
func (w *HighPerformanceCoreAuditWriter) isAuditRelatedCall(path string) bool {
	auditPaths := []string{
		"/audit-log/", // 审计日志相关API
		"/audit_log/", // 下划线版本
		"audit-log",   // 不带斜杠的版本
		"audit_log",   // 不带斜杠的下划线版本
	}

	for _, auditPath := range auditPaths {
		if strings.Contains(strings.ToLower(path), auditPath) {
			return true
		}
	}
	return false
}
