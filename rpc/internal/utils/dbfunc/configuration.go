package dbfunc

import (
	"context"

	"github.com/coder-lulu/newbee-common/utils/dynamicconf"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/coder-lulu/newbee-core/rpc/ent"
)

func RefreshDynamicConfiguration(db *ent.Client, rds redis.UniversalClient) {
	allData, err := db.Configuration.Query().All(context.Background())
	if err != nil {
		return
	}

	for _, v := range allData {
		err := dynamicconf.SetDynamicConfigurationToRedis(rds, v.Category, v.Key, v.Value)
		if err != nil {
			logx.Errorw("fail to refresh dynamic configuration", logx.Field("details", err), logx.Field("data", v))
			return
		}
	}
}
