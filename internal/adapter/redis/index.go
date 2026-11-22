package redis

import (
	"context"
	"fmt"
	"strconv"

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

func (i *Index) FindAny(ctx context.Context, words []string) ([]int64, error) {
	if len(words) == 0 {
		return []int64{}, nil
	}

	keys := make([]string, 0, len(words))
	for _, word := range words {
		keys = append(keys, indexKeyPrefix+word)
	}
	members, err := i.client.SUnion(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("find movie IDs: %w", err)
	}

	ids := make([]int64, 0, len(members))
	for _, member := range members {
		id, err := strconv.ParseInt(member, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse movie ID %q: %w", member, err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}
