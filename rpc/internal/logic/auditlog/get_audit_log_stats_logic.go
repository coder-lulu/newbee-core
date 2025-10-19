package auditlog

import (
	"context"
	"time"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"
	"github.com/coder-lulu/newbee-core/rpc/ent/auditlog"
	"github.com/coder-lulu/newbee-core/rpc/ent/predicate"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAuditLogStatsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAuditLogStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAuditLogStatsLogic {
	return &GetAuditLogStatsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetAuditLogStatsLogic) GetAuditLogStats(in *core.AuditLogStatsReq) (*core.AuditLogStatsResp, error) {
	var predicates []predicate.AuditLog

	// Build time range predicates
	if in.StartTime != nil {
		startTime := time.Unix(*in.StartTime, 0)
		predicates = append(predicates, auditlog.CreatedAtGTE(startTime))
	}
	if in.EndTime != nil {
		endTime := time.Unix(*in.EndTime, 0)
		predicates = append(predicates, auditlog.CreatedAtLTE(endTime))
	}

	// Build filtering predicates
	if in.OperationType != nil {
		predicates = append(predicates, auditlog.OperationTypeEQ(auditlog.OperationType(*in.OperationType)))
	}
	if in.ResourceType != nil {
		predicates = append(predicates, auditlog.ResourceTypeEQ(*in.ResourceType))
	}
	if in.UserId != nil {
		predicates = append(predicates, auditlog.UserIDEQ(*in.UserId))
	}

	// Get total operations count
	totalOperations, err := l.svcCtx.DB.AuditLog.Query().Where(predicates...).Count(l.ctx)
	if err != nil {
		logx.Errorw("Failed to count audit log operations", logx.Field("error", err))
		return nil, err
	}

	// Get unique users count - simplified approach
	uniqueUsers, err := l.svcCtx.DB.AuditLog.Query().Where(predicates...).Select(auditlog.FieldUserID).All(l.ctx)
	if err != nil {
		logx.Errorw("Failed to get user records", logx.Field("error", err))
		return nil, err
	}
	// Count unique user IDs manually
	userSet := make(map[string]bool)
	for _, record := range uniqueUsers {
		if record.UserID != "" {
			userSet[record.UserID] = true
		}
	}
	uniqueUserCount := len(userSet)

	// Get error count (response status >= 400)
	errorPredicates := append(predicates, auditlog.ResponseStatusGTE(400))
	errorCount, err := l.svcCtx.DB.AuditLog.Query().Where(errorPredicates...).Count(l.ctx)
	if err != nil {
		logx.Errorw("Failed to count errors", logx.Field("error", err))
		return nil, err
	}

	// Calculate average duration
	avgDuration := 0.0
	if totalOperations > 0 {
		// Get all duration values and calculate average manually
		records, err := l.svcCtx.DB.AuditLog.Query().Where(predicates...).Select(auditlog.FieldDurationMs).All(l.ctx)
		if err != nil {
			logx.Errorw("Failed to get duration records", logx.Field("error", err))
		} else if len(records) > 0 {
			sum := int64(0)
			for _, record := range records {
				sum += record.DurationMs
			}
			avgDuration = float64(sum) / float64(len(records))
		}
	}

	// Get operation type statistics
	operationStats, err := l.getOperationTypeStats(predicates, uint64(totalOperations))
	if err != nil {
		logx.Errorw("Failed to get operation type stats", logx.Field("error", err))
		return nil, err
	}

	// Get resource type statistics
	resourceStats, err := l.getResourceTypeStats(predicates, uint64(totalOperations))
	if err != nil {
		logx.Errorw("Failed to get resource type stats", logx.Field("error", err))
		return nil, err
	}

	// Get duration statistics
	durationStats, err := l.getDurationStats(predicates, uint64(totalOperations))
	if err != nil {
		logx.Errorw("Failed to get duration stats", logx.Field("error", err))
		return nil, err
	}

	return &core.AuditLogStatsResp{
		TotalOperations: uint64(totalOperations),
		UniqueUsers:     uint64(uniqueUserCount),
		ErrorCount:      uint64(errorCount),
		AvgDurationMs:   avgDuration,
		OperationStats:  operationStats,
		ResourceStats:   resourceStats,
		DurationStats:   durationStats,
	}, nil
}

func (l *GetAuditLogStatsLogic) getOperationTypeStats(basePredicates []predicate.AuditLog, total uint64) ([]*core.OperationTypeStats, error) {
	// Get operation type counts manually
	operationTypes := []auditlog.OperationType{
		auditlog.OperationTypeCREATE,
		auditlog.OperationTypeREAD,
		auditlog.OperationTypeUPDATE,
		auditlog.OperationTypeDELETE,
	}

	var stats []*core.OperationTypeStats
	for _, opType := range operationTypes {
		predicates := append(basePredicates, auditlog.OperationTypeEQ(opType))
		count, err := l.svcCtx.DB.AuditLog.Query().Where(predicates...).Count(l.ctx)
		if err != nil {
			return nil, err
		}
		
		if count > 0 {
			percentage := 0.0
			if total > 0 {
				percentage = float64(count) / float64(total) * 100.0
			}
			
			stats = append(stats, &core.OperationTypeStats{
				OperationType: string(opType),
				Count:         uint64(count),
				Percentage:    percentage,
			})
		}
	}
	
	return stats, nil
}

func (l *GetAuditLogStatsLogic) getResourceTypeStats(basePredicates []predicate.AuditLog, total uint64) ([]*core.ResourceTypeStats, error) {
	// Get unique resource types first
	resourceTypes, err := l.svcCtx.DB.AuditLog.Query().
		Where(basePredicates...).
		Select(auditlog.FieldResourceType).
		All(l.ctx)
	if err != nil {
		return nil, err
	}
	
	// Get unique resource types manually
	resourceTypeSet := make(map[string]bool)
	for _, record := range resourceTypes {
		if record.ResourceType != "" {
			resourceTypeSet[record.ResourceType] = true
		}
	}

	var stats []*core.ResourceTypeStats
	for resourceType := range resourceTypeSet {
		predicates := append(basePredicates, auditlog.ResourceTypeEQ(resourceType))
		count, err := l.svcCtx.DB.AuditLog.Query().Where(predicates...).Count(l.ctx)
		if err != nil {
			return nil, err
		}
		
		if count > 0 {
			percentage := 0.0
			if total > 0 {
				percentage = float64(count) / float64(total) * 100.0
			}
			
			stats = append(stats, &core.ResourceTypeStats{
				ResourceType: resourceType,
				Count:        uint64(count),
				Percentage:   percentage,
			})
		}
	}
	
	return stats, nil
}

func (l *GetAuditLogStatsLogic) getDurationStats(basePredicates []predicate.AuditLog, total uint64) ([]*core.DurationStats, error) {
	// Define duration ranges in milliseconds
	ranges := []struct {
		label string
		min   int64
		max   int64
	}{
		{"0-100ms", 0, 100},
		{"100-500ms", 100, 500},
		{"500ms-1s", 500, 1000},
		{"1s-5s", 1000, 5000},
		{"5s+", 5000, 999999999},
	}

	var stats []*core.DurationStats
	for _, r := range ranges {
		predicates := append(basePredicates, 
			auditlog.DurationMsGTE(r.min),
			auditlog.DurationMsLT(r.max),
		)
		
		count, err := l.svcCtx.DB.AuditLog.Query().Where(predicates...).Count(l.ctx)
		if err != nil {
			return nil, err
		}
		
		percentage := 0.0
		if total > 0 {
			percentage = float64(count) / float64(total) * 100.0
		}
		
		stats = append(stats, &core.DurationStats{
			RangeLabel: r.label,
			Count:      uint64(count),
			Percentage: percentage,
		})
	}
	
	return stats, nil
}
