package cache

import "github.com/go-redis/redis/v8"

type RateLimitkk struct {
	rdb *redis.Client
}
