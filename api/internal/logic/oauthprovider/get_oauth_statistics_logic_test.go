package oauthprovider

import (
	"context"
	"errors"
	"fmt"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/coreclient"
	"github.com/coder-lulu/newbee-core/rpc/types/core"
	"google.golang.org/grpc"
	"testing"
)

type statsClient struct {
	coreclient.Core
	rows     []*core.OauthProviderInfo
	calls    int
	ctx      context.Context
	failPage uint64
}

func (c *statsClient) GetOauthProviderList(ctx context.Context, r *core.OauthProviderListReq, _ ...grpc.CallOption) (*core.OauthProviderListResp, error) {
	c.calls++
	c.ctx = ctx
	if r.Page == c.failPage {
		return nil, errors.New("RPC failed")
	}
	start := (r.Page - 1) * r.PageSize
	end := start + r.PageSize
	if end > uint64(len(c.rows)) {
		end = uint64(len(c.rows))
	}
	return &core.OauthProviderListResp{Total: uint64(len(c.rows)), Data: c.rows[start:end]}, nil
}
func TestOAuthStatisticsRealPagesAndFilter(t *testing.T) {
	ctx := context.WithValue(context.Background(), struct{}{}, "tenant")
	c := &statsClient{}
	for i := uint64(1); i <= 32; i++ {
		id := i
		success, failure := int32(3), int32(1)
		name := fmt.Sprintf("provider-%d", i)
		c.rows = append(c.rows, &core.OauthProviderInfo{Id: &id, Name: &name, SuccessCount: &success, FailureCount: &failure})
	}
	logic := NewGetOauthStatisticsLogic(ctx, &svc.ServiceContext{CoreRpc: c})
	response, err := logic.GetOauthStatistics(&types.OauthStatisticsReq{})
	if err != nil {
		t.Fatal(err)
	}
	if c.calls != 2 || c.ctx != ctx || response.Data.TotalProviders != 32 || response.Data.TotalLogins != 128 || response.Data.SuccessRate != 75 || len(response.Data.ProviderStats) != 32 {
		t.Fatalf("bad aggregation: %+v calls=%d", response.Data, c.calls)
	}
	if len(response.Data.LoginTrend) != 0 || response.Data.TotalUsers != 0 || response.Data.TodayLogins != 0 || response.Data.AvgResponseTime != 0 {
		t.Fatal("unavailable event statistics invented")
	}
	id := uint64(32)
	response, err = logic.GetOauthStatistics(&types.OauthStatisticsReq{ProviderId: &id})
	if err != nil || response.Data.TotalProviders != 1 || response.Data.TotalLogins != 4 || response.Data.ProviderStats[0].ProviderId != 32 {
		t.Fatalf("filter failed: %+v %v", response, err)
	}
	c.failPage = 2
	if result, err := logic.GetOauthStatistics(&types.OauthStatisticsReq{}); err == nil || result != nil {
		t.Fatal("partial RPC failure became successful statistics")
	}
}
func TestOAuthStatisticsEmpty(t *testing.T) {
	result, err := NewGetOauthStatisticsLogic(context.Background(), &svc.ServiceContext{CoreRpc: &statsClient{}}).GetOauthStatistics(&types.OauthStatisticsReq{})
	if err != nil || result.Data.TotalProviders != 0 || result.Data.SuccessRate != 0 || result.Data.ProviderStats == nil || result.Data.LoginTrend == nil {
		t.Fatalf("wrong empty data: %+v %v", result, err)
	}
}
