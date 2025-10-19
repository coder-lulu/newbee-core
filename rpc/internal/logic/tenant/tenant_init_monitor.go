package tenant

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/ent/auditlog"
)

// InitializationProgress 初始化进度跟踪
type InitializationProgress struct {
	TenantID      uint64                   `json:"tenant_id"`
	StartedAt     time.Time                `json:"started_at"`
	CompletedAt   *time.Time               `json:"completed_at,omitempty"`
	Status        string                   `json:"status"` // initializing, completed, failed
	TotalSteps    int                      `json:"total_steps"`
	CurrentStep   int                      `json:"current_step"`
	CurrentAction string                   `json:"current_action"`
	Components    []ComponentProgress      `json:"components"`
	Errors        []InitializationError    `json:"errors,omitempty"`
	Metrics       InitializationMetrics    `json:"metrics"`
}

// ComponentProgress 组件初始化进度
type ComponentProgress struct {
	Name        string     `json:"name"`
	Status      string     `json:"status"` // pending, running, completed, failed
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	RecordsCount int       `json:"records_count"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// InitializationError 初始化错误
type InitializationError struct {
	Component string    `json:"component"`
	Message   string    `json:"message"`
	Details   string    `json:"details"`
	Timestamp time.Time `json:"timestamp"`
}

// InitializationMetrics 初始化指标
type InitializationMetrics struct {
	Duration       time.Duration `json:"duration"`
	TotalRecords   int           `json:"total_records"`
	SuccessCount   int           `json:"success_count"`
	FailureCount   int           `json:"failure_count"`
	ComponentStats map[string]ComponentMetrics `json:"component_stats"`
}

// ComponentMetrics 组件指标
type ComponentMetrics struct {
	Duration     time.Duration `json:"duration"`
	RecordCount  int           `json:"record_count"`
	RetryCount   int           `json:"retry_count"`
}

// ProgressTracker 进度跟踪器
type ProgressTracker struct {
	db       *ent.Client
	logger   interface{ Infow(msg string, keysAndValues ...interface{}) }
	progress *InitializationProgress
}

// NewProgressTracker 创建进度跟踪器
func NewProgressTracker(db *ent.Client, logger interface{ Infow(msg string, keysAndValues ...interface{}) }, tenantID uint64) *ProgressTracker {
	return &ProgressTracker{
		db:     db,
		logger: logger,
		progress: &InitializationProgress{
			TenantID:    tenantID,
			StartedAt:   time.Now(),
			Status:      "initializing",
			Components:  make([]ComponentProgress, 0),
			Errors:      make([]InitializationError, 0),
			Metrics:     InitializationMetrics{ComponentStats: make(map[string]ComponentMetrics)},
		},
	}
}

// StartComponent 开始组件初始化
func (pt *ProgressTracker) StartComponent(name string) {
	now := time.Now()
	pt.progress.CurrentAction = fmt.Sprintf("正在初始化 %s", name)
	pt.progress.CurrentStep++
	
	// 查找现有组件或添加新组件
	found := false
	for i, comp := range pt.progress.Components {
		if comp.Name == name {
			pt.progress.Components[i].Status = "running"
			pt.progress.Components[i].StartedAt = &now
			found = true
			break
		}
	}
	
	if !found {
		pt.progress.Components = append(pt.progress.Components, ComponentProgress{
			Name:      name,
			Status:    "running",
			StartedAt: &now,
		})
	}
	
	pt.logger.Infow("开始初始化组件", 
		"tenant_id", pt.progress.TenantID,
		"component", name,
		"step", pt.progress.CurrentStep,
		"total", pt.progress.TotalSteps)
}

// CompleteComponent 完成组件初始化
func (pt *ProgressTracker) CompleteComponent(name string, recordsCount int) {
	now := time.Now()
	
	for i, comp := range pt.progress.Components {
		if comp.Name == name {
			pt.progress.Components[i].Status = "completed"
			pt.progress.Components[i].CompletedAt = &now
			pt.progress.Components[i].RecordsCount = recordsCount
			
			// 计算组件耗时
			if comp.StartedAt != nil {
				duration := now.Sub(*comp.StartedAt)
				pt.progress.Metrics.ComponentStats[name] = ComponentMetrics{
					Duration:    duration,
					RecordCount: recordsCount,
				}
			}
			break
		}
	}
	
	pt.progress.Metrics.SuccessCount++
	pt.progress.Metrics.TotalRecords += recordsCount
	
	pt.logger.Infow("组件初始化完成", 
		"tenant_id", pt.progress.TenantID,
		"component", name,
		"records_count", recordsCount)
}

// FailComponent 组件初始化失败
func (pt *ProgressTracker) FailComponent(name string, err error) {
	now := time.Now()
	
	for i, comp := range pt.progress.Components {
		if comp.Name == name {
			pt.progress.Components[i].Status = "failed"
			pt.progress.Components[i].CompletedAt = &now
			pt.progress.Components[i].ErrorMessage = err.Error()
			break
		}
	}
	
	pt.progress.Errors = append(pt.progress.Errors, InitializationError{
		Component: name,
		Message:   err.Error(),
		Timestamp: now,
	})
	
	pt.progress.Metrics.FailureCount++
	
	pt.logger.Infow("组件初始化失败", 
		"tenant_id", pt.progress.TenantID,
		"component", name,
		"error", err.Error())
}

// SetTotalSteps 设置总步数
func (pt *ProgressTracker) SetTotalSteps(total int) {
	pt.progress.TotalSteps = total
}

// Complete 完成初始化
func (pt *ProgressTracker) Complete() {
	now := time.Now()
	pt.progress.CompletedAt = &now
	pt.progress.Status = "completed"
	pt.progress.CurrentAction = "初始化完成"
	pt.progress.Metrics.Duration = now.Sub(pt.progress.StartedAt)
	
	pt.logger.Infow("租户初始化完成", 
		"tenant_id", pt.progress.TenantID,
		"duration", pt.progress.Metrics.Duration,
		"total_records", pt.progress.Metrics.TotalRecords,
		"success_count", pt.progress.Metrics.SuccessCount,
		"failure_count", pt.progress.Metrics.FailureCount)
}

// Fail 初始化失败
func (pt *ProgressTracker) Fail(err error) {
	now := time.Now()
	pt.progress.CompletedAt = &now
	pt.progress.Status = "failed"
	pt.progress.CurrentAction = fmt.Sprintf("初始化失败: %s", err.Error())
	pt.progress.Metrics.Duration = now.Sub(pt.progress.StartedAt)
	
	pt.progress.Errors = append(pt.progress.Errors, InitializationError{
		Component: "system",
		Message:   err.Error(),
		Timestamp: now,
	})
	
	pt.logger.Infow("租户初始化失败", 
		"tenant_id", pt.progress.TenantID,
		"error", err.Error(),
		"duration", pt.progress.Metrics.Duration)
}

// GetProgress 获取当前进度
func (pt *ProgressTracker) GetProgress() *InitializationProgress {
	return pt.progress
}

// SaveAuditLog 保存审计日志
func (pt *ProgressTracker) SaveAuditLog(ctx context.Context) error {
	progressData, err := json.Marshal(pt.progress)
	if err != nil {
		return fmt.Errorf("序列化进度数据失败: %w", err)
	}
	
	_, err = pt.db.AuditLog.Create().
		SetResourceType("tenant_initialization").
		SetResourceID(fmt.Sprintf("%d", pt.progress.TenantID)).
		SetOperationType("CREATE").
		SetRequestMethod("POST").
		SetRequestPath("/api/v1/tenant/init").
		SetResponseStatus(getAuditResponseStatus(pt.progress.Status)).
		SetTenantID(fmt.Sprintf("%d", pt.progress.TenantID)).
		SetUserID("system").
		SetUserName("System").
		SetIPAddress("127.0.0.1").
		SetDurationMs(int64(pt.progress.Metrics.Duration.Milliseconds())).
		SetResponseData(string(progressData)).
		Save(ctx)
	
	if err != nil {
		return fmt.Errorf("保存审计日志失败: %w", err)
	}
	
	return nil
}

// getAuditResponseStatus 获取审计响应状态
func getAuditResponseStatus(status string) int {
	switch status {
	case "completed":
		return 200 // HTTP OK
	case "failed":
		return 500 // HTTP Internal Server Error
	default:
		return 202 // HTTP Accepted (in progress)
	}
}

// InitializationMonitor 初始化监控器
type InitializationMonitor struct {
	db     *ent.Client
	logger interface{ Infow(msg string, keysAndValues ...interface{}) }
}

// NewInitializationMonitor 创建初始化监控器
func NewInitializationMonitor(db *ent.Client, logger interface{ Infow(msg string, keysAndValues ...interface{}) }) *InitializationMonitor {
	return &InitializationMonitor{
		db:     db,
		logger: logger,
	}
}

// GetInitializationHistory 获取初始化历史
func (im *InitializationMonitor) GetInitializationHistory(ctx context.Context, tenantID uint64, limit int) ([]InitializationProgress, error) {
	logs, err := im.db.AuditLog.Query().
		Where(
			auditlog.ResourceTypeEQ("tenant_initialization"),
			auditlog.ResourceIDEQ(fmt.Sprintf("%d", tenantID)),
		).
		Order(ent.Desc(auditlog.FieldCreatedAt)).
		Limit(limit).
		All(ctx)
	
	if err != nil {
		return nil, fmt.Errorf("查询初始化历史失败: %w", err)
	}
	
	var history []InitializationProgress
	for _, log := range logs {
		if log.ResponseData == "" {
			continue
		}
		
		var progress InitializationProgress
		if err := json.Unmarshal([]byte(log.ResponseData), &progress); err != nil {
			im.logger.Infow("解析历史记录失败", "log_id", log.ID, "error", err.Error())
			continue
		}
		
		history = append(history, progress)
	}
	
	return history, nil
}

// GetSystemInitializationStats 获取系统初始化统计
func (im *InitializationMonitor) GetSystemInitializationStats(ctx context.Context, days int) (*SystemInitializationStats, error) {
	since := time.Now().AddDate(0, 0, -days)
	
	// 获取指定时间范围内的初始化记录
	logs, err := im.db.AuditLog.Query().
		Where(
			auditlog.ResourceTypeEQ("tenant_initialization"),
			auditlog.CreatedAtGTE(since),
		).
		All(ctx)
	
	if err != nil {
		return nil, fmt.Errorf("查询初始化统计失败: %w", err)
	}
	
	stats := &SystemInitializationStats{
		TotalInitializations: len(logs),
		SuccessCount:         0,
		FailureCount:         0,
		AverageDuration:      0,
		ComponentStats:       make(map[string]ComponentStats),
	}
	
	var totalDuration time.Duration
	componentCounts := make(map[string]int)
	
	for _, log := range logs {
		if log.ResponseData == "" {
			continue
		}
		
		var progress InitializationProgress
		if err := json.Unmarshal([]byte(log.ResponseData), &progress); err != nil {
			continue
		}
		
		if progress.Status == "completed" {
			stats.SuccessCount++
		} else if progress.Status == "failed" {
			stats.FailureCount++
		}
		
		totalDuration += progress.Metrics.Duration
		
		// 统计组件
		for _, comp := range progress.Components {
			componentCounts[comp.Name]++
			if compStats, exists := stats.ComponentStats[comp.Name]; exists {
				compStats.TotalUsage++
				if comp.Status == "completed" {
					compStats.SuccessRate = float64(compStats.SuccessRate*float64(compStats.TotalUsage-1)+1) / float64(compStats.TotalUsage)
				}
				stats.ComponentStats[comp.Name] = compStats
			} else {
				successRate := 0.0
				if comp.Status == "completed" {
					successRate = 1.0
				}
				stats.ComponentStats[comp.Name] = ComponentStats{
					TotalUsage:  1,
					SuccessRate: successRate,
				}
			}
		}
	}
	
	if stats.TotalInitializations > 0 {
		stats.AverageDuration = totalDuration / time.Duration(stats.TotalInitializations)
		stats.SuccessRate = float64(stats.SuccessCount) / float64(stats.TotalInitializations) * 100
	}
	
	return stats, nil
}

// SystemInitializationStats 系统初始化统计
type SystemInitializationStats struct {
	TotalInitializations int                      `json:"total_initializations"`
	SuccessCount         int                      `json:"success_count"`
	FailureCount         int                      `json:"failure_count"`
	SuccessRate          float64                  `json:"success_rate"`
	AverageDuration      time.Duration            `json:"average_duration"`
	ComponentStats       map[string]ComponentStats `json:"component_stats"`
}

// ComponentStats 组件统计
type ComponentStats struct {
	TotalUsage  int     `json:"total_usage"`
	SuccessRate float64 `json:"success_rate"`
}