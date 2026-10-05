package oauthprovider

import (
	"context"
	"fmt"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetOauthStatisticsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOauthStatisticsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOauthStatisticsLogic {
	return &GetOauthStatisticsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}
func (l *GetOauthStatisticsLogic) GetOauthStatistics(req *types.OauthStatisticsReq) (*types.OauthStatisticsResp, error) {
	data := types.OauthStatisticsData{ProviderStats: []types.ProviderStatData{}, LoginTrend: []types.LoginTrendData{}}
	var read uint64
	var success int64
	for page := uint64(1); ; page++ {
		providers, err := l.svcCtx.CoreRpc.GetOauthProviderList(l.ctx, &core.OauthProviderListReq{Page: page, PageSize: 20})
		if err != nil {
			return nil, err
		}
		if providers == nil {
			return nil, fmt.Errorf("OAuth provider list response is missing")
		}
		read += uint64(len(providers.Data))
		for _, provider := range providers.Data {
			if provider == nil {
				return nil, fmt.Errorf("OAuth provider list contains an empty row")
			}
			if req.ProviderId != nil && provider.GetId() != *req.ProviderId {
				continue
			}
			ok, failed := int64(provider.GetSuccessCount()), int64(provider.GetFailureCount())
			total := ok + failed
			rate := float64(0)
			if total > 0 {
				rate = float64(ok) / float64(total) * 100
			}
			name := provider.GetDisplayName()
			if name == "" {
				name = provider.GetName()
			}
			data.ProviderStats = append(data.ProviderStats, types.ProviderStatData{ProviderId: provider.GetId(), ProviderName: provider.GetName(), DisplayName: name, Type: provider.GetType(), IconUrl: provider.IconUrl, TotalUsage: total, SuccessCount: ok, FailureCount: failed, SuccessRate: rate, LastUsed: provider.LastUsedAt})
			data.TotalLogins += total
			success += ok
		}
		if read >= providers.Total {
			break
		}
		if len(providers.Data) == 0 || page >= 1000 {
			return nil, fmt.Errorf("OAuth provider pagination ended before all records were read")
		}
	}
	data.TotalProviders = int64(len(data.ProviderStats))
	if data.TotalLogins > 0 {
		data.SuccessRate = float64(success) / float64(data.TotalLogins) * 100
	}
	// Provider counters have no event timestamps, distinct-user counts or latency.
	// Leave unavailable legacy numeric fields at zero and return no invented trend.
	return &types.OauthStatisticsResp{BaseDataInfo: types.BaseDataInfo{Code: 0, Msg: "success"}, Data: data}, nil
}
