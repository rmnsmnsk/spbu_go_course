package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

const (
	indexKeyPrefix = "movie:index:"
	indexKeysKey   = "movie:index-keys"
)

type Index struct {
	client *goredis.Client
}

func NewIndex(client *goredis.Client) *Index {
	return &Index{client: client}
}

func (i *Index) Replace(ctx context.Context, entries map[string][]int64) error {
	oldKeys, err := i.client.SMembers(ctx, indexKeysKey).Result()
	if err != nil {
		return fmt.Errorf("get old index keys: %w", err)
	}
	oldKeys = append(oldKeys, indexKeysKey)
	if err := i.client.Del(ctx, oldKeys...).Err(); err != nil {
		return fmt.Errorf("delete old index: %w", err)
	}

	if len(entries) == 0 {
		return nil
	}

	pipe := i.client.TxPipeline()
	keys := make([]any, 0, len(entries))
	for word, ids := range entries {
		key := indexKeyPrefix + word
		members := make([]any, 0, len(ids))
		for _, id := range ids {
			members = append(members, id)
		}
		pipe.SAdd(ctx, key, members...)
		keys = append(keys, key)
	}
	pipe.SAdd(ctx, indexKeysKey, keys...)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("write movie index: %w", err)
	}
	return nil
}
