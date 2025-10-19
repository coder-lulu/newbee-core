package svc

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/coder-lulu/newbee-common/v2/middleware/framework"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/hooks"
	"github.com/coder-lulu/newbee-core/rpc/coreclient"
	"github.com/coder-lulu/newbee-core/rpc/types/core"
)

// RpcTenantInfoProvider fetches tenant metadata via core RPC service.
type RpcTenantInfoProvider struct {
	client coreclient.Core
}

// NewRpcTenantInfoProvider creates a tenant info provider backed by core RPC.
func NewRpcTenantInfoProvider(client coreclient.Core) framework.TenantInfoProvider {
	return &RpcTenantInfoProvider{client: client}
}

// GetTenantInfo returns normalized tenant metadata for middleware validation.
func (p *RpcTenantInfoProvider) GetTenantInfo(ctx context.Context, tenantID string) (*framework.TenantInfo, error) {
	idStr := strings.TrimSpace(tenantID)
	if idStr == "" {
		return nil, fmt.Errorf("tenant id is empty")
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant id %q: %w", idStr, err)
	}

	rpcCtx := hooks.NewSystemContext(ctx)
	info, err := p.client.GetTenantById(rpcCtx, &core.IDReq{Id: id})
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, fmt.Errorf("tenant %s not found", idStr)
	}

	status := framework.TenantStatusSuspended
	if info.GetStatus() == 1 {
		status = framework.TenantStatusActive
	}

	updatedAt := time.Now()
	if ts := info.GetUpdatedAt(); ts != 0 {
		updatedAt = time.Unix(ts, 0)
	}

	return &framework.TenantInfo{
		ID:        idStr,
		Status:    status,
		UpdatedAt: updatedAt,
	}, nil
}
