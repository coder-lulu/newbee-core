package oauthprovider

import (
	"context"
	"time"

	"github.com/coder-lulu/newbee-common/i18n"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOauthStatisticsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOauthStatisticsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOauthStatisticsLogic {
	return &GetOauthStatisticsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOauthStatisticsLogic) GetOauthStatistics(req *types.OauthStatisticsReq) (resp *types.OauthStatisticsResp, err error) {
	// 生成模拟统计数据
	// 在实际项目中，这里应该从数据库查询真实数据

	// 生成提供商统计数据
	providerStats := []types.ProviderStatData{
		{
			ProviderId:      1,
			ProviderName:    "google",
			DisplayName:     "Google",
			Type:            "google",
			IconUrl:         stringPtr(""),
			TotalUsage:      2150,
			SuccessCount:    2105,
			FailureCount:    45,
			SuccessRate:     97.9,
			AvgResponseTime: 85,
			LastUsed:        int64Ptr(time.Now().Unix()),
		},
		{
			ProviderId:      2,
			ProviderName:    "wechat",
			DisplayName:     "微信",
			Type:            "wechat",
			IconUrl:         stringPtr(""),
			TotalUsage:      1890,
			SuccessCount:    1812,
			FailureCount:    78,
			SuccessRate:     95.9,
			AvgResponseTime: 120,
			LastUsed:        int64Ptr(time.Now().Unix()),
		},
		{
			ProviderId:      3,
			ProviderName:    "github",
			DisplayName:     "GitHub",
			Type:            "github",
			IconUrl:         stringPtr(""),
			TotalUsage:      1245,
			SuccessCount:    1222,
			FailureCount:    23,
			SuccessRate:     98.2,
			AvgResponseTime: 95,
			LastUsed:        int64Ptr(time.Now().Unix()),
		},
		{
			ProviderId:      4,
			ProviderName:    "qq",
			DisplayName:     "QQ",
			Type:            "qq",
			IconUrl:         stringPtr(""),
			TotalUsage:      856,
			SuccessCount:    822,
			FailureCount:    34,
			SuccessRate:     96.0,
			AvgResponseTime: 110,
			LastUsed:        int64Ptr(time.Now().Unix()),
		},
	}

	// 生成登录趋势数据
	var loginTrend []types.LoginTrendData
	now := time.Now()
	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		loginTrend = append(loginTrend, types.LoginTrendData{
			Date:         date.Format("2006-01-02"),
			Count:        int64(200 + i*20),
			SuccessCount: int64(190 + i*19),
			FailureCount: int64(10 + i),
		})
	}

	// 计算总体统计
	var totalLogins, totalSuccessLogins, totalFailureLogins int64
	for _, stat := range providerStats {
		totalLogins += stat.TotalUsage
		totalSuccessLogins += stat.SuccessCount
		totalFailureLogins += stat.FailureCount
	}

	successRate := float64(totalSuccessLogins) / float64(totalLogins) * 100
	avgResponseTime := int64(108) // 模拟平均响应时间

	data := types.OauthStatisticsData{
		TotalLogins:     totalLogins,
		TotalUsers:      3456,
		TotalProviders:  int64(len(providerStats)),
		TodayLogins:     234,
		AvgResponseTime: avgResponseTime,
		SuccessRate:     successRate,
		WeeklyGrowth:    12.5,
		MonthlyGrowth:   8.3,
		ProviderStats:   providerStats,
		LoginTrend:      loginTrend,
	}

	return &types.OauthStatisticsResp{
		BaseDataInfo: types.BaseDataInfo{
			Code: 0,
			Msg:  l.svcCtx.Trans.Trans(l.ctx, i18n.Success),
		},
		Data: data,
	}, nil
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}
