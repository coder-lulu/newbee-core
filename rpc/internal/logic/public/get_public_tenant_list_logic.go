package public

import (
	"context"
	"strconv"

	"github.com/coder-lulu/newbee-core/rpc/ent/tenant"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/coder-lulu/newbee-common/v2/orm/ent/hooks"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPublicTenantListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPublicTenantListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPublicTenantListLogic {
	return &GetPublicTenantListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetPublicTenantListLogic) GetPublicTenantList(in *core.Empty) (*core.PublicTenantListResp, error) {
	// 🔍 添加调试日志
	logx.Infow("🔍 GetPublicTenantList called",
		logx.Field("original_ctx", hooks.DiagnoseTenantContext(l.ctx)))

	// 创建系统上下文绕过租户隔离
	systemCtx := hooks.NewSystemContext(l.ctx)

	// 🔍 验证系统上下文
	logx.Infow("🔍 SystemContext created",
		logx.Field("system_ctx", hooks.DiagnoseTenantContext(systemCtx)))

	// 查询所有激活状态的租户，使用系统上下文绕过租户隔离
	result, err := l.svcCtx.DB.Tenant.Query().
		Where(tenant.StatusEQ(1)). // 1 表示激活状态
		Order(tenant.ByCreatedAt()).
		All(systemCtx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 转换结果，返回与前端 TenantResp 格式匹配的数据
	resp := &core.PublicTenantListResp{
		TenantEnabled: true, // 假设总是启用租户功能
		VoList:        make([]*core.PublicTenantInfo, 0, len(result)),
	}

	for _, t := range result {
		publicTenantInfo := &core.PublicTenantInfo{
			TenantId:    strconv.FormatUint(t.ID, 10), // 将 uint64 转换为 string
			CompanyName: t.Name,
			Domain:      nil, // 暂时设为 nil，可以后续从配置中获取
		}
		resp.VoList = append(resp.VoList, publicTenantInfo)
	}

	return resp, nil
}
