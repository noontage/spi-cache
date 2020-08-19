//go:generate go run github.com/golang/mock/mockgen -package mock_redis -destination ./mock/redis/redis.go github.com/go-redis/redis/v8 UniversalClient
//go:generate go run github.com/golang/mock/mockgen -package mock_local_cache -destination ./mock/local_cache/local_cache.go . LocalCache
package sic

import "time"

type FetchFunc func() (interface{}, error)

type Config struct {
	LocalCacheExpire time.Duration
	RedisCacheExpire time.Duration
}

type LocalCache interface {
	Get(key interface{}) (value interface{}, found bool)
	Set(key interface{}, value interface{}, dur time.Duration)
	Delete(key interface{})
}
