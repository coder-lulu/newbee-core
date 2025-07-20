package redisfunc

import (
	"context"

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
