package redisfunc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/coder-lulu/newbee-common/v2/config"
	"github.com/redis/go-redis/v9"
)

// RemoveAllKeyByPrefix removes all key by prefix in redis
func RemoveAllKeyByPrefix(ctx context.Context, prefix string, rds redis.UniversalClient) error {
	var cursor uint64

	for {
		var keys []string
		var err error
		keys, cursor, err = rds.Scan(ctx, cursor, prefix+"*", 0).Result()
		if err != nil {
			return err
		}

		for _, key := range keys {
			rds.Del(ctx, key)
		}

		if cursor == 0 {
			break
		}
	}

	return nil
}

type casbinReloadPayload struct {
	Method    string `json:"Method"`
	Source    string `json:"Source"`
	TenantID  uint64 `json:"TenantID"`
	Timestamp int64  `json:"Timestamp"`
}

// PublishCasbinReload 向Redis频道广播Casbin策略重新加载通知
func PublishCasbinReload(ctx context.Context, rds redis.UniversalClient, db int, tenantID uint64, source string) error {
	if rds == nil {
		return fmt.Errorf("redis client is nil")
	}

	if source == "" {
		source = "tenant_init"
	}

	payload := casbinReloadPayload{
		Method:    "LoadPolicy",
		Source:    source,
		TenantID:  tenantID,
		Timestamp: time.Now().UnixNano(),
	}

	message, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal casbin reload payload: %w", err)
	}

	channel := fmt.Sprintf("%s-%d", config.RedisCasbinChannel, db)
	return rds.Publish(ctx, channel, message).Err()
}
