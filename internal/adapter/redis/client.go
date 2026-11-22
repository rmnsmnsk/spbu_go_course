package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

func Open(ctx context.Context, address string) (*goredis.Client, error) {
	client := goredis.NewClient(&goredis.Options{Addr: address})
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping Redis: %w", err)
	}
	return client, nil
}
