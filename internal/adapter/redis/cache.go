package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"movie-service/internal/domain"
)

const (
	searchCacheKeyPrefix = "movie:search-cache:"
	searchCacheKeysKey   = "movie:search-cache-keys"
	searchCacheTTL       = time.Hour
)

type SearchCache struct {
	client *goredis.Client
}

func NewSearchCache(client *goredis.Client) *SearchCache {
	return &SearchCache{client: client}
}

func (c *SearchCache) Get(ctx context.Context, query string) ([]domain.Movie, bool, error) {
	data, err := c.client.Get(ctx, searchCacheKeyPrefix+query).Bytes()
	if err == goredis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("read search cache: %w", err)
	}

	var movies []domain.Movie
	if err := json.Unmarshal(data, &movies); err != nil {
		return nil, false, fmt.Errorf("decode search cache: %w", err)
	}
	if movies == nil {
		movies = []domain.Movie{}
	}
	return movies, true, nil
}

func (c *SearchCache) Set(ctx context.Context, query string, movies []domain.Movie) error {
	data, err := json.Marshal(movies)
	if err != nil {
		return fmt.Errorf("encode search cache: %w", err)
	}

	key := searchCacheKeyPrefix + query
	pipe := c.client.TxPipeline()
	pipe.Set(ctx, key, data, searchCacheTTL)
	pipe.SAdd(ctx, searchCacheKeysKey, key)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("write search cache: %w", err)
	}
	return nil
}

func clearSearchCache(ctx context.Context, client *goredis.Client) error {
	keys, err := client.SMembers(ctx, searchCacheKeysKey).Result()
	if err != nil {
		return err
	}
	keys = append(keys, searchCacheKeysKey)
	return client.Del(ctx, keys...).Err()
}
