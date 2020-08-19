package sic

import (
	"context"
	"github.com/go-redis/redis/v8"
	"golang.org/x/sync/singleflight"
)

type cache struct {
	config Config
	redis  redis.UniversalClient
	local  LocalCache

	sf singleflight.Group
}

func NewCache(config Config, redis redis.UniversalClient, local LocalCache) *cache {
	return &cache{
		redis:  redis,
		local:  local,
		config: config,
		sf:     singleflight.Group{},
	}
}

func (c *cache) Get(ctx context.Context, key string, f FetchFunc) (interface{}, error) {
	return c.get(ctx, key, f)
}

func (c *cache) get(ctx context.Context, key string, f FetchFunc) (v interface{}, err error) {
	var found bool

	// get local cache
	v, found = c.local.Get(key)
	if found {
		return v, nil
	}

	// get redis cache
	v, err, _ = c.sf.Do(key, func() (interface{}, error) {
		return c.redis.Get(ctx, key).Bytes()
	})
	if err != nil {
		if err != redis.Nil {
			return nil, err
		}
	} else {
		// TODO: custom metrics of no-hit local cache
		c.local.Set(key, v, c.config.LocalCacheExpire)
		return v, nil
	}

	// get origin data
	v, err, _ = c.sf.Do(key, func() (interface{}, error) {
		// TODO: custom metrics of no-hit any cache
		return f()
	})
	if err != nil {
		return nil, err
	}
	// キャッシュに失敗してもエラーにしないほうがよい
	c.local.Set(key, v, c.config.LocalCacheExpire)
	if err := c.redis.Set(ctx, key, v, c.config.RedisCacheExpire).Err(); err != nil {
		// TODO: logging
	}

	return
}
